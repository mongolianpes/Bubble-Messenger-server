package messenger

import (
	pb "bubble/internal/messenger/proto"
	"bubble/internal/models"
	"context"
	"encoding/json"
	"fmt"

	"bubble/internal/users"
)

func (c *Client) Send(ctx context.Context, usersService users.UsersService, senderLogin, receiverLogin, message string) error {
	infoSender, err := usersService.GetInfoByLogin(ctx, senderLogin)
	if err != nil {
		return err
	}
	infoReceiver, err := usersService.GetInfoByLogin(ctx, receiverLogin)
	if err != nil {
		return err
	}

	_, err = c.service.SendMessage(ctx, &pb.SendMessageRequest{
		SenderId:   int64(infoSender.ID),
		ReceivedId: int64(infoReceiver.ID),
		Text:       message,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Check(ctx context.Context, usersService users.UsersService, login string) ([]byte, error) {
	newMessages := []byte{}

	userInfo, err := usersService.GetInfoByLogin(ctx, login)
	if err != nil {
		return newMessages, err
	}

	resp, err := c.service.CheckChats(ctx, &pb.CheckChatsRequest{
		UserId: int64(userInfo.ID),
	})
	if err != nil {
		return newMessages, err
	}

	for _, message := range resp.Messages {
		senderInfo, err := usersService.GetInfoByID(ctx, int(message.SenderId))
		if err != nil {
			return newMessages, err
		}

		messageStruct := models.Message{
			SenderLogin: senderInfo.Login,
			Text:        message.Text,
			SendTime:    message.CreateAt.AsTime(),
		}

		messagesBytes, err := json.Marshal(&messageStruct)
		if err != nil {
			return newMessages, err
		}

		newMessages = fmt.Appendf(nil, "%s\\\\%s", newMessages, messagesBytes)
	}

	return newMessages, nil
}

func (c *Client) Del(ctx context.Context, usersService users.UsersService, login string) error {
	userInfo, err := usersService.GetInfoByLogin(ctx, login)
	if err != nil {
		return err
	}

	_, err = c.service.DelMessage(ctx, &pb.DelMessageRequest{
		UserId: int64(userInfo.ID),
	})
	if err != nil {
		return err
	}

	return nil
}
