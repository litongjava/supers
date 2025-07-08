package events

// EventType represents a type of event emitted by superd.
type EventType string

const (
	EventProcessExited    EventType = "process.exited"
	EventProcessRestarted EventType = "process.restarted"
)

// Event carries information about a process event.
type Event struct {
	Name     string    `json:"name"`
	Type     EventType `json:"type"`
	ExitCode int       `json:"exit_code"`
}

// Handler defines how to consume an Event.
type Handler interface {
	Handle(e Event)
}

var handlers []Handler

// Register adds a new Handler.
func Register(h Handler) {
	handlers = append(handlers, h)
}

// Emit dispatches the Event to all registered handlers.
func Emit(e Event) {
	for _, h := range handlers {
		go h.Handle(e)
	}
}
