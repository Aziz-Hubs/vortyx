// Package mq provides message queue integration for reliable agent-backend communication.
// This package defines interfaces and an in-memory implementation for message publishing
// and subscribing, crucial for inter-service communication within the Vortyx agent.
package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// Error Definitions
// -----------------------------------------------------------------------------

var (
	// ErrNotConnected indicates that the message queue is not connected or initialized.
	ErrNotConnected = errors.New("not connected to message queue")
	// ErrPublishFailed indicates a failure to publish a message, possibly due to a full queue.
	ErrPublishFailed = errors.New("failed to publish message")
)

// -----------------------------------------------------------------------------
// Message Structure
// -----------------------------------------------------------------------------

// Message represents a standardized message format for communication within the Vortyx system.
// It includes metadata for routing, identification, and retry mechanisms.
type Message struct {
	ID         string                 `json:"id"`                 // Unique identifier for the message.
	Type       string                 `json:"type"`               // Categorization of the message (e.g., "command", "data", "heartbeat").
	AgentID    string                 `json:"agent_id,omitempty"` // Optional: ID of the target or source agent.
	Payload    map[string]interface{} `json:"payload"`            // The actual data content of the message.
	Timestamp  time.Time              `json:"timestamp"`          // Time when the message was created.
	RetryCount int                    `json:"retry_count"`        // Number of times the message has been retried.
}

// -----------------------------------------------------------------------------
// Publisher Interface
// -----------------------------------------------------------------------------

// Publisher defines the interface for sending messages to the message queue.
// Implementations can vary (e.g., in-memory, RabbitMQ, Kafka).
type Publisher interface {
	// Publish sends a message to a specified exchange and routing key.
	Publish(ctx context.Context, exchange, routingKey string, msg *Message) error
	// PublishWithDelay sends a message after a specified delay.
	PublishWithDelay(ctx context.Context, exchange, routingKey string, msg *Message, delay time.Duration) error
	// PublishCommand is a convenience method for publishing agent commands.
	PublishCommand(ctx context.Context, agentID string, msg *Message) error
	// PublishAgentData is a convenience method for publishing agent-generated data.
	PublishAgentData(ctx context.Context, agentID string, msg *Message) error
	// PublishHeartbeat is a convenience method for publishing agent heartbeats.
	PublishHeartbeat(ctx context.Context, agentID string, msg *Message) error
	// Close shuts down the publisher, releasing any resources.
	Close() error
}

// -----------------------------------------------------------------------------
// Subscriber Interface
// -----------------------------------------------------------------------------

// Subscriber defines the interface for receiving messages from the message queue.
type Subscriber interface {
	// Subscribe registers a handler function for a specific queue.
	Subscribe(ctx context.Context, queue string, handler MessageHandler) error
	// Unsubscribe removes a handler for a specific queue.
	Unsubscribe(queue string) error
	// Close shuts down the subscriber, releasing any resources.
	Close() error
}

// MessageHandler defines the function signature for processing incoming messages.
type MessageHandler func(ctx context.Context, msg *Message) error

// -----------------------------------------------------------------------------
// InMemoryMQ Implementation
// -----------------------------------------------------------------------------

// InMemoryMQ is a simple, non-persistent, thread-safe in-memory message queue implementation.
// It's suitable for testing and scenarios where message persistence is not required.
type InMemoryMQ struct {
	mu          sync.RWMutex              // Mutex to protect concurrent access to queues and handlers.
	queues      map[string][]*Message     // Stores messages for each queue.
	handlers    map[string]MessageHandler // Maps routing keys to their respective message handlers.
	connected   bool                      // Indicates if the MQ is considered connected.
	exchange    string                    // Default exchange name for publishing.
	queuePrefix string                    // Prefix for queue names to ensure uniqueness.
}

// InMemoryMQConfig holds configuration parameters for the InMemoryMQ.
type InMemoryMQConfig struct {
	Exchange    string // The default exchange name.
	QueuePrefix string // A prefix to apply to all queue names.
}

// NewInMemoryMQ creates and initializes a new InMemoryMQ instance.
func NewInMemoryMQ(config *InMemoryMQConfig) *InMemoryMQ {
	return &InMemoryMQ{
		queues:      make(map[string][]*Message),
		handlers:    make(map[string]MessageHandler),
		connected:   true,
		exchange:    config.Exchange,
		queuePrefix: config.QueuePrefix,
	}
}

