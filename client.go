package client

import (
	"github.com/dgrijalva/jwt-go"
	gContext "github.com/gorilla/context"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"net/http"
	"reflect"
	"time"
)

type Client struct {
	context         context.Context `datastore:"-"`
	HttpRequest     *http.Request   `datastore:"-"`
	key             *datastore.Key  `datastore:"-"`
	Time            time.Time
	Device          Device
	URL             string
	Method          string
	LogList         []string
	SessionKey      *datastore.Key `datastore:"ClientSession"`
	Session         Session        `datastore:"-"`
	IsAuthenticated bool
	IsBlocked       bool
	IsExpired       bool
	IsPublic        bool // true if accessed via an unprotected route
}

const ClientRequestKind = "_clientRequest"

var (
	LogAuthHeader            = "authorizationHeaderPresent"
	LogNoAuthHeader          = "noAuthorizationHeader"
	LogInvalidToken          = "invalidToken"
	LogInvalidClaims         = "invalidClaims"
	LogErrDecodingSessionKey = "errDecodingSessionKey"
	LogErrFetchingSession    = "errFetchingSession"
	LogUserAuthenticated     = "userAuthenticated"
	LogSessionBlocked        = "sessionBlocked"
	LogSessionExpired        = "sessionExpired"
)

func New(ctx context.Context, r *http.Request) *Client {
	var err error
	c := &Client{
		context:     ctx,
		HttpRequest: r,
		Time:        time.Now(),
		URL:         r.URL.String(),
		Method:      r.Method,
		Device:      GetDevice(r),
	}

	tkn := gContext.Get(r, "auth")
	if tkn != nil {
		c.Log(LogAuthHeader)
		if token, ok := tkn.(*jwt.Token); ok && token.Valid {
			if claims, ok := token.Claims.(*jwt.StandardClaims); ok {
				if c.SessionKey, err = datastore.DecodeKey(claims.Id); err == nil {
					if err = datastore.Get(c.context, c.SessionKey, &c.Session); err == nil {
						c.IsExpired = c.Session.ExpiresAt.Before(c.Time)
						c.IsBlocked = c.Session.IsBlocked
						c.IsAuthenticated = !c.IsExpired && !c.IsBlocked
						if c.IsAuthenticated {
							c.Log(LogUserAuthenticated)
						} else if c.IsBlocked {
							c.Log(LogSessionBlocked)
						} else {
							c.Log(LogSessionExpired)
						}
					} else {
						c.Log(LogErrFetchingSession)
					}
				} else {
					c.Log(LogErrDecodingSessionKey)
				}
			} else {
				c.Log(LogInvalidClaims)
				c.Log(reflect.TypeOf(token.Claims).String())
			}
		} else {
			c.Log(LogInvalidToken)
		}
	} else {
		c.Log(LogNoAuthHeader)
		c.IsPublic = true
	}

	c.key = datastore.NewIncompleteKey(c.context, ClientRequestKind, nil)
	c.key, err = datastore.Put(c.context, c.key, c)
	return c
}

func (c *Client) Log(msg string) {
	c.LogList = append(c.LogList, msg)
	if c.key != nil {
		datastore.Put(c.context, c.key, c)
	}
	log.Infof(c.context, "logged: %s", msg)
}
