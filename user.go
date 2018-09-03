package client

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"reflect"
	"strings"
	"time"
)

type User struct {
	Key       *datastore.Key         `json:"-"`
	Id        string                 `json:"id"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	Scopes    []string               `json:"scopes"`
	Profile   map[string]interface{} `json:"profile"`
}

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

const (
	UserKindName = "_user"
)

func NewUser(ctx context.Context, name string, profile map[string]interface{}, scopes ...string) (user *User, err error) {
	createdAt := time.Now()
	user = &User{
		Id:        name,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
		Scopes:    scopes,
		Profile:   profile,
	}
	user.Key = datastore.NewKey(ctx, UserKindName, name, 0, nil)
	var dst datastore.PropertyList
	err = datastore.Get(ctx, user.Key, &dst)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			user.Key, err = datastore.Put(ctx, user.Key, user)
			return user, err
		}
		return user, err
	}
	return user, ErrUserAlreadyExists
}

func getUser(ctx context.Context, key *datastore.Key) (user *User, err error) {
	user = new(User)
	err = datastore.Get(ctx, key, user)
	user.Key = key
	return user, err
}

func (u *User) Load(ps []datastore.Property) error {
	u.Profile = map[string]interface{}{}
	for _, p := range ps {
		switch p.Name {
		case "Id":
			u.Id = p.Value.(string)
		case "CreatedAt":
			u.CreatedAt = p.Value.(time.Time)
		case "UpdatedAt":
			u.UpdatedAt = p.Value.(time.Time)
		case "Scopes":
			u.Scopes = append(u.Scopes, p.Value.(string))
		default:
			// profile map
			var holder = u.Profile
			splitName := strings.Split(p.Name, ".")
			if len(splitName) > 1 {
				splitName = splitName[:len(splitName)-1]
				for _, name := range splitName {
					if _, ok := holder[name]; !ok {
						holder[name] = map[string]interface{}{}
					}
					holder = holder[name].(map[string]interface{})
				}
			}
			holder[splitName[len(splitName)-1]] = p.Value
		}
	}
	return nil
}

func (u *User) Save() ([]datastore.Property, error) {
	var ps = []datastore.Property{
		{Name: "Id", Value: u.Id},
		{Name: "CreatedAt", Value: u.CreatedAt},
		{Name: "UpdatedAt", Value: u.UpdatedAt},
	}
	for _, scope := range u.Scopes {
		ps = append(ps, datastore.Property{
			Multiple: true,
			Name:     "Scopes",
			Value:    scope,
		})
	}
	ps = append(ps, mapToDatastoreProperties(u.Profile)...)
	return ps, nil
}

func mapToDatastoreProperties(m map[string]interface{}) []datastore.Property {
	var ps []datastore.Property
	for name, value := range m {
		if isMap(value) {
			ps = append(ps, mapToDatastoreProperties(value.(map[string]interface{}))...)
		} else {
			rt := reflect.TypeOf(value)
			multiple := rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array
			ps = append(ps, datastore.Property{Name: name, Value: value, NoIndex: true, Multiple: multiple})
		}
	}
	return ps
}

func isMap(x interface{}) bool {
	t := fmt.Sprintf("%T", x)
	return strings.HasPrefix(t, "map[string]interface{}")
}
