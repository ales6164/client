package client

import (
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"time"
)

type Session struct {
	CreatedAt time.Time      `json:"createdAt"`
	ExpiresAt time.Time      `json:"expiresAt"`
	IsBlocked bool           `json:"-"`
	Scopes    []string       `json:"-"`
	Identity  *datastore.Key `json:"-"`
	User      *datastore.Key `json:"-"`
}

type Claims struct {
	*jwt.StandardClaims
}

type RoleProvider interface {
	Roles() map[string][]string
}

const SessionKind = "_clientSession"

func NewSession(ctx context.Context, identity *Identity, roleProvider RoleProvider, scopes ...string) (session *Session, token *jwt.Token, err error) {
	var tokenScopes []string
	var scopeMap = map[string]interface{}{}
	for _, roles := range roleProvider.Roles() {
		for _, scope := range roles {
			scopeMap[scope] = true
		}
	}
	for _, scope := range scopes {
		scopeMap[scope] = true
	}
	for scope := range scopeMap {
		tokenScopes = append(tokenScopes, scope)
	}

	createdAt := time.Now()
	expiresAt := createdAt.Add(time.Hour * time.Duration(72))
	session = &Session{
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		Identity:  identity.Key,
		User:      identity.UserKey,
		Scopes:    tokenScopes,
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
