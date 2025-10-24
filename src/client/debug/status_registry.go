package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"aegis/client"
	"aegis/utils"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type EntryType string

const (
	EntryTypeReadOnly  EntryType = "readonly"
	EntryTypeReadWrite EntryType = "readwrite"

	HistoryKey string = "rcabench:debug:history"

	DefaultHistoryLimit int = 100
)

type DebugEntry struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Category    string              `json:"category"`
	Type        EntryType           `json:"type"` // "readonly", "readwrite", "action", "health_check"
	GetFunc     func() (any, error) `json:"-"`
	SetFunc     func(any) error     `json:"-"`
	AutoFix     bool                `json:"auto_fix"` // Whether auto-fix is supported
}

// HistoryEntry operation history
type HistoryEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	OldValue  any       `json:"old_value,omitempty"`
	NewValue  any       `json:"new_value,omitempty"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

type DebugRegistry struct {
	mu      sync.RWMutex
	entries map[string]*DebugEntry

	ctx    context.Context
	cancel context.CancelFunc

	// State variable
	debugMode int32 // atomic operation
}

func NewDebugRegistry() *DebugRegistry {
	ctx, cancel := context.WithCancel(context.Background())

	registry := &DebugRegistry{
		entries: make(map[string]*DebugEntry),
		ctx:     ctx,
		cancel:  cancel,
	}
	registry.registerEntries()

	return registry
}

func (r *DebugRegistry) Get(name string) (map[string]any, error) {
	r.mu.RLock()
	entry, exists := r.entries[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("entry %s not found", name)
	}

	entryData := utils.StructToMap(entry)
	if entry.GetFunc != nil {
		value, err := entry.GetFunc()
		if err != nil {
			entryData["value"] = fmt.Sprintf("Error: %v", err)
			entryData["error"] = true
		} else {
			entryData["value"] = value
			entryData["error"] = false
		}
	}

	return entryData, nil
}

func (r *DebugRegistry) GetAll() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]any)
	for name, entry := range r.entries {
		entryData := utils.StructToMap(entry)
		if entry.GetFunc != nil {
			if value, err := entry.GetFunc(); err != nil {
				entryData["value"] = fmt.Sprintf("Error: %v", err)
				entryData["error"] = true
			} else {
				entryData["value"] = value
				entryData["error"] = false
			}
		}

		result[name] = entryData
	}

	return result
}

func (r *DebugRegistry) GetHistory(limit int) ([]HistoryEntry, error) {
	if limit <= 0 {
		limit = DefaultHistoryLimit
	}

	streamResult, err := client.GetRedisClient().XRead(r.ctx, &redis.XReadArgs{
		Streams: []string{HistoryKey, "0"},
		Count:   int64(limit),
		Block:   -1,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read history from redis: %v", err)
	}

	errorTemplate := "invalid or missing '%s' in task payload"

	var history []HistoryEntry
	for _, result := range streamResult {
		for _, message := range result.Messages {
			entry, err := utils.MapToStruct[HistoryEntry](message.Values, "", errorTemplate)
			if err != nil {
				return nil, fmt.Errorf("failed to parse history entry: %v", err)
			}

			history = append(history, *entry)
		}
	}

	return history, nil
}

func (r *DebugRegistry) Register(entry *DebugEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[entry.Name] = entry
}

func (r *DebugRegistry) Set(name string, value any) error {
	r.mu.RLock()
	entry, exists := r.entries[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("entry %s not found", name)
	}

	if entry.Type == EntryTypeReadOnly {
		return fmt.Errorf("entry %s is readonly", name)
	}

	if entry.SetFunc == nil {
		return fmt.Errorf("set function not implemented for %s", name)
	}

	var oldValue any
	if entry.GetFunc != nil {
		oldValue, _ = entry.GetFunc()
	}

	err := entry.SetFunc(value)
	r.addHistory(HistoryEntry{
		ID:        fmt.Sprintf("%s_%d", name, time.Now().UnixNano()),
		Timestamp: time.Now(),
		Action:    "set",
		Target:    name,
		OldValue:  oldValue,
		NewValue:  value,
		Success:   err == nil,
		Error: func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}(),
	})

	return err
}

func (r *DebugRegistry) addHistory(entry HistoryEntry) {
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return
	}

	_, err = client.GetRedisClient().XAdd(r.ctx, &redis.XAddArgs{
		Stream: HistoryKey,
		MaxLen: 10000,
		Approx: true,
		ID:     "*",
		Values: entryJSON,
	}).Result()
	if err != nil {
		logrus.Errorf("failed to add event to Redis stream %s: %v", HistoryKey, err)
	}
}

func (r *DebugRegistry) registerEntries() {
	r.Register(&DebugEntry{
		Name:        "debug_mode",
		Description: "Debug mode status",
		Category:    "system",
		Type:        EntryTypeReadWrite,
		GetFunc: func() (any, error) {
			return atomic.LoadInt32(&r.debugMode) == 1, nil
		},
		SetFunc: func(value any) error {
			var newValue int32
			switch v := value.(type) {
			case bool:
				if v {
					newValue = 1
				}
			case string:
				if v == "true" || v == "1" {
					newValue = 1
				}
			default:
				return fmt.Errorf("invalid value type: %T", value)
			}

			atomic.StoreInt32(&r.debugMode, newValue)
			return nil
		},
	})
}
