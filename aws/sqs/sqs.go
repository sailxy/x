package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Message = types.Message
type SendMessageInput = sqs.SendMessageInput
type SendMessageOutput = sqs.SendMessageOutput
type SendMessageBatchInput = sqs.SendMessageBatchInput
type SendMessageBatchOutput = sqs.SendMessageBatchOutput
type SendMessageBatchRequestEntry = types.SendMessageBatchRequestEntry
type ReceiveMessageInput = sqs.ReceiveMessageInput
type ReceiveMessageOutput = sqs.ReceiveMessageOutput
type DeleteMessageInput = sqs.DeleteMessageInput
type DeleteMessageOutput = sqs.DeleteMessageOutput

type Client struct {
	client *sqs.Client
}

func New() (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := sqs.NewFromConfig(cfg)

	return &Client{
		client: client,
	}, nil
}

func (c *Client) SendMessage(ctx context.Context, input *SendMessageInput) (*SendMessageOutput, error) {
	return c.client.SendMessage(ctx, input)
}

func (c *Client) SendMessageBatch(ctx context.Context, input *SendMessageBatchInput) (*SendMessageBatchOutput, error) {
	return c.client.SendMessageBatch(ctx, input)
}
func (c *Client) ReceiveMessage(ctx context.Context, input *ReceiveMessageInput) (*ReceiveMessageOutput, error) {
	return c.client.ReceiveMessage(ctx, input)
}

func (c *Client) DeleteMessage(ctx context.Context, input *DeleteMessageInput) (*DeleteMessageOutput, error) {
	return c.client.DeleteMessage(context.Background(), input)
}
