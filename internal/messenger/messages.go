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
		return fmt.Errorf("Cant get info by login %s. ERR %s", senderLogin, err)
	}
	infoReceiver, err := usersService.GetInfoByLogin(ctx, receiverLogin)
	if err != nil {
		return fmt.Errorf("Cant get info by login %s. ERR %s", receiverLogin, err)
	}

	_, err = c.service.SendMessage(ctx, &pb.SendMessageRequest{
		SenderId:   int64(infoSender.ID),
		ReceivedId: int64(infoReceiver.ID),
		Text:       message,
	})
	if err != nil {
		return fmt.Errorf("Failed to send message. Send ID: %v, Received ID: %v, ERR %s", infoSender.ID, infoReceiver.ID, err)
	}

	return nil
}

func (c *Client) SendFile(ctx context.Context, usersService users.UsersService, senderLogin, receiverLogin, fileName, storagePath string) error {
	textMessage := fmt.Sprintf("%s%s\\%s", models.IdentificatorSendFile, fileName, storagePath)
	return c.Send(ctx, usersService, senderLogin, receiverLogin, textMessage)
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

		if len(newMessages) == 0 {
			newMessages = messagesBytes
		} else {
			newMessages = fmt.Appendf(newMessages, "\n%s", messagesBytes)
		}
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
