package oidc

import (
	"errors"
	"sync"
)

type Provider interface {
}

type OIDCStore struct {
	sync.RWMutex

	clients map[string]Provider
}

func NewOIDCStore() *OIDCStore {
	return &OIDCStore{
		clients: make(map[string]Provider),
	}
}

func (s *OIDCStore) AddProvider(name string, provider Provider) {
	s.Lock()
	defer s.Unlock()

	s.clients[name] = provider
}

func (s *OIDCStore) GetProvider(name string) (Provider, error) {
	s.RLock()
	defer s.RUnlock()

	provider, ok := s.clients[name]
	if !ok {
		return nil, errors.New("invalid name for provider")
	}

	return provider, nil
}
