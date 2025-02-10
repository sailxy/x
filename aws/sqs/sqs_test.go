package sqs

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var queueName = "your-aws-sqs-quene-name"

func TestSendMessage(t *testing.T) {
	ctx := context.Background()
	message := "hello world"
	c, err := New()
	assert.NoError(t, err)

	start := time.Now()
	o, err := c.SendMessage(ctx, &SendMessageInput{
		QueueUrl:    &queueName,
		MessageBody: &message,
	})
	t.Log("send message time", time.Since(start))
	if assert.NoError(t, err) {
		t.Log(*o.MessageId)
	}
}

func TestSendMessageBatch(t *testing.T) {
	ctx := context.Background()

	var msg []SendMessageBatchRequestEntry
	id := "1"
	message := "Hello World1"
	msg = append(msg, SendMessageBatchRequestEntry{
		Id:          &id,
		MessageBody: &message,
	})

	id2 := "2"
	message2 := "Hello World2"
	msg = append(msg, SendMessageBatchRequestEntry{
		Id:          &id2,
		MessageBody: &message2,
	})

	c, err := New()
	assert.NoError(t, err)

	start := time.Now()
	_, err = c.SendMessageBatch(ctx, &SendMessageBatchInput{
		QueueUrl: &queueName,
		Entries:  msg,
	})
	t.Log("send batch message time", time.Since(start))
	assert.NoError(t, err)
}

func TestReceiveMessage(t *testing.T) {
	ctx := context.Background()
	c, err := New()
	assert.NoError(t, err)

	start := time.Now()
	o, err := c.ReceiveMessage(ctx, &ReceiveMessageInput{
		QueueUrl:            &queueName,
		MaxNumberOfMessages: 1,
	})
	t.Log("receive message time", time.Since(start))
	if assert.NoError(t, err) {
		t.Log(*o.Messages[0].Body)
	}
}
