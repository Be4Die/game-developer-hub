package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubGameAccess struct{ allowed map[int64]bool }

func (s stubGameAccess) CheckManageServers(_ context.Context, gameID int64) error {
	if s.allowed[gameID] {
		return nil
	}
	return errors.New("forbidden")
}

func ctxWithRole(role uint8) context.Context {
	ctx := context.WithValue(context.Background(), UserIDKey, "user-1")
	return context.WithValue(ctx, UserRoleKey, role)
}

func TestGameAuthorizer(t *testing.T) {
	t.Parallel()

	a := gameAuthorizer{checker: stubGameAccess{allowed: map[int64]bool{10: true}}}

	if err := a.authorize(ctxWithRole(1), 10); err != nil {
		t.Errorf("developer with access: unexpected error %v", err)
	}
	if err := a.authorize(ctxWithRole(1), 11); status.Code(err) != codes.PermissionDenied {
		t.Errorf("developer without access: expected PermissionDenied, got %v", err)
	}
	if err := a.authorize(ctxWithRole(2), 11); err != nil {
		t.Errorf("moderator: unexpected error %v", err)
	}
	if err := a.authorize(ctxWithRole(1), 0); status.Code(err) != codes.InvalidArgument {
		t.Errorf("empty game id: expected InvalidArgument, got %v", err)
	}
}

// Чужие билды и политики недоступны: проверка срабатывает до обращения к сервису.
func TestHandlers_RejectForeignGame(t *testing.T) {
	t.Parallel()

	deny := stubGameAccess{}
	ctx := ctxWithRole(1)

	build := NewBuildHandler(nil).WithGameAccess(deny)
	policy := NewGamePolicyHandler(nil).WithGameAccess(deny)
	instance := NewInstanceHandler(nil, 100).WithGameAccess(deny)

	calls := map[string]func() error{
		"Build.List":   func() error { _, err := build.List(ctx, &pb.BuildServiceListRequest{GameId: 5}); return err },
		"Build.Get":    func() error { _, err := build.Get(ctx, &pb.BuildServiceGetRequest{GameId: 5}); return err },
		"Build.Delete": func() error { _, err := build.Delete(ctx, &pb.BuildServiceDeleteRequest{GameId: 5}); return err },
		"Build.Upload": func() error { _, err := build.Upload(ctx, &pb.BuildServiceUploadRequest{GameId: 5}); return err },
		"Policy.Get":   func() error { _, err := policy.Get(ctx, &pb.GamePolicyServiceGetRequest{GameId: 5}); return err },
		"Policy.Set":   func() error { _, err := policy.Set(ctx, &pb.GamePolicyServiceSetRequest{GameId: 5}); return err },
		"Instance.Start": func() error {
			_, err := instance.Start(ctx, &pb.InstanceServiceStartRequest{GameId: 5})
			return err
		},
	}
	for name, call := range calls {
		if code := status.Code(call()); code != codes.PermissionDenied {
			t.Errorf("%s: expected PermissionDenied, got %v", name, code)
		}
	}
}

func TestBackupTicket(t *testing.T) {
	ticket := generateBackupTicket(7, "game-db", "backup-1", time.Minute)

	if !ValidateBackupTicket(ticket, 7, "game-db", "backup-1") {
		t.Fatal("freshly issued ticket must be valid")
	}
	if ValidateBackupTicket(ticket, 8, "game-db", "backup-1") {
		t.Error("ticket must be bound to the node")
	}
	if ValidateBackupTicket(ticket, 7, "game-db", "backup-2") {
		t.Error("ticket must be bound to the backup")
	}
	if ValidateBackupTicket(generateBackupTicket(7, "game-db", "backup-1", -time.Second), 7, "game-db", "backup-1") {
		t.Error("expired ticket must be rejected")
	}

	// Тикет, подписанный прежним зашитым в код ключом, больше не принимается.
	old := backupTicketSecret
	backupTicketSecret = []byte("gdh-backup-secret-key-2026")
	forged := generateBackupTicket(7, "game-db", "backup-1", time.Minute)
	backupTicketSecret = old
	if ValidateBackupTicket(forged, 7, "game-db", "backup-1") {
		t.Error("ticket signed with the old hardcoded key must be rejected")
	}

	if err := SetBackupTicketSecret("short"); err == nil {
		t.Error("short secret must be rejected")
	}
}
