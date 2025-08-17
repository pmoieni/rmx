package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pmoieni/rmx/internal/lib"
	"github.com/pmoieni/rmx/internal/oauth"
	"golang.org/x/oauth2"
)

var (
	stateLength   uint = 16
	defaultScopes      = []string{"profile", "email"}
	issuer             = "https://accounts.google.com"
)

type Provider struct {
	*oauth2.Config

	oidc     *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

func NewOIDC(ctx context.Context, clientID, clientSecret, redirectURL string) *Provider {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		log.Fatal(err)
	}

	return &Provider{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       append([]string{oidc.ScopeOpenID}, defaultScopes...),
		},
		oidc:     provider,
		verifier: provider.Verifier(&oidc.Config{ClientID: clientID}),
	}
}

func (p *Provider) HandleAuthorizationRequest(w http.ResponseWriter, r *http.Request) {
	state, err := lib.RandomString(stateLength)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	nonce, err := lib.RandomString(stateLength)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	setCallbackCookie(w, r, "state", state)
	setCallbackCookie(w, r, "nonce", nonce)

	http.Redirect(w, r, p.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}

// based on Google's spec https://developers.google.com/identity/openid-connect/openid-connect#obtainuserinfo
func (p *Provider) GetCallbackResult(r *http.Request) (*oauth.CallbackResult, error) {
	oauthToken, err := p.getOAuthToken(r)
	if err != nil {
		return nil, err
	}

	idToken, err := p.getIDToken(r, oauthToken)
	if err != nil {
		return nil, err
	}

	var claims struct {
		Email         string           `json:"email"`
		EmailVerified lib.StringAsBool `json:"email_verified"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	tbs, err := json.Marshal(oauthToken)
	if err != nil {
		return nil, err
	}

	return &oauth.CallbackResult{
		Issuer:        idToken.Issuer,
		UserID:        idToken.Subject,
		Email:         claims.Email,
		EmailVerified: bool(claims.EmailVerified),
		Token:         string(tbs),
	}, nil
}

func (p *Provider) getOAuthToken(r *http.Request) (*oauth2.Token, error) {
	state, err := r.Cookie("state")
	if err != nil {
		return nil, errors.New("state not found")
	}

	if r.URL.Query().Get("state") != state.Value {
		return nil, errors.New("state didn't match")
	}

	return p.Exchange(r.Context(), r.URL.Query().Get("code"))
}

func (p *Provider) getIDToken(r *http.Request, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	fmt.Println(rawIDToken)
	idToken, err := p.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("failed to verify ID Token")
	}

	nonce, err := r.Cookie("nonce")
	if err != nil {
		return nil, errors.New("nonce not found")
	}

	if idToken.Nonce != nonce.Value {
		return nil, errors.New("nonce didn't match")
	}

	return idToken, nil
}

func (p *Provider) VerifyAccessToken(ctx context.Context, token string) error {
	var oauthToken *oauth2.Token
	if err := json.Unmarshal([]byte(token), &oauthToken); err != nil {
		return err
	}

	if _, err := p.oidc.UserInfo(ctx, oauth2.StaticTokenSource(oauthToken)); err != nil {
		return err
	}

	return nil
}
