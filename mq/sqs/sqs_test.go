package sqs

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendMessage(t *testing.T) {
	queueName := testQueueURL(t)
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
	queueName := testQueueURL(t)
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
	queueName := testQueueURL(t)
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

func TestDeleteMessageUsesContext(t *testing.T) {
	transport := &recordingTransport{}
	client := &Client{client: awssqs.NewFromConfig(aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("key", "secret", ""),
		HTTPClient:  &http.Client{Transport: transport},
	})}
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "value")
	queueURL, receipt := "https://sqs.us-east-1.amazonaws.com/123/queue", "receipt"
	_, err := client.DeleteMessage(ctx, &DeleteMessageInput{QueueUrl: &queueURL, ReceiptHandle: &receipt})
	require.Error(t, err)
	assert.Equal(t, "value", transport.ctx.Value(key{}))
}

func testQueueURL(t *testing.T) string {
	t.Helper()
	queueURL := os.Getenv("AWS_SQS_TEST_QUEUE_URL")
	if queueURL == "" {
		t.Skip("set AWS_SQS_TEST_QUEUE_URL to run SQS integration tests")
	}
	return queueURL
}

type recordingTransport struct {
	ctx context.Context
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.ctx = req.Context()
	return nil, errors.New("stop")
}
