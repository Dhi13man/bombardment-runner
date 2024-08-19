package models_dto_clients

import (
	"time"

	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
)

type ClientContext struct {
	Channel models_enums.ClientChannel `json:"channel"`
	// Maximum time a dial will wait for a connect to complete.
	DialTimeout time.Duration `json:"dialTimeout,omitempty"`
	// Maximum time a connection will be kept alive.
	DialKeepAlive time.Duration `json:"dialKeepAlive,omitempty"`
	// Maximum time waiting to perform a TLS handshake.
	TlsHandshakeTimeout time.Duration `json:"tlsHandshakeTimeout,omitempty"`
	// Maximum time waiting to read the response headers.
	ResponseHeaderTimeout time.Duration `json:"responseHeaderTimeout,omitempty"`
	// Maximum time waiting for a server's first response headers after fully writing the request headers.
	ExpectContinueTimeout time.Duration `json:"expectContinueTimeout,omitempty"`
}
