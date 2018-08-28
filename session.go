package client

import (
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"time"
)

type Session struct {
	CreatedAt time.Time
	ExpiresAt time.Time
	IsBlocked bool
	Scopes    []string
	Identity  *datastore.Key
	User      *datastore.Key
}

const SessionKind = "_clientSession"

func NewSession(ctx context.Context, identity *Identity, scopes ...string) (session *Session, token *jwt.Token, err error) {
	createdAt := time.Now()
	expiresAt := createdAt.Add(time.Hour * time.Duration(72))
	session = &Session{
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		Identity:  identity.Key,
		User:      identity.UserKey,
		Scopes:    scopes,
	}
	key := datastore.NewIncompleteKey(ctx, SessionKind, nil)
	key, err = datastore.Put(ctx, key, session)
	if err != nil {
		return session, token, err
	}

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Audience:  "all",
		Id:        key.Encode(),
		ExpiresAt: expiresAt.Unix(),
		IssuedAt:  createdAt.Unix(),
		Issuer:    "goapp",
		NotBefore: createdAt.Add(-time.Minute).Unix(),
		Subject:   identity.UserKey.Encode(),
	})

	return session, token, err
}

func (s *Session) GetUser(ctx context.Context) (user *User, err error) {
	return getUser(ctx, s.User)
}

func (s *Session) GetIdentity(ctx context.Context) (identity *Identity, err error) {
	return getIdentity(ctx, s.Identity)
}