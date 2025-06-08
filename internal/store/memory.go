package store

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/hardikphalet/go-redis/internal/commands/options"
	"github.com/hardikphalet/go-redis/internal/types"
)

type SortedSetMember struct {
	Score  float64
	Member string
}

type SortedSet struct {
	dict    map[string]float64 // For O(1) member lookups
	sl      *skiplist          // For ordered operations
	scores  []float64
	members []string
}

func (s *SortedSet) Add(member string, score float64) {
	s.dict[member] = score

	s.sl.insert(score, member)

	s.scores = append(s.scores, score)
	s.members = append(s.members, member)
}

func (s *SortedSet) Range(start, stop int, withScores bool) []interface{} {
	if s == nil || s.sl == nil || len(s.dict) == 0 {
		return []interface{}{}
	}

	nodes := s.sl.getRange(start, stop)

	capacity := len(nodes)
	if withScores {
		capacity *= 2
	}
	result := make([]interface{}, 0, capacity)

	for _, node := range nodes {
		result = append(result, node.member)
		if withScores {
			result = append(result, float64(node.score))
		}
	}

	return result
}

func (s *SortedSet) RangeByScore(min, max float64, rev bool, withScores bool) []interface{} {
	if s == nil || s.sl == nil || len(s.dict) == 0 {
		return []interface{}{}
	}

	var result []interface{}
	current := s.sl.head

	if rev {
		for i := s.sl.level - 1; i >= 0; i-- {
			for current.forward[i] != nil && current.forward[i].score > max {
				current = current.forward[i]
			}
		}
		current = current.forward[0]

		for current != nil && current.score >= min {
			if withScores {
				result = append(result, current.member, current.score)
			} else {
				result = append(result, current.member)
			}
			current = current.forward[0]
		}
	} else {
		for i := s.sl.level - 1; i >= 0; i-- {
			for current.forward[i] != nil && current.forward[i].score < min {
				current = current.forward[i]
			}
		}
		current = current.forward[0]

		for current != nil && current.score <= max {
			if withScores {
				result = append(result, current.member, current.score)
			} else {
				result = append(result, current.member)
			}
			current = current.forward[0]
		}
	}

	return result
}

func (s *SortedSet) RangeByLex(min, max string, rev bool) []interface{} {
	var result []interface{}

	if rev {
		// Reverse order
		for i := len(s.members) - 1; i >= 0; i-- {
			member := s.members[i]
			if member >= min && member <= max {
				result = append(result, member)
			}
		}
	} else {
		// Forward order
		for i := 0; i < len(s.members); i++ {
			member := s.members[i]
			if member >= min && member <= max {
				result = append(result, member)
			}
		}
	}

	return result
}

type MemoryStore struct {
	data    map[string]interface{}
	expires map[string]time.Time
	mu      sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data:    make(map[string]interface{}),
		expires: make(map[string]time.Time),
	}
}

// Between reading for expiry and reading from the map, there is a race condition
// func (s *MemoryStore) Get(key string) (interface{}, error) {
// 	s.mu.RLock()
// 	expired := s.isExpired(key)
// 	s.mu.RUnlock()

// 	if expired {
// 		s.Del(key)
// 		return nil, nil
// 	}

//		s.mu.RLock()
//		defer s.mu.RUnlock()
//		if val, ok := s.data[key]; ok {
//			return val, nil
//		}
//		return nil, nil
//	}
func (s *MemoryStore) Get(key string) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isExpired(key) {
		delete(s.data, key)
		delete(s.expires, key)
		return nil, nil
	}

	if val, ok := s.data[key]; ok {
		switch v := val.(type) {
		case string:
			return v, nil
		case *SortedSet:
			return nil, fmt.Errorf("WRONGTYPE Operation against a key holding a sorted set")
		default:
			return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
	}

	return nil, nil
}

