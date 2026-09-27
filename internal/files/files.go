package files

import (
	"context"
	"os"

	pb "bubble/internal/files/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	service pb.FilesServiceClient
	conn    *grpc.ClientConn
}

type FilesService interface {
	SaveFile(ctx context.Context, file []byte) (string, error)
}

var usersServiceHost = os.Getenv("FILES_SERVICE_HOST_GRPC_PORT")

func NewClient() (*Client, error) {
	client := &Client{}
	if client.service != nil {
		return client, nil
	}

	conn, err := grpc.NewClient(usersServiceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return client, err
	}

	client.service = pb.NewFilesServiceClient(conn)
	client.conn = conn
	return client, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) SaveFile(ctx context.Context, file []byte) (string, error) {
	resp, err := c.service.SaveFile(ctx, &pb.SaveFileRequest{
		File: file,
	})
	if err != nil {
		return "", err
	}

	return resp.StoragePath, nil
}
