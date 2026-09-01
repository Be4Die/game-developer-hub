package main

import (
	"bytes"
	"testing"

	gwpb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProtocol(t *testing.T) {
	assert.Equal(t, gwpb.Protocol_PROTOCOL_TCP, parseProtocol("tcp"))
	assert.Equal(t, gwpb.Protocol_PROTOCOL_UDP, parseProtocol("udp"))
	assert.Equal(t, gwpb.Protocol_PROTOCOL_WEBSOCKET, parseProtocol("websocket"))
	assert.Equal(t, gwpb.Protocol_PROTOCOL_WEBRTC, parseProtocol("webrtc"))
	assert.Equal(t, gwpb.Protocol_PROTOCOL_WEBSOCKET, parseProtocol("unknown"))
}

func TestParseUint32(t *testing.T) {
	assert.Equal(t, uint32(8080), parseUint32("", 8080))
	assert.Equal(t, uint32(9000), parseUint32("9000", 8080))
	assert.Equal(t, uint32(8080), parseUint32("invalid", 8080))
}

func TestStreamFileChunks(t *testing.T) {
	data := []byte("testing chunk streaming functionality in gateway")
	r := bytes.NewReader(data)

	var received bytes.Buffer
	err := streamFileChunks(r, func(chunk []byte) error {
		received.Write(chunk)
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, data, received.Bytes())
}
