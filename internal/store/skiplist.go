package store

import (
	"math/rand"
)

const (
	maxLevel    = 32   // Maximum level for skip list
	probability = 0.25 // Probability for level promotion
)

type skiplistNode struct {
	member   string
	score    float64
	forward  []*skiplistNode // Array of forward pointers
	backward *skiplistNode   // Backward pointer for reverse iteration
	level    int             // Current node level
}

type skiplist struct {
	head   *skiplistNode // Header node
	tail   *skiplistNode // Tail node
	length int           // Number of nodes in the skip list
	level  int           // Current maximum level of the skip list
}

func newSkiplist() *skiplist {
	header := &skiplistNode{
		forward: make([]*skiplistNode, maxLevel),
		level:   maxLevel,
	}
	return &skiplist{
		head:  header,
		level: 1,
	}
}

func randomLevel() int {
	level := 1
	for level < maxLevel && rand.Float64() < probability {
		level++
	}
	return level
}

func (sl *skiplist) insert(score float64, member string) bool {
	update := make([]*skiplistNode, maxLevel)
	current := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil &&
			(current.forward[i].score < score ||
				(current.forward[i].score == score && current.forward[i].member < member)) {
			current = current.forward[i]
		}
		update[i] = current
	}

	current = current.forward[0]

	if current != nil && current.member == member {
		oldScore := current.score
		current.score = score

		if oldScore == score {
			return false
		}

		sl.delete(oldScore, member)
		return sl.insert(score, member)
	}

	level := randomLevel()
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			update[i] = sl.head
		}
		sl.level = level
	}

	newNode := &skiplistNode{
		member:  member,
		score:   score,
		forward: make([]*skiplistNode, level),
		level:   level,
	}

	for i := 0; i < level; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}

	if update[0] == sl.head {
		newNode.backward = nil
	} else {
		newNode.backward = update[0]
	}

	if newNode.forward[0] != nil {
		newNode.forward[0].backward = newNode
	} else {
		sl.tail = newNode
	}

	sl.length++
	return true
}

func (sl *skiplist) delete(score float64, member string) bool {
	update := make([]*skiplistNode, maxLevel)
	current := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil &&
			(current.forward[i].score < score ||
				(current.forward[i].score == score && current.forward[i].member < member)) {
			current = current.forward[i]
		}
		update[i] = current
	}

	current = current.forward[0]

	if current == nil || current.member != member {
		return false
	}

	for i := 0; i < sl.level; i++ {
		if update[i].forward[i] != current {
			break
		}
		update[i].forward[i] = current.forward[i]
	}

	if current.forward[0] != nil {
		current.forward[0].backward = current.backward
	} else {
		sl.tail = current.backward
	}

	for sl.level > 1 && sl.head.forward[sl.level-1] == nil {
		sl.level--
	}

	sl.length--
	return true
}

// getRange returns a slice of skiplistNodes from start to stop (inclusive).
// If the range exceeds the number of elements in the skiplist, it returns
// as many elements as are available from the start index onward.
func (sl *skiplist) getRange(start, stop int) []*skiplistNode {
	var result []*skiplistNode

	if start < 0 {
		start = sl.length + start
	}
	if stop < 0 {
		stop = sl.length + stop
	}

	if start < 0 {
		start = 0
	}
	if stop >= sl.length {
		stop = sl.length - 1
	}
	if start > stop {
		return result
	}

	current := sl.head.forward[0]
	for i := 0; i < start && current != nil; i++ {
		current = current.forward[0]
	}

	for i := start; i <= stop && current != nil; i++ {
		result = append(result, current)
		current = current.forward[0]
	}

	return result
}
