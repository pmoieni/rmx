package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"time"

	"aidanwoods.dev/go-paseto"
)

var (
	privKey paseto.V4AsymmetricSecretKey
	pubKey  paseto.V4AsymmetricPublicKey
	parser  = paseto.NewParser()
)

func init() {
	// TODO: static keys
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	privKey, err = paseto.NewV4AsymmetricSecretKeyFromEd25519(priv)
	if err != nil {
		panic(err)
	}

	pubKey, err = paseto.NewV4AsymmetricPublicKeyFromEd25519(pub)
	if err != nil {
		panic(err)
	}
}

func New(userID string, email string, exp time.Duration) (string, error) {
	now := time.Now().UTC()

	token := paseto.NewToken()
	token.SetIssuedAt(now)
	token.SetNotBefore(now)
	token.SetExpiration(now.Add(exp))
	token.SetSubject(userID)
	token.SetString("email", email)

	return token.V4Sign(privKey, nil), nil
}

func Parse(token string) (*paseto.Token, error) {
	return parser.ParseV4Public(pubKey, token, nil)
}
