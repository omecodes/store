package client

import (
	"encoding/base64"
	"fmt"
	"github.com/omecodes/store/common"
)

type Authentication interface {
	HeaderKey() string
	HeaderValue() string
}

type userBasicAuthentication struct {
	username string
	password string
}

func (b *userBasicAuthentication) HeaderKey() string {
	return common.HttpHeaderUserAuthorization
}

func (b *userBasicAuthentication) HeaderValue() string {
	basic := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", b.username, b.password)))
	return fmt.Sprintf("Basic %s", basic)
}

type bearerTokenAuthentication struct {
	token string
}

func (b *bearerTokenAuthentication) HeaderKey() string {
	return common.HttpHeaderUserAuthorization
}

func (b *bearerTokenAuthentication) HeaderValue() string {
	return fmt.Sprintf("Bearer %s", b.token)
}

type appAuthentication struct {
	key    string
	secret string
}

func (a *appAuthentication) HeaderKey() string {
	return common.HttpHeaderAppAuthorization
}

func (a *appAuthentication) HeaderValue() string {
	basic := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", a.key, a.secret)))
	return fmt.Sprintf("Basic %s", basic)
}