// Publish sends a message to the specified exchange and routing key in the in-memory queue.
// If a handler is registered for the routing key, it's invoked asynchronously.
func (m *InMemoryMQ) Publish(ctx context.Context, exchange, routingKey string, msg *Message) error {
	if !m.connected {
		return ErrNotConnected
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Assign a unique ID and timestamp if not already set.
	if msg.ID == "" {
		msg.ID = generateMessageID()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	fullQueue := m.queuePrefix + routingKey
	m.queues[fullQueue] = append(m.queues[fullQueue], msg)

	// If a handler exists for this routing key, execute it in a new goroutine.
	if handler, exists := m.handlers[routingKey]; exists && handler != nil {
		go handler(ctx, msg)
	}

	return nil
}

// PublishWithDelay publishes a message after a specified duration.
// The message is published asynchronously after the delay.
func (m *InMemoryMQ) PublishWithDelay(ctx context.Context, exchange, routingKey string, msg *Message, delay time.Duration) error {
	go func() {
		time.Sleep(delay)
		// Re-use the standard Publish method after the delay.
		m.Publish(ctx, exchange, routingKey, msg)
	}()
	return nil
}

// PublishCommand publishes a message specifically as a command for a given agent.
// It uses a predefined routing key format for commands.
func (m *InMemoryMQ) PublishCommand(ctx context.Context, agentID string, msg *Message) error {
	routingKey := fmt.Sprintf("command.%s", agentID)
	return m.Publish(ctx, m.exchange, routingKey, msg)
}

// PublishAgentData publishes a message containing data from a given agent.
// It uses a predefined routing key format for agent data.
func (m *InMemoryMQ) PublishAgentData(ctx context.Context, agentID string, msg *Message) error {
	routingKey := fmt.Sprintf("data.%s", agentID)
	return m.Publish(ctx, m.exchange, routingKey, msg)
}

// PublishHeartbeat publishes a heartbeat message from a given agent.
// It uses a predefined routing key format for heartbeats.
func (m *InMemoryMQ) PublishHeartbeat(ctx context.Context, agentID string, msg *Message) error {
	routingKey := fmt.Sprintf("heartbeat.%s", agentID)
	return m.Publish(ctx, m.exchange, routingKey, msg)
}

// Subscribe registers a MessageHandler for a given queue (routing key).
// Messages published to this routing key will be processed by the handler.
func (m *InMemoryMQ) Subscribe(ctx context.Context, queue string, handler MessageHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers[queue] = handler
	// Initialize the queue if it doesn't exist.
	m.queues[m.queuePrefix+queue] = []*Message{}

	return nil
}

// Unsubscribe removes the handler and clears the queue for a given routing key.
func (m *InMemoryMQ) Unsubscribe(queue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.handlers, queue)
	delete(m.queues, m.queuePrefix+queue)

	return nil
}

// Close marks the InMemoryMQ as disconnected and clears all queues and handlers.
func (m *InMemoryMQ) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connected = false
	m.queues = make(map[string][]*Message)
	m.handlers = make(map[string]MessageHandler)

	return nil
}

// GetQueueMessages retrieves all messages currently in a specific queue.
// This is primarily for testing and inspection purposes.
func (m *InMemoryMQ) GetQueueMessages(queue string) []*Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.queues[m.queuePrefix+queue]
}

// IsConnected returns the connection status of the InMemoryMQ.
func (m *InMemoryMQ) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// generateMessageID creates a simple unique ID for messages based on current time.
func generateMessageID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// -----------------------------------------------------------------------------
// MessageBus Implementation (Worker Pool for Message Processing)
// -----------------------------------------------------------------------------

// MessageBus provides a concurrent message processing mechanism using a worker pool.
// It allows multiple subscribers to process messages of a specific type.
type MessageBus struct {
	mu          sync.RWMutex                // Mutex to protect concurrent access to subscribers.
	subscribers map[string][]MessageHandler // Maps message types to a list of handlers.
	queue       chan *Message               // Buffered channel for incoming messages.
	workers     int                         // Number of worker goroutines to process messages.
	connected   bool                        // Indicates if the message bus is running.
}

// NewMessageBus creates and initializes a new MessageBus with a specified number of workers
// and queue size.
func NewMessageBus(workers int, queueSize int) *MessageBus {
	return &MessageBus{
		subscribers: make(map[string][]MessageHandler),
		queue:       make(chan *Message, queueSize),
		workers:     workers,
		connected:   true,
	}
}

// Start initiates the worker pool, launching goroutines to process messages from the queue.
func (mb *MessageBus) Start(ctx context.Context) {
	for i := 0; i < mb.workers; i++ {
		go mb.worker(ctx, i)
	}
}

// worker is a goroutine that continuously pulls messages from the queue and dispatches them.
func (mb *MessageBus) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			// Context cancelled, worker should exit.
			return
		case msg, ok := <-mb.queue:
			if !ok {
				// Channel closed, worker should exit.
				return
			}
			mb.dispatch(msg)
		}
	}
}

// dispatch sends a message to all registered handlers for its specific type.
// Each handler is invoked in a new goroutine to avoid blocking the dispatcher.
func (mb *MessageBus) dispatch(msg *Message) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	handlers := mb.subscribers[msg.Type]
	for _, handler := range handlers {
		go handler(context.Background(), msg) // Execute handler asynchronously.
	}
}

// Subscribe registers a MessageHandler for a specific message type.
// Multiple handlers can be registered for the same message type.
func (mb *MessageBus) Subscribe(msgType string, handler MessageHandler) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.subscribers[msgType] = append(mb.subscribers[msgType], handler)
}

// Unsubscribe removes a specific MessageHandler for a given message type.
func (mb *MessageBus) Unsubscribe(msgType string, handler MessageHandler) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	var newHandlers []MessageHandler
	for _, h := range mb.subscribers[msgType] {
		// Compare handlers by their memory address to ensure correct removal.
		if fmt.Sprintf("%p", h) != fmt.Sprintf("%p", handler) {
			newHandlers = append(newHandlers, h)
		}
	}
	mb.subscribers[msgType] = newHandlers
}

// Publish adds a message to the internal queue for processing by workers.
// Returns an error if the message bus is not connected or the queue is full.
func (mb *MessageBus) Publish(msg *Message) error {
	if !mb.connected {
		return ErrNotConnected
	}

	select {
	case mb.queue <- msg: // Attempt to send message to the queue.
		return nil
	default:
		// If the queue is full, return an error.
		return ErrPublishFailed
	}
}

// Close shuts down the MessageBus, stopping all workers and closing the message queue.
func (mb *MessageBus) Close() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.connected = false
	close(mb.queue) // Close the channel to signal workers to exit.

	return nil
}

// -----------------------------------------------------------------------------
// Message Serialization/Deserialization
// -----------------------------------------------------------------------------

// MarshalMessage serializes a Message struct into a JSON byte array.
func MarshalMessage(msg *Message) ([]byte, error) {
	return json.Marshal(msg)
}

// UnmarshalMessage deserializes a JSON byte array into a Message struct.
func UnmarshalMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}
