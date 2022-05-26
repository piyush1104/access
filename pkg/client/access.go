package client

import (
	"context"
	"errors"

	accesspb "github.com/piyush1104/access/pkg/internal"
)

var (
	// ErrClientNotConnected ...
	ErrClientNotConnected = errors.New("Error: client not connected")
)

func (client *Client) AuthorizeToken(ctx context.Context, token, resource, action string) (bool, error) {
	if !client.Connected() {
		client.logger.Println(ErrClientNotConnected.Error())
		return false, ErrClientNotConnected
	}

	reply, err := client.rpc.AuthorizeToken(ctx, &accesspb.AuthorizeTokenRequest{
		Token:    token,
		Resource: resource,
		Action:   action,
	})
	if err != nil {
		return false, err
	}

	return reply.Authorized, nil
}

func (client *Client) Authorize(ctx context.Context, subject, resource, action string) (bool, error) {
	if !client.Connected() {
		client.logger.Println(ErrClientNotConnected.Error())
		return false, ErrClientNotConnected
	}

	reply, err := client.rpc.Authorize(ctx, &accesspb.AuthorizeRequest{
		Subject:  subject,
		Resource: resource,
		Action:   action,
	})
	if err != nil {
		return false, err
	}

	return reply.Authorized, nil
}
