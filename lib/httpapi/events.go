package httpapi

import (
	"sync"
	"time"

	st "github.com/coder/openagent/lib/screentracker"
)

type EventType string

const (
	EventTypeMessageUpdate EventType = "message_update"
	EventTypeStatusChange  EventType = "status_change"
)

type MessageUpdateBody struct {
	Id      int
	Role    st.ConversationRole
	Message string
	Time    time.Time
}

type StatusChangeBody struct {
	Status st.ConversationStatus
}

type Event struct {
	// Id's are monotonically increasing integers within a single subscription.
	// They are not globally unique.
	Id      int
	Type    EventType
	Payload any
}

type EventEmitter struct {
	mu                  sync.Mutex
	messages            []st.ConversationMessage
	status              st.ConversationStatus
	chans               map[int]chan Event
	chanEventIdx        map[int]int
	chanIdx             int
	subscriptionBufSize int
}

// subscriptionBufSize is the size of the buffer for each subscription.
// Once the buffer is full, the channel will be closed.
// Listeners must actively drain the channel, so it's important to
// set this to a value that is large enough to handle the expected
// number of events.
func NewEventEmitter(subscriptionBufSize int) *EventEmitter {
	return &EventEmitter{
		mu:                  sync.Mutex{},
		messages:            make([]st.ConversationMessage, 0),
		status:              st.ConversationStatusInitializing,
		chans:               make(map[int]chan Event),
		chanEventIdx:        make(map[int]int),
		chanIdx:             0,
		subscriptionBufSize: subscriptionBufSize,
	}
}

// Assumes the caller holds the lock.
func (e *EventEmitter) notifyChannels(eventType EventType, payload any) {
	chanIds := make([]int, 0, len(e.chans))
	for chanId := range e.chans {
		chanIds = append(chanIds, chanId)
	}
	for _, chanId := range chanIds {
		ch := e.chans[chanId]
		event := Event{
			Id:      e.chanEventIdx[chanId],
			Type:    eventType,
			Payload: payload,
		}
		e.chanEventIdx[chanId]++

		select {
		case ch <- event:
		default:
			// If the channel is full, close it.
			// Listeners must actively drain the channel.
			close(ch)
			delete(e.chans, chanId)
			delete(e.chanEventIdx, chanId)
		}
	}
}

// Assumes that only the last message can change or new messages can be added.
// If a new message is injected between existing messages (identified by Id), the behavior is undefined.
func (e *EventEmitter) UpdateMessagesAndEmitChanges(newMessages []st.ConversationMessage) {
	e.mu.Lock()
	defer e.mu.Unlock()

	maxLength := max(len(e.messages), len(newMessages))
	for i := range maxLength {
		var oldMsg st.ConversationMessage
		var newMsg st.ConversationMessage
		if i < len(e.messages) {
			oldMsg = e.messages[i]
		}
		if i < len(newMessages) {
			newMsg = newMessages[i]
		}
		if oldMsg != newMsg {
			e.notifyChannels(EventTypeMessageUpdate, MessageUpdateBody{
				Id:      newMessages[i].Id,
				Role:    newMessages[i].Role,
				Message: newMessages[i].Message,
				Time:    newMessages[i].Time,
			})
		}
	}

	e.messages = newMessages
}

func (e *EventEmitter) UpdateStatusAndEmitChanges(newStatus st.ConversationStatus) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status == newStatus {
		return
	}

	e.notifyChannels(EventTypeStatusChange, StatusChangeBody{Status: newStatus})
	e.status = newStatus
}

// Assumes the caller holds the lock.
func (e *EventEmitter) currentStateAsEvents() []Event {
	events := make([]Event, 0, len(e.messages)+1)
	for i, msg := range e.messages {
		events[i] = Event{
			Id:      i,
			Type:    EventTypeMessageUpdate,
			Payload: MessageUpdateBody{Id: msg.Id, Role: msg.Role, Message: msg.Message, Time: msg.Time},
		}
	}
	events = append(events, Event{
		Id:      len(e.messages),
		Type:    EventTypeStatusChange,
		Payload: StatusChangeBody{Status: e.status},
	})
	return events
}

// Subscribe returns a channel for receiving events. It also returns a list of events that allow to
// recreate the state of the conversation right before the subscription was created.
func (e *EventEmitter) Subscribe() (<-chan Event, []Event) {
	e.mu.Lock()
	defer e.mu.Unlock()
	stateEvents := e.currentStateAsEvents()

	// Once a channel becomes full, it will be closed.
	ch := make(chan Event, e.subscriptionBufSize)
	e.chans[e.chanIdx] = ch
	e.chanEventIdx[e.chanIdx] = len(stateEvents)
	e.chanIdx++
	return ch, stateEvents
}
