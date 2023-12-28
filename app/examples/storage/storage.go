package storage

import "context"

type Storage interface {
	Save(ctx context.Context, id string) error
}

type storage struct{}

func NewAzureStorage() Storage {
	return &storage{}
}

func (s *storage) Save(ctx context.Context, id string) error {
	return nil
}
