package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GameAccessChecker проверяет право текущего пользователя управлять серверной частью игры.
// Идентификатор игры в оркестраторе совпадает с идентификатором проекта в project-manager,
// поэтому владельца и участников команды знает только он.
type GameAccessChecker interface {
	CheckManageServers(ctx context.Context, gameID int64) error
}

// gameAuthorizer общая проверка доступа к игре для обработчиков билдов, инстансов и политик.
type gameAuthorizer struct {
	checker GameAccessChecker
}

// authorize разрешает операцию модераторам и администраторам, а разработчикам — только
// при подтверждённом доступе к проекту. Без настроенной проверки (локальный запуск без
// project-manager) доступ не ограничивается; в продуктивной среде адрес project-manager обязателен.
func (a gameAuthorizer) authorize(ctx context.Context, gameID int64) error {
	if gameID <= 0 {
		return status.Error(codes.InvalidArgument, "game_id is required")
	}
	if isStaff(ctx) || a.checker == nil {
		return nil
	}
	if err := a.checker.CheckManageServers(ctx, gameID); err != nil {
		return status.Error(codes.PermissionDenied, "no access to manage servers of this game")
	}
	return nil
}
