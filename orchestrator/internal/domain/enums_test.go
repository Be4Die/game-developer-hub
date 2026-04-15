package domain

import "testing"

func TestInstanceStatus_String(t *testing.T) {
	tests := []struct {
		status InstanceStatus
		want   string
	}{
		{InstanceStatusStarting, "starting"},
		{InstanceStatusRunning, "running"},
		{InstanceStatusStopping, "stopping"},
		{InstanceStatusStopped, "stopped"},
		{InstanceStatusCrashed, "crashed"},
		{InstanceStatus(0), "unknown"},
		{InstanceStatus(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("InstanceStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestProtocol_String(t *testing.T) {
	tests := []struct {
		proto Protocol
		want  string
	}{
		{ProtocolTCP, "tcp"},
		{ProtocolUDP, "udp"},
		{ProtocolWebSocket, "websocket"},
		{ProtocolWebRTC, "webrtc"},
		{Protocol(0), "unknown"},
		{Protocol(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.proto.String(); got != tt.want {
				t.Errorf("Protocol(%d).String() = %q, want %q", tt.proto, got, tt.want)
			}
		})
	}
}

func TestNodeStatus_String(t *testing.T) {
	tests := []struct {
		status NodeStatus
		want   string
	}{
		{NodeStatusUnauthorized, "unauthorized"},
		{NodeStatusOnline, "online"},
		{NodeStatusOffline, "offline"},
		{NodeStatusMaintenance, "maintenance"},
		{NodeStatus(0), "unknown"},
		{NodeStatus(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("NodeStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestLogSource_String(t *testing.T) {
	tests := []struct {
		source LogSource
		want   string
	}{
		{LogSourceStdout, "stdout"},
		{LogSourceStderr, "stderr"},
		{LogSource(0), "unknown"},
		{LogSource(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.source.String(); got != tt.want {
				t.Errorf("LogSource(%d).String() = %q, want %q", tt.source, got, tt.want)
			}
		})
	}
}
