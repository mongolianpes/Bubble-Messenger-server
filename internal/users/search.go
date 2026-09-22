package users

import (
	"context"
	"server/internal/models"
	pb "server/internal/users/proto"
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
		})
	}

	return result, nil
}
