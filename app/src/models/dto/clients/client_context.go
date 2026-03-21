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

	// gRPC proto file support: paths to .proto files for protobuf encoding.
	// When set, the gRPC client sends standard protobuf instead of JSON codec.
	ProtoFiles       []string `json:"proto_files,omitempty"`
	ProtoImportPaths []string `json:"proto_import_paths,omitempty"`

	// gRPC connection tuning
	MaxRecvMsgSize   int           `json:"max_recv_msg_size,omitempty"`   // bytes, 0 = default (4MB)
	MaxSendMsgSize   int           `json:"max_send_msg_size,omitempty"`   // bytes, 0 = default (4MB)
	KeepaliveTime    time.Duration `json:"keepalive_time,omitempty"`      // nanoseconds, 0 = disabled
	KeepaliveTimeout time.Duration `json:"keepalive_timeout,omitempty"`   // nanoseconds, 0 = default (20s)
}
