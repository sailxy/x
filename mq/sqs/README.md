# AWS SQS 消息队列

`mq/sqs` 提供 AWS SQS 的消息发送、批量发送、接收和删除操作，并沿用 AWS SDK 的请求和响应类型。

迁移：将 `github.com/sailxy/x/aws/sqs` 改为 `github.com/sailxy/x/mq/sqs`。本包不提供跨平台统一消息队列接口。

```go
client, err := sqs.New()

result, err := client.SendMessage(ctx, &sqs.SendMessageInput{
	QueueUrl:    aws.String("https://sqs.us-east-1.amazonaws.com/123/queue"),
	MessageBody: aws.String("hello"),
})
```

SQS 是托管队列 API；RabbitMQ 额外提供 exchange、绑定、ack/nack 和 QoS，因此两个包保持独立 API。
