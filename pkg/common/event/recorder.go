package event

type Recorder struct {
	Events []Event
}

func (r *Recorder) AddEvent(eventType, reason, message string) {
	r.Events = append(r.Events, Event{
		EventType: eventType,
		Reason:    reason,
		Message:   message,
	})
}

func NewEventRecorder() *Recorder {
	return &Recorder{Events: []Event{}}
}