func (s *MemoryStore) Set(key string, value interface{}, opts *options.SetOptions) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exists := false
	var oldValue interface{}
	if val, ok := s.data[key]; ok {
		exists = true
		oldValue = val
	}

	if opts != nil && opts.IsNX() && exists {
		return nil, fmt.Errorf("key already exists")
	}

	if opts != nil && opts.IsXX() && !exists {
		return nil, fmt.Errorf("key does not exist")
	}

	s.data[key] = value

	if opts != nil {
		if opts.IsKEEPTTL() {
			if _, ok := s.expires[key]; !ok {
				delete(s.expires, key)
			}
		} else if opts.ExpiryType != "" {
			s.expires[key] = opts.ExpiryTime
		} else {
			delete(s.expires, key)
		}
	} else {
		delete(s.expires, key)
	}

	if opts != nil && opts.IsGET() {
		return oldValue, nil
	}

	return nil, nil
}

func (s *MemoryStore) Del(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	delete(s.expires, key)
	return nil
}

func (s *MemoryStore) Expire(key string, ttl time.Duration, opts *options.ExpireOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; !exists {
		return fmt.Errorf("key does not exist")
	}

	if opts != nil {
		hasExpiry := false
		if expiry, ok := s.expires[key]; ok {
			hasExpiry = !time.Now().After(expiry)
		}

		if opts.IsNX() && hasExpiry {
			return fmt.Errorf("key already has an expiry")
		}

		if opts.IsXX() && !hasExpiry {
			return fmt.Errorf("key has no expiry")
		}

		if opts.IsGT() && hasExpiry {
			currentTTL := time.Until(s.expires[key])
			if ttl <= currentTTL {
				return fmt.Errorf("new expiry is not greater than current one")
			}
		}

		if opts.IsLT() && hasExpiry {
			currentTTL := time.Until(s.expires[key])
			if ttl >= currentTTL {
				return fmt.Errorf("new expiry is not less than current one")
			}
		}
	}

	if ttl <= 0 {
		delete(s.expires, key)
		delete(s.data, key)
		return nil
	}

	s.expires[key] = time.Now().Add(ttl)
	return nil
}

func (s *MemoryStore) TTL(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.data[key]; !exists {
		return -2, nil // Key does not exist
	}

	if expiry, ok := s.expires[key]; ok {
		remaining := time.Until(expiry)
		if remaining <= 0 {
			delete(s.data, key)
			delete(s.expires, key)
			return -2, nil
		}
		return int(remaining.Seconds()), nil
	}

	return -1, nil
}

func (s *MemoryStore) Keys(pattern string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		if !s.isExpired(k) && matchPattern(k, pattern) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (s *MemoryStore) isExpired(key string) bool {
	if expiry, ok := s.expires[key]; ok {
		return time.Now().After(expiry)
	}
	return false
}

// matchPattern implements Redis-style pattern matching
// Supports:
// * - matches zero or more characters
// ? - matches exactly one character
// [...] - matches any character within the brackets
// [^...] - matches any character not within the brackets
func matchPattern(str, pattern string) bool {
	if pattern == "*" {
		return true
	}

	regexPattern := ""
	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '*':
			regexPattern += ".*"
		case '?':
			regexPattern += "."
		case '[':
			j := i + 1
			if j < len(pattern) && pattern[j] == '^' {
				regexPattern += "[^"
				j++
			} else {
				regexPattern += "["
			}
			for j < len(pattern) && pattern[j] != ']' {
				if pattern[j] == '\\' && j+1 < len(pattern) {
					regexPattern += "\\" + string(pattern[j+1])
					j += 2
				} else {
					regexPattern += string(pattern[j])
					j++
				}
			}
			if j < len(pattern) && pattern[j] == ']' {
				regexPattern += "]"
				i = j
			} else {
				regexPattern += "\\["
			}
		case '\\':
			if i+1 < len(pattern) {
				regexPattern += "\\" + string(pattern[i+1])
				i++
			} else {
				regexPattern += "\\\\"
			}
		default:
			if c := string(pattern[i]); c == "." || c == "+" || c == "(" || c == ")" || c == "|" || c == "{" || c == "}" || c == "$" || c == "^" {
				regexPattern += "\\" + c
			} else {
				regexPattern += c
			}
		}
		i++
	}

	matched, err := regexp.MatchString("^"+regexPattern+"$", str)
	if err != nil {
		return false
	}
	return matched
}

