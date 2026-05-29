package stream

import "sync"

// Event is what gets broadcast to all subscribers when a flag changes.
type Event struct {
	ProjectID int64  `json:"project_id"`
	EnvKey    string `json:"env_key"`
	FlagKey   string `json:"flag_key"`
	Action    string `json:"action"` // "updated" | "deleted"
}

// Broker is an in-process pub/sub hub sharded by project ID.
// Each project has its own subscriber set so a Publish for project A
// only iterates project A's channels — not every connected SSE client.
type Broker struct {
	mu   sync.RWMutex
	subs map[int64]map[chan Event]struct{}
}

func New() *Broker {
	return &Broker{subs: make(map[int64]map[chan Event]struct{})}
}

// Subscribe returns a channel scoped to projectID. The caller must call
// Unsubscribe with the same projectID and channel when done.
func (b *Broker) Subscribe(projectID int64) chan Event {
	ch := make(chan Event, 16)
	b.mu.Lock()
	if b.subs[projectID] == nil {
		b.subs[projectID] = make(map[chan Event]struct{})
	}
	b.subs[projectID][ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broker) Unsubscribe(projectID int64, ch chan Event) {
	b.mu.Lock()
	if subs, ok := b.subs[projectID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(b.subs, projectID)
		}
	}
	b.mu.Unlock()
	close(ch)
}

// Publish fans out an event only to subscribers of the same project.
// Non-blocking send: if a subscriber's buffer is full we skip it rather than block.
func (b *Broker) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs[e.ProjectID] {
		select {
		case ch <- e:
		default:
		}
	}
}

// PublishAll broadcasts an event to every connected subscriber across all projects.
func (b *Broker) PublishAll(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, subs := range b.subs {
		for ch := range subs {
			select {
			case ch <- e:
			default:
			}
		}
	}
}
