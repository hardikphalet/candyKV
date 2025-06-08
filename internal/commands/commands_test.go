package commands

import (
	"testing"
	"time"

	"github.com/hardikphalet/go-redis/internal/commands/options"
	"github.com/hardikphalet/go-redis/internal/store"
	"github.com/hardikphalet/go-redis/internal/types"
)

func TestPingCommand_Execute(t *testing.T) {
	cmd := &PingCommand{}
	s := store.NewMemoryStore()

	result, err := cmd.Execute(s)
	if err != nil {
		t.Errorf("PingCommand.Execute() error = %v", err)
	}

	if result != types.SimpleString("PONG") {
		t.Errorf("PingCommand.Execute() = %v, want PONG", result)
	}
}

func TestSetCommand_Execute(t *testing.T) {
	s := store.NewMemoryStore()

	tests := []struct {
		name     string
		cmd      *SetCommand
		want     interface{}
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "simple set",
			cmd: &SetCommand{
				Key:     "key1",
				Value:   "value1",
				Options: nil,
			},
			want: nil,
		},
		{
			name: "set with NX on non-existing key",
			cmd: &SetCommand{
				Key:     "key2",
				Value:   "value2",
				Options: func() *options.SetOptions { o := options.NewSetOptions(); o.Set("NX"); return o }(),
			},
			want: nil,
		},
		{
			name: "set with GET",
			cmd: &SetCommand{
				Key:     "key1",
				Value:   "newvalue",
				Options: func() *options.SetOptions { o := options.NewSetOptions(); o.Set("GET"); return o }(),
			},
			want: "value1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cmd.Execute(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("SetCommand.Execute() error = %v, did not match error check", err)
				}
				return
			}

			if got != tt.want {
				t.Errorf("SetCommand.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCommand_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	s.Set("key1", "value1", nil)

	tests := []struct {
		name     string
		cmd      *GetCommand
		want     interface{}
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "get existing key",
			cmd:  &GetCommand{Key: "key1"},
			want: "value1",
		},
		{
			name: "get non-existing key",
			cmd:  &GetCommand{Key: "nonexistent"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cmd.Execute(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("GetCommand.Execute() error = %v, did not match error check", err)
				}
				return
			}

			if got != tt.want {
				t.Errorf("GetCommand.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDelCommand_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	s.Set("key1", "value1", nil)
	s.Set("key2", "value2", nil)
	s.Set("key3", "value3", nil)

	tests := []struct {
		name     string
		cmd      *DelCommand
		want     interface{}
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "delete single existing key",
			cmd:  &DelCommand{Keys: []string{"key1"}},
			want: 1,
		},
		{
			name: "delete multiple keys",
			cmd:  &DelCommand{Keys: []string{"key2", "key3"}},
			want: 2,
		},
		{
			name: "delete non-existing key",
			cmd:  &DelCommand{Keys: []string{"nonexistent"}},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cmd.Execute(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("DelCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("DelCommand.Execute() error = %v, did not match error check", err)
				}
				return
			}

			if got != tt.want {
				t.Errorf("DelCommand.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpireCommand_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	s.Set("key1", "value1", nil)

	tests := []struct {
		name     string
		cmd      *ExpireCommand
		want     interface{}
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "expire existing key",
			cmd: &ExpireCommand{
				Key: "key1",
				TTL: time.Second,
			},
			want: nil,
		},
		{
			name: "expire non-existing key",
			cmd: &ExpireCommand{
				Key: "nonexistent",
				TTL: time.Second,
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "key does not exist"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cmd.Execute(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpireCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("ExpireCommand.Execute() error = %v, did not match error check", err)
				}
				return
			}

			if got != tt.want {
				t.Errorf("ExpireCommand.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTtlCommand_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	s.Set("key1", "value1", nil)
	s.Expire("key1", 3*time.Second, nil)
	s.Set("key2", "value2", nil)

	tests := []struct {
		name     string
		cmd      *TtlCommand
		want     int
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "get TTL of key with expiry",
			cmd:  &TtlCommand{Key: "key1"},
			want: 1,
		},
		{
			name: "get TTL of key without expiry",
			cmd:  &TtlCommand{Key: "key2"},
			want: -1,
		},
		{
			name: "get TTL of non-existing key",
			cmd:  &TtlCommand{Key: "nonexistent"},
			want: -2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cmd.Execute(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("TtlCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("TtlCommand.Execute() error = %v, did not match error check", err)
				}
				return
			}

			if tt.want >= 0 {
				if got.(int) < 0 {
					t.Errorf("TtlCommand.Execute() = %v, want positive value", got)
				}
			} else {
				if got != tt.want {
					t.Errorf("TtlCommand.Execute() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestZAddCommand_Execute(t *testing.T) {
	s := store.NewMemoryStore()

	tests := []struct {
		name     string
		cmd      *ZAddCommand
		want     interface{}
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "add new members",
			cmd: &ZAddCommand{
				Key: "zset1",
				Members: []types.ScoreMember{
					{Score: 1.0, Member: "one"},
					{Score: 2.0, Member: "two"},
				},
			},
			want: 2,
		},
		{
			name: "update existing member",
			cmd: &ZAddCommand{
				Key: "zset1",
				Members: []types.ScoreMember{
					{Score: 3.0, Member: "one"},
				},
			},
			want: 0,
		},
		{
			name: "add with NX option",
			cmd: &ZAddCommand{
				Key: "zset1",
				Members: []types.ScoreMember{
					{Score: 4.0, Member: "one"},
				},
				Options: func() *options.ZAddOptions { o := options.NewZAddOptions(); o.Set("NX"); return o }(),
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cmd.Execute(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ZAddCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("ZAddCommand.Execute() error = %v, did not match error check", err)
				}
				return
			}

			if got != tt.want {
				t.Errorf("ZAddCommand.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZRangeCommand_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	members := []types.ScoreMember{
		{Score: 1.0, Member: "one"},
		{Score: 2.0, Member: "two"},
		{Score: 3.0, Member: "three"},
	}
	s.ZAdd("zset1", members, nil)

	tests := []struct {
		name     string
		cmd      *ZRangeCommand
		want     []interface{}
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "range by index",
			cmd: &ZRangeCommand{
				Key:     "zset1",
				Start:   0,
				Stop:    -1,
				Options: options.NewZRangeOptions(),
			},
			want: []interface{}{"one", "two", "three"},
		},
		{
			name: "range by score",
			cmd: &ZRangeCommand{
				Key:   "zset1",
				Start: 1.0,
				Stop:  2.0,
				Options: func() *options.ZRangeOptions {
					o := options.NewZRangeOptions()
					o.SetRangeType("BYSCORE")
					return o
				}(),
			},
			want: []interface{}{"one", "two"},
		},
		{
			name: "range with scores",
			cmd: &ZRangeCommand{
				Key:   "zset1",
				Start: 0,
				Stop:  -1,
				Options: func() *options.ZRangeOptions {
					o := options.NewZRangeOptions()
					o.Set("WITHSCORES")
					return o
				}(),
			},
			want: []interface{}{"one", 1.0, "two", 2.0, "three", 3.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cmd.Execute(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ZRangeCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.errCheck != nil && err != nil {
				if !tt.errCheck(err) {
					t.Errorf("ZRangeCommand.Execute() error = %v, did not match error check", err)
				}
				return
			}

			result := got.([]interface{})
			if len(result) != len(tt.want) {
				t.Logf("result: %v", result)
				t.Errorf("ZRangeCommand.Execute() returned %v elements, want %v", len(result), len(tt.want))
				return
			}

			for i, v := range result {
				if v != tt.want[i] {
					t.Errorf("ZRangeCommand.Execute() result[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}
