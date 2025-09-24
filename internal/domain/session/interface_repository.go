package session

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, session *Session) error
	CreateWithGeneratedID(ctx context.Context, session *Session) (string, error)
	GetByID(ctx context.Context, id string) (*Session, error)
	GetByName(ctx context.Context, name string) (*Session, error)
	GetByApiKey(ctx context.Context, apiKey string) (*Session, error)
	GetByDeviceJID(ctx context.Context, deviceJID string) (*Session, error)
	GetAll(ctx context.Context) ([]*Session, error)
	List(ctx context.Context, limit, offset int, status string) ([]*Session, int, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, name string) (bool, error)

	GetActive(ctx context.Context) ([]*Session, error)
	GetInactive(ctx context.Context) ([]*Session, error)
}