func (s *MemoryStore) ZAdd(key string, members []types.ScoreMember, opts *options.ZAddOptions) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get or create sorted set
	var set *SortedSet
	if val, ok := s.data[key]; ok {
		if ss, ok := val.(*SortedSet); ok {
			set = ss
		} else {
			return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
	} else {
		set = &SortedSet{
			dict: make(map[string]float64),
			sl:   newSkiplist(),
		}
		s.data[key] = set
	}

	if opts != nil && opts.IsINCR() {
		if len(members) != 1 {
			return nil, fmt.Errorf("INCR option requires exactly one score-member pair")
		}
		member := members[0]
		oldScore, exists := set.dict[member.Member]

		if opts.IsNX() && exists {
			return nil, nil
		}
		if opts.IsXX() && !exists {
			return nil, nil
		}

		newScore := oldScore + member.Score
		set.Add(member.Member, newScore)
		return newScore, nil
	}

	changed := 0
	added := 0
	for _, member := range members {
		oldScore, exists := set.dict[member.Member]

		if opts != nil && opts.IsNX() && exists {
			continue
		}

		if opts != nil && opts.IsXX() && !exists {
			continue
		}

		if opts != nil && opts.IsGT() && exists && member.Score <= oldScore {
			continue
		}

		if opts != nil && opts.IsLT() && exists && member.Score >= oldScore {
			continue
		}

		if !exists {
			added++
		}
		if !exists || oldScore != member.Score {
			changed++
		}
		set.Add(member.Member, member.Score)
	}

	if opts != nil && opts.IsCH() {
		return changed, nil
	}

	return added, nil
}

func (s *MemoryStore) ZRange(key string, start, stop interface{}, opts *options.ZRangeOptions) ([]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if val, exists := s.data[key]; exists {
		if zset, ok := val.(*SortedSet); ok {
			var result []interface{}
			var withScores bool
			if opts != nil {
				withScores = opts.IsWithScores()
			}

			if opts != nil && opts.IsByScore() {
				minScore, ok := start.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid score range start")
				}
				maxScore, ok := stop.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid score range stop")
				}
				result = zset.RangeByScore(minScore, maxScore, opts.IsRev(), withScores)
			} else if opts != nil && opts.IsByLex() {
				minLex, ok := start.(string)
				if !ok {
					return nil, fmt.Errorf("invalid lex range start")
				}
				maxLex, ok := stop.(string)
				if !ok {
					return nil, fmt.Errorf("invalid lex range stop")
				}
				result = zset.RangeByLex(minLex, maxLex, opts.IsRev())
			} else {
				startIdx, ok := start.(int)
				if !ok {
					return nil, fmt.Errorf("invalid index range start")
				}
				stopIdx, ok := stop.(int)
				if !ok {
					return nil, fmt.Errorf("invalid index range stop")
				}

				setLen := len(zset.dict)

				if startIdx < 0 {
					startIdx = setLen + startIdx
				}
				if stopIdx < 0 {
					stopIdx = setLen + stopIdx
				}

				if startIdx < 0 {
					startIdx = 0
				}
				if stopIdx >= setLen {
					stopIdx = setLen - 1
				}
				if startIdx > stopIdx || startIdx >= setLen {
					return []interface{}{}, nil
				}

				result = zset.Range(startIdx, stopIdx, withScores)
			}

			if opts != nil && opts.Limit.Count > 0 {
				offset := opts.Limit.Offset
				count := opts.Limit.Count

				multiplier := 1
				if withScores {
					multiplier = 2
				}

				// Adjust offset and count based on multiplier
				adjustedOffset := offset * multiplier
				adjustedCount := count * multiplier

				if adjustedOffset >= len(result) {
					return []interface{}{}, nil
				}

				end := adjustedOffset + adjustedCount
				if end > len(result) {
					end = len(result)
				}

				result = result[adjustedOffset:end]
			}

			return result, nil
		}
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return []interface{}{}, nil
}
