package domain

import (
	"context"
	"io"
	"time"
)

// ContainerRuntime manages container lifecycle.
type ContainerRuntime interface {
	// Image operations.
	LoadImage(ctx context.Context, imageTag string, data io.Reader) error

	// Container lifecycle.
	CreateContainer(ctx context.Context, opts ContainerOpts) (containerID string, err error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	RemoveContainer(ctx context.Context, containerID string) error

	// Observability.
	ContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error)
	ContainerStats(ctx context.Context, containerID string) (ResourcesUsage, error)
}

// ContainerOpts holds parameters for creating a container.
type ContainerOpts struct {
	ImageTag     string
	InternalPort uint32
	HostPort     uint32
	EnvVars      map[string]string
	Args         []string
	CPUMillis    *uint32
	MemoryBytes  *uint64
}
