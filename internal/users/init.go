package users

import (
	"context"
	"log/slog"
	"os"

	"bubble/internal/models"
	pb "bubble/internal/users/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	service pb.UsersServiceClient
	conn    *grpc.ClientConn
}

type UsersService interface {
	TLS(ctx context.Context, isRegistring bool, clientPublicKey, id string) (string, error)
	Register(ctx context.Context, login, name, password, device string) (string, error)
	Auth(ctx context.Context, login, password, device string) (string, string, error)
	GetAuthInfo(ctx context.Context, device string) (string, int, error)
	Search(ctx context.Context, login string) ([]models.FindUser, error)
	Close() error
}

var usersServiceHost = os.Getenv("USERS_SERVICE_HOST_GRPC_PORT")

func NewClient() (*Client, error) {
	client := &Client{}
	if client.service != nil {
		return client, nil
	}

	conn, err := grpc.NewClient(usersServiceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Не удалось создать подключение к микросервису Announcements", "error", err)
		return client, err
	}

	client.service = pb.NewUsersServiceClient(conn)
	client.conn = conn
	return client, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
