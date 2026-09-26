package messenger

import (
	"context"
	"log/slog"
	"os"

	pb "bubble/internal/messenger/proto"
	"bubble/internal/users"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	service pb.MessengerServiceClient
	conn    *grpc.ClientConn
}

type MessengerService interface {
	Send(ctx context.Context, usersService users.UsersService, senderLogin, receiverLogin, message string) error
	Check(ctx context.Context, usersService users.UsersService, login string) ([]byte, error)
	Del(ctx context.Context, usersService users.UsersService, login string) error
	Close() error
}

var messengerServiceHost = os.Getenv("MESSENGER_SERVICE_HOST_GRPC_PORT")

func NewClient() (*Client, error) {
	client := &Client{}
	if client.service != nil {
		return client, nil
	}

	conn, err := grpc.NewClient(messengerServiceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Не удалось создать подключение к микросервису Announcements", "error", err)
		return client, err
	}

	client.service = pb.NewMessengerServiceClient(conn)
	client.conn = conn
	return client, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
