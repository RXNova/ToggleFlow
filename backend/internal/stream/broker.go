package stream

import "sync"

// Event is what gets broadcast to all subscribers when a flag changes.
type Event struct {
	ProjectID int64  `json:"project_id"`
	EnvKey    string `json:"env_key"`
	FlagKey   string `json:"flag_key"`
	Action    string `json:"action"` // "updated" | "deleted"
}

// Broker is an in-process pub/sub hub. Think of it as Node's EventEmitter
// but designed for concurrent goroutines — each subscriber gets its own channel
// and the broker fans out to all of them under a mutex.
type Broker struct {
	mu   sync.RWMutex
	subs map[chan Event]struct{}
}

func New() *Broker {
	return &Broker{subs: make(map[chan Event]struct{})}
}

// Subscribe returns a channel that will receive events. The caller must call
// Unsubscribe with the same channel when done (e.g. when the HTTP connection closes).
func (b *Broker) Subscribe() chan Event {
	ch := make(chan Event, 16)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broker) Unsubscribe(ch chan Event) {
	b.mu.Lock()
	delete(b.subs, ch)
	b.mu.Unlock()
	close(ch)
}

// Publish fans out an event to every active subscriber.
// Non-blocking send: if a subscriber's buffer is full we skip it rather than block.
func (b *Broker) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs {
		select {
		case ch <- e:
		default:
		}
	}
}
