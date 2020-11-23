package events

import (
	"crypto/tls"
	"github.com/omecodes/omestore/common"
)

type Config struct {
	Address   string
	Table     string
	DBUri     string
	validator common.CredentialsValidator
	TlsConfig *tls.Config
}
