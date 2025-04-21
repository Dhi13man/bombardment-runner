package models_dto_clients

import (
	"time"

	models_enums "dhi13man.github.io/credit_card_bombardment/src/models/enums"
)

type ClientContext struct {
	Channel models_enums.ClientChannel `json:"channel"`
	// Maximum time a dial will wait for a connect to complete.
	DialTimeout time.Duration `json:"dial_timeout,omitempty"`
	// Maximum time a connection will be kept alive.
	DialKeepAlive time.Duration `json:"dial_keep_alive,omitempty"`
	// Maximum time waiting to perform a TLS handshake.
	TlsHandshakeTimeout time.Duration `json:"tls_handshake_timeout,omitempty"`
	// Maximum time waiting to read the response headers.
	ResponseHeaderTimeout time.Duration `json:"response_header_timeout,omitempty"`
	// Maximum time waiting for a server's first response headers after fully writing the request headers.
	ExpectContinueTimeout time.Duration `json:"expect_continue_timeout,omitempty"`
}
