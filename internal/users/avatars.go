package users

import (
	pb "bubble/internal/users/proto"
	"context"
)

func (c *Client) SetAvatar(ctx context.Context, login, avatarPath string) error {
	_, err := c.service.AddUserAvatarPath(ctx, &pb.AddUserAvatarPathRequest{
		Login:             login,
		StorageAvatarPath: avatarPath,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetAvatar(ctx context.Context, login string) (string, error) {
	resp, err := c.service.GetUserAvatarPath(ctx, &pb.GetUserAvatarPathRequest{
		Login: login,
	})
	if err != nil {
		return "", err
	}

	return resp.StoragePath, nil
}
