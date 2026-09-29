package audiodialog

import (
	pb "bubble/internal/audiodialog/proto"
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	service pb.AudioDialogServiceClient
	conn    *grpc.ClientConn
}

type AudioDialogService interface {
	CreateAudioDialog(ctx context.Context) (dialogID, senderID, receiverID string, err error)
	DeleteAudioDialog(ctx context.Context, dialogID string) error
	Close() error
}

var audioDialogServiceHost = "localhost:50053"

func NewClient() (*Client, error) {
	client := &Client{}
	if client.service != nil {
		return client, nil
	}

	conn, err := grpc.NewClient(audioDialogServiceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Не удалось создать подключение к микросервису Announcements", "error", err)
		return client, err
	}

	client.service = pb.NewAudioDialogServiceClient(conn)
	client.conn = conn
	return client, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
