package audiodialog

import (
	pb "bubble/internal/audiodialog/proto"
	"context"
)

func (c *Client) CreateAudioDialog(ctx context.Context) (dialogID, senderID, receiverID string, err error) {
	var resp *pb.CreateAudioDialogResponse
	resp, err = c.service.CreateAudioDialog(ctx, &pb.CreateAudioDialogRequest{})
	if err != nil {
		return "", "", "", err
	}

	dialogID = resp.DialogId
	senderID = resp.SenderId
	receiverID = resp.ReceiverId
	return
}

func (c *Client) DeleteAudioDialog(ctx context.Context, dialogID string) error {
	_, err := c.service.DeleteAudioDialog(ctx, &pb.DeleteAudioDialogRequest{DialogId: dialogID})
	return err
}
