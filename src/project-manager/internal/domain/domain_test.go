package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDraft_IsReadyForModeration(t *testing.T) {
	tests := []struct {
		name    string
		draft   Draft
		wantErr error
	}{
		{
			name: "valid with ru and en fields and active build",
			draft: Draft{
				TitleRu:            "Тестовая игра",
				TitleEn:            "Test Game",
				AboutRu:            "Описание игры",
				AboutEn:            "Description of the game",
				ActiveBuildVersion: "v1.0.0",
			},
			wantErr: nil,
		},
		{
			name: "valid with ru only",
			draft: Draft{
				TitleRu:            "Тестовая игра",
				AboutRu:            "Описание игры",
				ActiveBuildVersion: "v1.0.0",
			},
			wantErr: nil,
		},
		{
			name: "valid with en only",
			draft: Draft{
				TitleEn:            "Test Game",
				AboutEn:            "Description of the game",
				ActiveBuildVersion: "v1.0.0",
			},
			wantErr: nil,
		},
		{
			name: "missing both titles",
			draft: Draft{
				AboutRu:            "Описание",
				ActiveBuildVersion: "v1.0.0",
			},
			wantErr: ErrDraftNotReady,
		},
		{
			name: "missing both abouts",
			draft: Draft{
				TitleRu:            "Игра",
				ActiveBuildVersion: "v1.0.0",
			},
			wantErr: ErrDraftNotReady,
		},
		{
			name: "missing active build version",
			draft: Draft{
				TitleRu: "Игра",
				AboutRu: "Описание",
			},
			wantErr: ErrDraftNotReady,
		},
		{
			name:    "empty draft",
			draft:   Draft{},
			wantErr: ErrDraftNotReady,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.draft.IsReadyForModeration()
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDomain_EntitiesInitialization(t *testing.T) {
	now := time.Now()
	proj := Project{
		ID:        1,
		OwnerID:   "user-123",
		Status:    ProjectStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
		Draft: &Draft{
			ProjectID: 1,
			TitleRu:   "Title",
		},
		Release: &Release{
			ProjectID: 1,
			Version:   "v1.0.0",
		},
	}

	assert.Equal(t, int64(1), proj.ID)
	assert.Equal(t, ProjectStatusDraft, proj.Status)
	assert.NotNil(t, proj.Draft)
	assert.NotNil(t, proj.Release)

	build := Build{
		ID:           10,
		ProjectID:    1,
		Version:      "v1.0.0",
		FilePath:     "/builds/1/v1.0.0.zip",
		FileSize:     1024,
		IsUnpacked:   true,
		UnpackedPath: "/builds/unpacked/1/v1.0.0",
		CreatedAt:    now,
	}
	assert.Equal(t, "v1.0.0", build.Version)
	assert.True(t, build.IsUnpacked)

	dep := DeploymentRecord{
		ID:          100,
		ProjectID:   1,
		Environment: DeploymentEnvProd,
		Version:     "v1.0.0",
		Status:      DeploymentStatusSuccess,
		DeployedAt:  now,
	}
	assert.Equal(t, "v1.0.0", dep.Version)
	assert.Equal(t, DeploymentStatusSuccess, dep.Status)
}
