package memory

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
)

func TestMemoryInstanceStorage_GetInstanceByID(t *testing.T) {
	tests := []struct {
		name          string
		presetData    map[int64]domain.Instance
		requestID     int64
		expectedError error
		expectedName  string
	}{
		{
			name: "found existing instance",
			presetData: map[int64]domain.Instance{
				1: {ID: 1, Name: "Lobby-1"},
			},
			requestID:     1,
			expectedError: nil,
			expectedName:  "Lobby-1",
		},
		{
			name:          "not found empty storage",
			presetData:    map[int64]domain.Instance{},
			requestID:     99,
			expectedError: domain.ErrNotFound,
		},
		{
			name: "not found wrong id",
			presetData: map[int64]domain.Instance{
				1: {ID: 1, Name: "Lobby-1"},
			},
			requestID:     2,
			expectedError: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MemoryInstanceStorage{
				data: tt.presetData,
			}
			ctx := context.Background()

			inst, err := storage.GetInstanceByID(ctx, tt.requestID)

			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}

			if err == nil && inst.Name != tt.expectedName {
				t.Errorf("expected name %s, got %s", tt.expectedName, inst.Name)
			}
		})
	}
}

func TestMemoryInstanceStorage_GetInstancesByGameID(t *testing.T) {
	storage := &MemoryInstanceStorage{
		data: map[int64]domain.Instance{
			1: {ID: 1, GameID: 42, Name: "Match-1"},
			2: {ID: 2, GameID: 42, Name: "Match-2"},
			3: {ID: 3, GameID: 99, Name: "Other-Game"},
		},
	}

	ctx := context.Background()
	instances, err := storage.GetInstancesByGameID(ctx, 42)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(instances))
	}

	// Order is not guaranteed in maps, so we check contents
	found1, found2 := false, false
	for _, inst := range instances {
		if inst.ID == 1 {
			found1 = true
		}
		if inst.ID == 2 {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("expected to find instances with ID 1 and 2")
	}
}

func TestMemoryInstanceStorage_GetInstanceByContainerID(t *testing.T) {
	storage := &MemoryInstanceStorage{
		data: map[int64]domain.Instance{
			1: {ID: 1, ContainerID: "docker-abc-123"},
			2: {ID: 2, ContainerID: "docker-xyz-789"},
		},
	}
	ctx := context.Background()

	// 1. Success case
	inst, err := storage.GetInstanceByContainerID(ctx, "docker-xyz-789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inst.ID != 2 {
		t.Errorf("expected instance ID 2, got %d", inst.ID)
	}

	// 2. Not found case
	_, err = storage.GetInstanceByContainerID(ctx, "non-existent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryInstanceStorage_GetAllInstances(t *testing.T) {
	storage := &MemoryInstanceStorage{
		data: map[int64]domain.Instance{
			1: {ID: 1},
			2: {ID: 2},
			3: {ID: 3},
		},
	}

	instances, err := storage.GetAllInstances(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(instances) != 3 {
		t.Errorf("expected 3 instances, got %d", len(instances))
	}
}

func TestMemoryInstanceStorage_DeleteInstance(t *testing.T) {
	storage := &MemoryInstanceStorage{
		data: map[int64]domain.Instance{
			1: {ID: 1, Name: "To-Be-Deleted"},
		},
	}
	ctx := context.Background()

	// 1. Delete existing
	err := storage.DeleteInstance(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(storage.data) != 0 {
		t.Errorf("expected storage to be empty, got %d items", len(storage.data))
	}

	// 2. Delete non-existing
	err = storage.DeleteInstance(ctx, 99)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryInstanceStorage_ConcurrentAccess(t *testing.T) {
	storage := NewMemoryInstanceStorage()
	ctx := context.Background()

	const workers = 100
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := range int64(workers) {
		go func(id int64) {
			defer wg.Done()

			err := storage.RecordInstance(ctx, domain.Instance{
				ID:   id,
				Name: "Concurrent-Lobby",
			})
			if err != nil {
				t.Errorf("failed to record: %v", err)
			}

			_, err = storage.GetInstanceByID(ctx, id)
			if err != nil {
				t.Errorf("failed to get: %v", err)
			}

			_, err = storage.GetAllInstances(ctx)
			if err != nil {
				t.Errorf("failed to get all: %v", err)
			}
		}(i)
	}

	wg.Wait()

	all, _ := storage.GetAllInstances(ctx)
	if len(all) != workers {
		t.Errorf("expected %d instances, got %d", workers, len(all))
	}
}
