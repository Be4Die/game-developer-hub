// Package projectmanager реализует клиент сервиса project-manager для проверки доступа к играм.
package projectmanager

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pmpb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
)

// permManageServers — право участника команды управлять серверами проекта.
const permManageServers = "PERM_MANAGE_SERVERS"

// ErrForbidden возвращается, когда у пользователя нет права управлять серверами игры.
var ErrForbidden = errors.New("forbidden")

// Client проверяет доступ пользователя к проекту через project-manager.
type Client struct {
	client pmpb.ProjectServiceClient
	conn   *grpc.ClientConn
}

// New создаёт клиент. Соединение устанавливается лениво при первом вызове.
func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("projectmanager.New: %w", err)
	}
	return &Client{client: pmpb.NewProjectServiceClient(conn), conn: conn}, nil
}

// Close закрывает соединение.
func (c *Client) Close() error {
	return c.conn.Close()
}

// CheckManageServers проверяет, что текущий пользователь — владелец проекта либо участник
// с правом управления серверами. Личность определяет project-manager по проброшенному JWT.
func (c *Client) CheckManageServers(ctx context.Context, gameID int64) error {
	const op = "projectmanager.Client.CheckManageServers"

	outCtx := ctx
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := md.Get("authorization"); len(auth) > 0 {
			outCtx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	resp, err := c.client.Get(outCtx, &pmpb.ProjectGetRequest{Id: gameID})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	project := resp.GetProject()
	if project == nil {
		return fmt.Errorf("%s: %w", op, ErrForbidden)
	}
	if project.GetIsOwner() {
		return nil
	}
	for _, perm := range project.GetCurrentUserPermissions() {
		if perm == permManageServers {
			return nil
		}
	}
	return fmt.Errorf("%s: %w", op, ErrForbidden)
}
