package users

import (
	"bubble/internal/models"
	pb "bubble/internal/users/proto"
	"context"
)

func (c *Client) Search(ctx context.Context, login string) ([]models.FindUser, error) {
	resp, err := c.service.Search(ctx, &pb.SearchRequest{
		Query: login,
	})
	if err != nil {
		return nil, err
	}

	result := []models.FindUser{}
	for _, user := range resp.Users {
		result = append(result, models.FindUser{
			Login: user.Login,
			Name:  user.Name,
			ID:    int(user.Id),
		})
	}

	return result, nil
}

func (c *Client) GetInfoByLogin(ctx context.Context, login string) (models.FindUser, error) {
	result, err := c.Search(ctx, "@"+login)
	return result[0], err
}

func (c *Client) GetInfoByID(ctx context.Context, id int) (models.FindUser, error) {
	resp, err := c.service.GetInfoByID(ctx, &pb.GetInfoByIDRequest{
		Id: int64(id),
	})
	if err != nil {
		return models.FindUser{}, err
	}

	return models.FindUser{
		Login: resp.Login,
		Name:  resp.Name,
	}, nil
}
