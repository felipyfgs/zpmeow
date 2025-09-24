package common

import (
	"time"
)

type DomainEvent interface {
	EventID() string

	EventType() string

	AggregateID() string

	OccurredAt() time.Time

	EventData() interface{}
}

type BaseDomainEvent struct {
	eventID     string
	eventType   string
	aggregateID string
	occurredAt  time.Time
	data        interface{}
}

func NewBaseDomainEvent(eventType, aggregateID string, data interface{}) BaseDomainEvent {
	return BaseDomainEvent{
		eventID:     GenerateID().Value(),
		eventType:   eventType,
		aggregateID: aggregateID,
		occurredAt:  time.Now(),
		data:        data,
	}
}

func (e BaseDomainEvent) EventID() string {
	return e.eventID
}

func (e BaseDomainEvent) EventType() string {
	return e.eventType
}

func (e BaseDomainEvent) AggregateID() string {
	return e.aggregateID
}

func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e BaseDomainEvent) EventData() interface{} {
	return e.data
}

type AggregateRoot struct {
	id     ID
	events []DomainEvent
}

func NewAggregateRoot(id ID) AggregateRoot {
	return AggregateRoot{
		id:     id,
		events: make([]DomainEvent, 0),
	}
}

func (ar *AggregateRoot) ID() ID {
	return ar.id
}

func (ar *AggregateRoot) AddEvent(event DomainEvent) {
	ar.events = append(ar.events, event)
}

func (ar *AggregateRoot) GetEvents() []DomainEvent {
	return ar.events
}

func (ar *AggregateRoot) ClearEvents() {
	ar.events = make([]DomainEvent, 0)
}

func (ar *AggregateRoot) HasEvents() bool {
	return len(ar.events) > 0
}
