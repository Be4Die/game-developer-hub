package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractAuthUser_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/1/chat/attachments/test", nil)
	userID, userRole, rawToken := extractAuthUser(req)
	assert.Empty(t, userID)
	assert.Empty(t, userRole)
	assert.Empty(t, rawToken)
}

func TestExtractAuthUser_FromQueryToken(t *testing.T) {
	// payload: {"sub":"user-42","role":"moderator"}
	// header.payload.signature
	fakeToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTQyIiwicm9sZSI6Im1vZGVyYXRvciJ9.fakeSig"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/1/chat/attachments/test?token="+fakeToken, nil)
	userID, userRole, rawToken := extractAuthUser(req)
	assert.Equal(t, "user-42", userID)
	assert.Equal(t, "moderator", userRole)
	assert.Equal(t, fakeToken, rawToken)
}

func TestIsRoleModeratorOrAdmin(t *testing.T) {
	assert.True(t, isRoleModeratorOrAdmin("admin"))
	assert.True(t, isRoleModeratorOrAdmin("moderator"))
	assert.True(t, isRoleModeratorOrAdmin("USER_ROLE_ADMIN"))
	assert.True(t, isRoleModeratorOrAdmin("USER_ROLE_MODERATOR"))
	assert.True(t, isRoleModeratorOrAdmin("2"))
	assert.True(t, isRoleModeratorOrAdmin("3"))
	assert.False(t, isRoleModeratorOrAdmin("developer"))
	assert.False(t, isRoleModeratorOrAdmin("user"))
	assert.False(t, isRoleModeratorOrAdmin(""))
}

func TestCheckProjectAccess_ModeratorAllowed(t *testing.T) {
	ctx := context.Background()
	allowed := checkProjectAccess(ctx, nil, 100, "user-mod", "moderator", "")
	assert.True(t, allowed)
}

func TestCheckProjectAccess_Unauthenticated(t *testing.T) {
	ctx := context.Background()
	allowed := checkProjectAccess(ctx, nil, 100, "", "", "")
	assert.False(t, allowed)
}

func TestNewUUID(t *testing.T) {
	u1 := newUUID()
	u2 := newUUID()
	assert.Len(t, u1, 36)
	assert.Len(t, u2, 36)
	assert.NotEqual(t, u1, u2)
}
