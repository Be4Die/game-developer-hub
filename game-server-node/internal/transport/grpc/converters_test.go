package grpc

import (
	"errors"
	"testing"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDomainErrToStatus(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode codes.Code
	}{
		{
			name:         "not found error translates to NotFound",
			err:          domain.ErrNotFound,
			expectedCode: codes.NotFound,
		},
		{
			name:         "already exists error translates to AlreadyExists",
			err:          domain.ErrAlreadyExists,
			expectedCode: codes.AlreadyExists,
		},
		{
			name:         "unknown generic error translates to Internal",
			err:          errors.New("database connection lost"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grpcErr := domainErrToStatus(tt.err)

			// Извлекаем gRPC статус из ошибки
			st, ok := status.FromError(grpcErr)
			if !ok {
				t.Fatalf("expected gRPC status error, got %v", grpcErr)
			}

			if st.Code() != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, st.Code())
			}
		})
	}
}

func TestInstanceToProto(t *testing.T) {
	// Arrange
	now := time.Now()
	players := uint32(5)

	inst := &domain.Instance{
		ID:          42,
		Name:        "Test-Server",
		GameID:      1,
		Port:        27015,
		Protocol:    domain.ProtocolUDP,
		Status:      domain.InstanceStatusRunning,
		PlayerCount: &players,
		MaxPlayers:  10,
		StartedAt:   now,
	}

	// Act
	pbInst := instanceToProto(inst)

	// Assert
	if pbInst.InstanceId != 42 {
		t.Errorf("expected ID 42, got %d", pbInst.InstanceId)
	}
	if pbInst.Protocol != pb.Protocol_PROTOCOL_UDP {
		t.Errorf("expected UDP protocol, got %v", pbInst.Protocol)
	}
	if pbInst.Status != pb.InstanceStatus_INSTANCE_STATUS_RUNNING {
		t.Errorf("expected Running status, got %v", pbInst.Status)
	}
	if pbInst.PlayerCount == nil || *pbInst.PlayerCount != 5 {
		t.Errorf("expected player count 5, got %v", pbInst.PlayerCount)
	}
}

func TestProtoToPortStrategy(t *testing.T) {
	// Проверяем маппинг стратегий выделения портов
	tests := []struct {
		name     string
		input    *pb.PortAllocation
		expected domain.PortStrategy
	}{
		{
			name:     "nil translates to Any",
			input:    nil,
			expected: domain.PortStrategy{Any: true},
		},
		{
			name: "exact port",
			input: &pb.PortAllocation{
				Strategy: &pb.PortAllocation_Exact{Exact: 7777},
			},
			expected: domain.PortStrategy{Exact: 7777},
		},
		{
			name: "port range",
			input: &pb.PortAllocation{
				Strategy: &pb.PortAllocation_Range{
					Range: &pb.PortRange{MinPort: 30000, MaxPort: 30010},
				},
			},
			expected: domain.PortStrategy{Range: &domain.PortRange{Min: 30000, Max: 30010}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := protoToPortStrategy(tt.input)
			if res.Any != tt.expected.Any {
				t.Errorf("Any mismatch: got %v, want %v", res.Any, tt.expected.Any)
			}
			if res.Exact != tt.expected.Exact {
				t.Errorf("Exact mismatch: got %v, want %v", res.Exact, tt.expected.Exact)
			}
			if tt.expected.Range != nil && (res.Range == nil || res.Range.Min != tt.expected.Range.Min) {
				t.Errorf("Range mismatch")
			}
		})
	}
}
