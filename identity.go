package client

import (
	"github.com/ales6164/apis/errors"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type Identity struct {
	Key      *datastore.Key `datastore:"-" json:"-"`
	Id       string         `json:"id"`
	Provider string         `json:"provider"`
	UserKey  *datastore.Key `json:"userKey"`
	Secret   []byte         `datastore:",noindex" json:"-"`
}

var (
	ErrIdentityAlreadyExists = errors.New("identity already exists")
)

const (
	UserIdentityKindName = "_userIdentity"
)

func NewIdentity(ctx context.Context, ip string, userKey *datastore.Key, name string, secret []byte) (identity *Identity, err error) {
	idName := ip + "|" + name
	identity = &Identity{
		Id:       idName,
		Provider: ip,
		UserKey:  userKey,
		Secret:   secret,
	}
	identity.Key = datastore.NewKey(ctx, UserIdentityKindName, idName, 0, nil)
	var dst datastore.PropertyList
	err = datastore.Get(ctx, identity.Key, &dst)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			identity.Key, err = datastore.Put(ctx, identity.Key, identity)
			return identity, err
		}
		return identity, err
	}
	return identity, ErrIdentityAlreadyExists
}

func getIdentity(ctx context.Context, key *datastore.Key) (identity *Identity, err error) {
	identity = new(Identity)
	err = datastore.Get(ctx, key, identity)
	identity.Key = key
	return identity, err
}

func GetIdentity(ctx context.Context, ip string, name string) (identity *Identity, err error) {
	idName := ip + "|" + name
	key := datastore.NewKey(ctx, UserIdentityKindName, idName, 0, nil)
	return getIdentity(ctx, key)
}

func (i *Identity) GetUser(ctx context.Context) (user *User, err error) {
	return getUser(ctx, i.UserKey)
}
