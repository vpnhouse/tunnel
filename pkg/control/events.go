package control

const (
	EventRestart = iota + 1
	EventSetLogLevel
	EventCriticalError
)

type Event struct {
	// rename to Type
	EventType int
	Info      interface{}
}

type EventManager struct {
	ch chan Event
}

func NewEventManager() *EventManager {
	return &EventManager{
		ch: make(chan Event),
	}
}

func (m *EventManager) EmitEvent(event int) {
	m.ch <- Event{
		EventType: event,
	}
}

func (m *EventManager) EmitEventWithInfo(event int, info interface{}) {
	m.ch <- Event{
		EventType: event,
		Info:      info,
	}
}

// ? считаю эту функцию лишней
func (m *EventManager) NextEvent() *Event {
	event, ok := <-m.ch
	if !ok {
		return nil
	} else {
		return &event
	}
}

func (m *EventManager) EventChannel() chan Event {
	return m.ch
}
