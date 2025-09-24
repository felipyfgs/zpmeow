package ports

import (
	"context"

	"zpmeow/internal/domain/common"
)

type EventPublisher interface {
	Publish(ctx context.Context, event common.DomainEvent) error

	PublishBatch(ctx context.Context, events []common.DomainEvent) error
}

type EventHandler interface {
	Handle(ctx context.Context, event common.DomainEvent) error

	CanHandle(eventType string) bool
}

type EventBus interface {
	Subscribe(eventType string, handler EventHandler) error

	Unsubscribe(eventType string, handler EventHandler) error

	Publish(ctx context.Context, event common.DomainEvent) error
}
