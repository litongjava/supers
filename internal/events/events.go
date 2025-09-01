package events

// EventType represents a type of event emitted by superd.
type EventType string

const (
  EventProcessStarted     EventType = "process.started"
  EventProcessStartFailed EventType = "process.start_failed"
  EventProcessExited      EventType = "process.exited"
  EventProcessRestarted   EventType = "process.restarted"
)

// Event carries information about a process event.
type Event struct {
  Name     string    `json:"name"`
  Type     EventType `json:"type"`
  ExitCode int       `json:"exit_code,omitempty"`
  PID      int       `json:"pid,omitempty"`
  Error    string    `json:"error,omitempty"`
}

// Handler defines how to consume an Event.
type Handler interface {
  Handle(e Event)
}
