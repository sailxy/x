package rabbitmq

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	URL string
}

// ExchangeDeclareConfig keeps exchange declaration arguments explicit so callers
// can see the exact flags they are sending to RabbitMQ.
type ExchangeDeclareConfig struct {
	Type       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

// QueueDeclareConfig keeps queue declaration arguments explicit so callers can
// see the exact flags they are sending to RabbitMQ.
type QueueDeclareConfig struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

// QueueBindConfig keeps queue binding arguments explicit so callers can see the
// exact flags they are sending to RabbitMQ.
type QueueBindConfig struct {
	NoWait bool
	Args   amqp.Table
}

// ConsumeConfig keeps consumer arguments explicit so callers can see the exact
// flags they are sending to RabbitMQ.
type ConsumeConfig struct {
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

// QosConfig keeps QoS arguments explicit so callers can see the exact values
// they are sending to RabbitMQ.
type QosConfig struct {
	PrefetchCount int
	PrefetchSize  int
	Global        bool
}

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQ(c Config) (*RabbitMQ, error) {
	conn, err := amqp.Dial(c.URL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		conn: conn,
		ch:   ch,
	}, nil
}

func (r *RabbitMQ) Close() error {
	return errors.Join(r.ch.Close(), r.conn.Close())
}

// ExchangeDeclare declares an exchange using the flags provided by the caller.
func (r *RabbitMQ) ExchangeDeclare(name string, c ExchangeDeclareConfig) error {
	return r.ch.ExchangeDeclare(name, c.Type, c.Durable, c.AutoDelete, c.Internal, c.NoWait, c.Args)
}

// QueueDeclare declares a queue using the flags provided by the caller.
func (r *RabbitMQ) QueueDeclare(name string, c QueueDeclareConfig) (amqp.Queue, error) {
	return r.ch.QueueDeclare(name, c.Durable, c.AutoDelete, c.Exclusive, c.NoWait, c.Args)
}

// QueueBind binds a queue to an exchange using the arguments provided by the caller.
func (r *RabbitMQ) QueueBind(name, key, exchange string, c QueueBindConfig) error {
	return r.ch.QueueBind(name, key, exchange, c.NoWait, c.Args)
}

// Publish publishes a message using the arguments provided by the caller.
func (r *RabbitMQ) Publish(ctx context.Context, exchange, key string, msg []byte) error {
	mandatory := false
	immediate := false

	return r.ch.PublishWithContext(ctx, exchange, key, mandatory, immediate, amqp.Publishing{
		Body: msg,
	})
}

// Consume starts consuming deliveries using the arguments provided by the caller.
func (r *RabbitMQ) Consume(ctx context.Context, queue string, c ConsumeConfig) (<-chan amqp.Delivery, error) {
	return r.ch.ConsumeWithContext(ctx, queue, c.Consumer, c.AutoAck, c.Exclusive, c.NoLocal, c.NoWait, c.Args)
}

// Qos sets the channel QoS using the values provided by the caller.
func (r *RabbitMQ) Qos(c QosConfig) error {
	return r.ch.Qos(c.PrefetchCount, c.PrefetchSize, c.Global)
}

// Ack acknowledges a delivery tag using the values provided by the caller.
func (r *RabbitMQ) Ack(tag uint64, multiple bool) error {
	return r.ch.Ack(tag, multiple)
}

// Reject rejects a delivery tag using the values provided by the caller.
func (r *RabbitMQ) Reject(tag uint64, requeue bool) error {
	return r.ch.Reject(tag, requeue)
}

// Nack negatively acknowledges a delivery tag using the values provided by the caller.
func (r *RabbitMQ) Nack(tag uint64, multiple, requeue bool) error {
	return r.ch.Nack(tag, multiple, requeue)
}
