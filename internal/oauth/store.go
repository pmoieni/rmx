package oauth

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

type Provider interface {
	HandleAuthorizationRequest(http.ResponseWriter, *http.Request)
	GetCallbackResult(*http.Request) (*CallbackResult, error)
	VerifyAccessToken(context.Context, string) error
}

type CallbackResult struct {
	Issuer        string
	UserID        string
	Email         string
	EmailVerified bool
	Token         string
}

type ClientStore struct {
	sync.RWMutex

	clients map[string]Provider
}

func NewClientStore() *ClientStore {
	return &ClientStore{
		clients: make(map[string]Provider),
	}
}

func (cs *ClientStore) AddProvider(name string, provider Provider) {
	cs.Lock()
	defer cs.Unlock()

	cs.clients[name] = provider
}

func (cs *ClientStore) GetProvider(name string) (Provider, error) {
	cs.RLock()
	defer cs.RUnlock()

	provider, ok := cs.clients[name]
	if !ok {
		return nil, errors.New("invalid name for provider")
	}

	return provider, nil
}
