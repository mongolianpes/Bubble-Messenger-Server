package users

import (
	"context"
	"fmt"

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

	switch m := resp.AuthMaterial.(type) {
	case *pb.AuthResponse_Key:
		return resp.UserName, m.Key, nil
	case *pb.AuthResponse_SessionID:
		return resp.UserName, m.SessionID, nil
	default:
		return "", "", fmt.Errorf("unknown auth material type: %T", m)
	}
}

func (c *Client) GetAuthInfo(ctx context.Context, login, password, device string) (string, int, error) {
	resp, err := c.service.GetAuthInfo(ctx, &pb.GetAuthInfoRequest{
		Device:   device,
		Login:    login,
		Password: password,
	})
	if err != nil {
		return "", 0, err
	}

	return resp.Key, int(resp.UserId), nil
}
