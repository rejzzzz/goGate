package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Client represents a RabbitMQ client that handles automatic reconnections.
type Client struct {
	url    string
	logger *zap.Logger
	conn   *amqp.Connection
	ch     *amqp.Channel
	mu     sync.RWMutex
	done   chan struct{}
}

// NewClient creates a new RabbitMQ client and starts the connection loop in the background.
func NewClient(url string, logger *zap.Logger) *Client {
	c := &Client{
		url:    url,
		logger: logger,
		done:   make(chan struct{}),
	}
	go c.connectLoop()
	return c
}

func (c *Client) connectLoop() {
	for {
		select {
		case <-c.done:
			return
		default:
		}

		conn, err := amqp.Dial(c.url)
		if err != nil {
			c.logger.Error("Failed to connect to RabbitMQ, retrying in 5s", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			c.logger.Error("Failed to open channel, retrying in 5s", zap.Error(err))
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		c.mu.Lock()
		c.conn = conn
		c.ch = ch
		c.mu.Unlock()

		c.logger.Info("Connected to RabbitMQ")

		errCh := make(chan *amqp.Error)
		conn.NotifyClose(errCh)

		select {
		case <-c.done:
			conn.Close()
			return
		case err := <-errCh:
			if err != nil {
				c.logger.Error("RabbitMQ connection closed", zap.Error(err))
			}
			c.mu.Lock()
			c.conn = nil
			c.ch = nil
			c.mu.Unlock()
		}
	}
}

// Publish sends a message to the specified exchange with the given routing key.
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, payload []byte) error {
	c.mu.RLock()
	ch := c.ch
	c.mu.RUnlock()

	if ch == nil {
		return fmt.Errorf("rabbitmq not connected")
	}

	return ch.PublishWithContext(ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         payload,
		})
}

// Close gracefully closes the RabbitMQ connection.
func (c *Client) Close() {
	close(c.done)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
