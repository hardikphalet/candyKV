package store

import (
	"testing"
	"time"

	"github.com/hardikphalet/go-redis/internal/commands/options"
	"github.com/hardikphalet/go-redis/internal/types"
)

func TestMemoryStore_Get(t *testing.T) {
	store := NewMemoryStore()

	tests := []struct {
		name     string
		setup    func()
		key      string
		want     interface{}
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "get existing string",
			setup: func() {
				store.Set("key1", "value1", nil)
			},
			key:  "key1",
			want: "value1",
		},
		{
			name: "get non-existing key",
			key:  "nonexistent",
			want: nil,
		},
		{
			name: "get expired key",
			setup: func() {
				store.Set("expired", "value", nil)
				store.Expire("expired", 1*time.Millisecond, nil)
				time.Sleep(2 * time.Millisecond)
			},
			key:  "expired",
			want: nil,
		},
		{
			name: "get wrong type",
			setup: func() {
				members := []types.ScoreMember{{Score: 1.0, Member: "member1"}}
				store.ZAdd("wrongtype", members, nil)
			},
			key:     "wrongtype",
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "WRONGTYPE Operation against a key holding a sorted set"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			got, err := store.Get(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStore.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("MemoryStore.Get() error = %v, did not match error check", err)
				}
				return
			}

			if got != tt.want {
				t.Errorf("MemoryStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryStore_Set(t *testing.T) {
	store := NewMemoryStore()

	tests := []struct {
		name     string
		key      string
		value    interface{}
		opts     *options.SetOptions
		want     interface{}
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name:  "simple set",
			key:   "key1",
			value: "value1",
			opts:  nil,
			want:  nil,
		},
		{
			name:  "set with NX on non-existing key",
			key:   "key2",
			value: "value2",
			opts:  func() *options.SetOptions { o := options.NewSetOptions(); o.Set("NX"); return o }(),
			want:  nil,
		},
		{
			name:    "set with NX on existing key",
			key:     "key1",
			value:   "newvalue",
			opts:    func() *options.SetOptions { o := options.NewSetOptions(); o.Set("NX"); return o }(),
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "key already exists"
			},
		},
		{
			name:  "set with XX on existing key",
			key:   "key1",
			value: "newvalue",
			opts:  func() *options.SetOptions { o := options.NewSetOptions(); o.Set("XX"); return o }(),
			want:  nil,
		},
		{
			name:    "set with XX on non-existing key",
			key:     "nonexistent",
			value:   "value",
			opts:    func() *options.SetOptions { o := options.NewSetOptions(); o.Set("XX"); return o }(),
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "key does not exist"
			},
		},
		{
			name:  "set with GET",
			key:   "key1",
			value: "newestvalue",
			opts:  func() *options.SetOptions { o := options.NewSetOptions(); o.Set("GET"); return o }(),
			want:  "newvalue", // returns the old value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Set(tt.key, tt.value, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStore.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("MemoryStore.Set() error = %v, did not match error check", err)
				}
				return
			}

			if got != tt.want {
				t.Errorf("MemoryStore.Set() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryStore_Del(t *testing.T) {
	store := NewMemoryStore()

	// Setup
	store.Set("key1", "value1", nil)
	store.Set("key2", "value2", nil)

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name: "delete existing key",
			key:  "key1",
		},
		{
			name: "delete non-existing key",
			key:  "nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Del(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStore.Del() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify key is deleted
			val, _ := store.Get(tt.key)
			if val != nil {
				t.Errorf("MemoryStore.Del() key still exists after deletion")
			}
		})
	}
}

func TestMemoryStore_Expire(t *testing.T) {
	store := NewMemoryStore()

	tests := []struct {
		name     string
		setup    func()
		key      string
		ttl      time.Duration
		opts     *options.ExpireOptions
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "expire existing key",
			setup: func() {
				store.Set("key1", "value1", nil)
			},
			key: "key1",
			ttl: 5 * time.Second,
		},
		{
			name:    "expire non-existing key",
			key:     "nonexistent",
			ttl:     time.Second,
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "key does not exist"
			},
		},
		{
			name: "expire with negative TTL",
			setup: func() {
				store.Set("key2", "value2", nil)
			},
			key: "key2",
			ttl: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := store.Expire(tt.key, tt.ttl, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStore.Expire() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("MemoryStore.Expire() error = %v, did not match error check", err)
				}
				return
			}

			if tt.ttl > 0 {
				// Verify TTL is set
				ttl, _ := store.TTL(tt.key)
				if ttl <= 0 {
					t.Errorf("MemoryStore.Expire() TTL not set correctly")
				}
			} else {
				// Verify key is deleted for negative TTL
				val, _ := store.Get(tt.key)
				if val != nil {
					t.Errorf("MemoryStore.Expire() key not deleted with negative TTL")
				}
			}
		})
	}
}

