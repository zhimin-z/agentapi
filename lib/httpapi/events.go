package httpapi

import (
	"fmt"
	"sync"
	"time"

	st "github.com/coder/agentapi/lib/screentracker"
)

type EventType string

const (
	EventTypeMessageUpdate EventType = "message_update"
	EventTypeStatusChange  EventType = "status_change"
)

type AgentStatus string

const (
	AgentStatusStable  AgentStatus = "stable"
	AgentStatusRunning AgentStatus = "running"
)

type MessageUpdateBody struct {
	Id      int                 `json:"id"`
	Role    st.ConversationRole `json:"role"`
	Message string              `json:"message"`
	Time    time.Time           `json:"time"`
}

type StatusChangeBody struct {
	Status AgentStatus `json:"status"`
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
	status              AgentStatus
	chans               map[int]chan Event
	chanEventIdx        map[int]int
	chanIdx             int
	subscriptionBufSize int
}

func convertStatus(status st.ConversationStatus) AgentStatus {
	switch status {
	case st.ConversationStatusInitializing:
		return AgentStatusRunning
	case st.ConversationStatusStable:
		return AgentStatusStable
	case st.ConversationStatusChanging:
		return AgentStatusRunning
	default:
		panic(fmt.Sprintf("unknown conversation status: %s", status))
	}
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
		status:              AgentStatusRunning,
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
			e.unsubscribeInner(chanId)
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

	newAgentStatus := convertStatus(newStatus)
	if e.status == newAgentStatus {
		return
	}

	e.notifyChannels(EventTypeStatusChange, StatusChangeBody{Status: newAgentStatus})
	e.status = newAgentStatus
}

// Assumes the caller holds the lock.
func (e *EventEmitter) currentStateAsEvents() []Event {
	events := make([]Event, 0, len(e.messages)+1)
	for i, msg := range e.messages {
		events = append(events, Event{
			Id:      i,
			Type:    EventTypeMessageUpdate,
			Payload: MessageUpdateBody{Id: msg.Id, Role: msg.Role, Message: msg.Message, Time: msg.Time},
		})
	}
	events = append(events, Event{
		Id:      len(e.messages),
		Type:    EventTypeStatusChange,
		Payload: StatusChangeBody{Status: e.status},
	})
	return events
}

// Subscribe returns:
// - a subscription ID that can be used to unsubscribe.
// - a channel for receiving events.
// - a list of events that allow to recreate the state of the conversation right before the subscription was created.
func (e *EventEmitter) Subscribe() (int, <-chan Event, []Event) {
	e.mu.Lock()
	defer e.mu.Unlock()
	stateEvents := e.currentStateAsEvents()

	// Once a channel becomes full, it will be closed.
	ch := make(chan Event, e.subscriptionBufSize)
	e.chans[e.chanIdx] = ch
	e.chanEventIdx[e.chanIdx] = len(stateEvents)
	e.chanIdx++
	return e.chanIdx - 1, ch, stateEvents
}

// Assumes the caller holds the lock.
func (e *EventEmitter) unsubscribeInner(chanId int) {
	close(e.chans[chanId])
	delete(e.chans, chanId)
	delete(e.chanEventIdx, chanId)
}

func (e *EventEmitter) Unsubscribe(chanId int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.unsubscribeInner(chanId)
}
