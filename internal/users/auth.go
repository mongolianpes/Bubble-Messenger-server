package users

import (
	"context"

	pb "bubble/internal/users/proto"
	"time"
)

type activeUserInfo struct {
	Key  string
	Time time.Time
}

func (c *Client) TLS(ctx context.Context, isRegistring bool, clientPublicKey, id string) (string, error) {
	resp, err := c.service.TLS(ctx, &pb.TLSRequest{
		IsRegistring:    isRegistring,
		ClientPublicKey: clientPublicKey,
		Id:              id,
	})
	if err != nil {
		return "", err
	}

	return resp.ServerPublicKey, nil
}

func (c *Client) Register(ctx context.Context, login, name, password, device string) (string, error) {
	resp, err := c.service.Register(ctx, &pb.RegisterRequest{
		Login:    login,
		Name:     name,
		Password: password,
		Device:   device,
	})
	if err != nil {
		return "", err
	}

	return resp.Key, nil
}

func (c *Client) Auth(ctx context.Context, login, password, device string) (string, string, error) {
	resp, err := c.service.Auth(ctx, &pb.AuthRequest{
		Login:    login,
		Password: password,
		Device:   device,
	})
	if err != nil {
		return "", "", err
	}

	return resp.Name, resp.Key, nil
}

func (c *Client) GetAuthInfo(ctx context.Context, device string) (string, int, error) {
	resp, err := c.service.GetAuthInfo(ctx, &pb.GetAuthInfoRequest{
		Device: device,
	})
	if err != nil {
		return "", 0, err
	}

	return resp.Key, int(resp.UserId), nil
}