func TestMemoryStore_TTL(t *testing.T) {
	store := NewMemoryStore()

	tests := []struct {
		name    string
		setup   func()
		key     string
		want    int
		wantErr bool
	}{
		{
			name: "get TTL of key with expiry",
			setup: func() {
				store.Set("key1", "value1", nil)
				store.Expire("key1", time.Second*5, nil)
			},
			key:  "key1",
			want: 1,
		},
		{
			name: "get TTL of key without expiry",
			setup: func() {
				store.Set("key2", "value2", nil)
			},
			key:  "key2",
			want: -1,
		},
		{
			name: "get TTL of non-existing key",
			key:  "nonexistent",
			want: -2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			got, err := store.TTL(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStore.TTL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want >= 0 {
				if got <= 0 {
					t.Errorf("MemoryStore.TTL() = %v, want positive value", got)
				}
			} else {
				if got != tt.want {
					t.Errorf("MemoryStore.TTL() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMemoryStore_Keys(t *testing.T) {
	store := NewMemoryStore()

	// Setup
	store.Set("key1", "value1", nil)
	store.Set("key2", "value2", nil)
	store.Set("test1", "value3", nil)
	store.Set("test2", "value4", nil)

	tests := []struct {
		name    string
		pattern string
		want    int
		wantErr bool
	}{
		{
			name:    "match all keys",
			pattern: "*",
			want:    4,
		},
		{
			name:    "match keys with prefix",
			pattern: "key*",
			want:    2,
		},
		{
			name:    "match keys with suffix",
			pattern: "*1",
			want:    2,
		},
		{
			name:    "match exact key",
			pattern: "key1",
			want:    1,
		},
		{
			name:    "match with question mark",
			pattern: "test?",
			want:    2,
		},
		{
			name:    "no matches",
			pattern: "nomatch*",
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Keys(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStore.Keys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.want {
				t.Errorf("MemoryStore.Keys() = %v, want %v matches", len(got), tt.want)
			}
		})
	}
}

func TestMemoryStore_ZAdd(t *testing.T) {
	store := NewMemoryStore()

	tests := []struct {
		name     string
		key      string
		members  []types.ScoreMember
		opts     *options.ZAddOptions
		want     interface{}
		wantErr  bool
		errCheck func(error) bool
		setup    func()
	}{
		{
			name: "add new members",
			key:  "zset1",
			members: []types.ScoreMember{
				{Score: 1.0, Member: "one"},
				{Score: 2.0, Member: "two"},
			},
			want: 2,
		},
		{
			name: "update existing member",
			key:  "zset1",
			members: []types.ScoreMember{
				{Score: 3.0, Member: "one"},
			},
			want: 0,
		},
		{
			name: "add with NX option",
			key:  "zset1",
			members: []types.ScoreMember{
				{Score: 4.0, Member: "one"},
			},
			opts: func() *options.ZAddOptions { o := options.NewZAddOptions(); o.Set("NX"); return o }(),
			want: 0,
		},
		{
			name: "add with XX option",
			key:  "zset1",
			members: []types.ScoreMember{
				{Score: 5.0, Member: "one"},
			},
			opts: func() *options.ZAddOptions { o := options.NewZAddOptions(); o.Set("XX"); return o }(),
			want: 0,
		},
		{
			name: "add to wrong type",
			setup: func() {
				store.Set("string", "value", nil)
			},
			key: "string",
			members: []types.ScoreMember{
				{Score: 1.0, Member: "one"},
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "WRONGTYPE Operation against a key holding the wrong kind of value"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			got, err := store.ZAdd(tt.key, tt.members, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStore.ZAdd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("MemoryStore.ZAdd() error = %v, did not match error check", err)
				}
				return
			}

			if got != tt.want {
				t.Errorf("MemoryStore.ZAdd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryStore_ZRange(t *testing.T) {
	store := NewMemoryStore()

	// Setup
	members := []types.ScoreMember{
		{Score: 1.0, Member: "one"},
		{Score: 2.0, Member: "two"},
		{Score: 3.0, Member: "three"},
	}
	store.ZAdd("zset1", members, nil)

	tests := []struct {
		name     string
		key      string
		start    interface{}
		stop     interface{}
		opts     *options.ZRangeOptions
		want     []interface{}
		wantErr  bool
		errCheck func(error) bool
		setup    func()
	}{
		{
			name:  "range by index",
			key:   "zset1",
			start: 0,
			stop:  -1,
			want:  []interface{}{"one", "two", "three"},
		},
		{
			name:  "range with negative indices",
			key:   "zset1",
			start: -2,
			stop:  -1,
			want:  []interface{}{"two", "three"},
		},
		{
			name:  "range with mixed indices",
			key:   "zset1",
			start: 1,
			stop:  -1,
			want:  []interface{}{"two", "three"},
		},
		{
			name:  "range with out of bounds indices",
			key:   "zset1",
			start: -5,
			stop:  5,
			want:  []interface{}{"one", "two", "three"},
		},
		{
			name:  "range with invalid range",
			key:   "zset1",
			start: 2,
			stop:  1,
			want:  []interface{}{},
		},
		{
			name:  "range by score",
			key:   "zset1",
			start: 1.0,
			stop:  2.0,
			opts: func() *options.ZRangeOptions {
				o := options.NewZRangeOptions()
				o.SetRangeType("BYSCORE")
				return o
			}(),
			want: []interface{}{"one", "two"},
		},
		{
			name:  "range by lex",
			key:   "zset1",
			start: "one",
			stop:  "three",
			opts: func() *options.ZRangeOptions {
				o := options.NewZRangeOptions()
				o.SetRangeType("BYLEX")
				return o
			}(),
			want: []interface{}{"one", "three"},
		},
		{
			name:  "range with scores",
			key:   "zset1",
			start: 0,
			stop:  -1,
			opts: func() *options.ZRangeOptions {
				o := options.NewZRangeOptions()
				o.WithScores = true
				return o
			}(),
			want: []interface{}{"one", 1.0, "two", 2.0, "three", 3.0},
		},
		{
			name:  "range from wrong type",
			key:   "string",
			start: 0,
			stop:  -1,
			setup: func() {
				store.Set("string", "value", nil)
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "WRONGTYPE Operation against a key holding the wrong kind of value"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			got, err := store.ZRange(tt.key, tt.start, tt.stop, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStore.ZRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("MemoryStore.ZRange() error = %v, did not match error check", err)
				}
				return
			}

			if err != nil {
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("MemoryStore.ZRange() returned %v elements, want %v", len(got), len(tt.want))
				return
			}

			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("MemoryStore.ZRange() result[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}
