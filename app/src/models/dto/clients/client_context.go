package modelsDtoClients

import (
	"time"

	modelsEnums "github.dhi13man.com/bombardment-runner/src/models/enums"
)

type ClientContext struct {
	Channel modelsEnums.ClientChannel `json:"channel"`
	// Maximum time a dial will wait for connect to complete.
	DialTimeout time.Duration `json:"dial_timeout,omitempty"`
	// Maximum time a connection will be kept alive.
	DialKeepAlive time.Duration `json:"dial_keep_alive,omitempty"`
	// Maximum time waiting to perform a TLS handshake.
	TlsHandshakeTimeout time.Duration `json:"tls_handshake_timeout,omitempty"`
	// Maximum time waiting to read the response headers.
	ResponseHeaderTimeout time.Duration `json:"response_header_timeout,omitempty"`
	// Maximum time waiting for a server's first response headers after fully writing the request headers.
	ExpectContinueTimeout time.Duration `json:"expect_continue_timeout,omitempty"`
	// Overall timeout for the entire request, from dial-to-response reading
	RequestTimeout time.Duration `json:"request_timeout,omitempty"`
	// Whether to skip TLS certificate verification
	InsecureSkipVerify bool `json:"insecure_skip_verify,omitempty"`
}
