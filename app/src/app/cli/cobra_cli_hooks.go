package appCli

import (
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	"github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	"github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	"github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	"github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
	"go.uber.org/zap"
)

const (
	RunModeGroupId string = "run-mode"

	BindAddrLongKey             string = "bind-addr"
	BindAddrShortKey            string = "b"
	ClientContextLongKey        string = "client-context"
	ClientContextShortKey       string = "C"
	DriverContextLongKey        string = "driver-context"
	DriverContextShortKey       string = "D"
	LoadBalancerContextLongKey  string = "load-balancer-context"
	LoadBalancerContextShortKey string = "L"
	ParserContextLongKey        string = "parser-context"
	ParserContextShortKey       string = "P"
	PortLongKey                 string = "port"
	PortShortKey                string = "p"
	TransformerContextLongKey   string = "transformer-context"
	TransformerContextShortKey  string = "T"

	ClientContextExampleJSON       string = `'{"channel":"REST","dial_keep_alive":10000000000,"dial_timeout":5000000000,"tls_handshake_timeout":5000000000,"response_header_timeout":5000000000,"expect_continue_timeout":500000,"request_timeout":30000000000,"insecure_skip_verify":false}'`
	DriverContextExampleJSON       string = `'{"batch_size":100,"should_store_responses":false,"responses_storage_path":"./responses"}'`
	LoadBalancerContextExampleJSON string = `'{"strategy":"ROUND_ROBIN","urls":["https://api.example.com","https://api-backup.example.com"]}'`
	ParserContextExampleJSON       string = `'{"strategy":"CSV","file_path":"./data/records.csv"}'`
	TransformerContextExampleJSON  string = `'{"strategy":"JSONATA","method_expression":"\"POST\"","endpoint_expression":"\"/api/v1/\" & resource","headers_expression":"{ \"Content-Type\": \"application/json\", \"X-Request-ID\": request_id }","body_expression":"{ \"id\": $number(id), \"timestamp\": $millis() }"}'`

	DefaultServerBindAddr string = "127.0.0.1"
	DefaultServerPort     int    = 8080
)

type cobraCliHooks struct {
	rootCmd *cobra.Command
}

func NewCobraCliHooks() CliHook {
	rootCmd := &cobra.Command{
		Use:   "bombardment",
		Short: "Run Bombardment in CLI or Server mode",
		Long: heredoc.Doc(
			`Bombardment is a lightweight automation tool intended to pick up data, transform it using a set of rules and then send it to a target system. 
			It is designed to perform small repetitive migrations of data from one system to another. 
			
			Bombardment supports concurrent processing of data, client-side load balancing strategies, and is designed to be extensible and reusable.

			Modes:
			cli     Run Bombardment in CLI mode. This mode reads data from a file, transforms it, and sends it to the server in batches.
			server  Run Bombardment in Server mode. This mode starts a server that listens for incoming data and sends it to the server in batches.`,
		),
		Example: heredoc.Doc(
			`# Run Bombardment in CLI mode
			bombardment cli --help

			# Run Bombardment in Server mode
			bombardment server --help`,
		),
		Version: "v0.0.1",
	}
	rootCmd.AddGroup(&cobra.Group{ID: RunModeGroupId, Title: "Run Mode"})
	return &cobraCliHooks{rootCmd: rootCmd}
}

func (c *cobraCliHooks) AttachCliRunCommand(
	runCliCallback func(
		clientContext modelsDtoClients.ClientContext,
		driverContext modelsDtoDriver.DriverContext,
		loadBalancerContext modelsDtoLoadBalancing.LoadBalancerContext,
		parserContext modelsDtoParsing.ParserContext,
		transformerContext modelsDtoTransforming.TransformerContext,
	) error,
) CliHook {
	var cliCommand = cobra.Command{
		Use:     fmt.Sprintf("cli {-C|--%s} {-D|--%s} {-L|--%s} {-P|--%s} {-T|--%s}", ClientContextLongKey, DriverContextLongKey, LoadBalancerContextLongKey, ParserContextLongKey, TransformerContextLongKey),
		Short:   "Run Bombardment in CLI mode",
		GroupID: RunModeGroupId,
		Long:    "Run Bombardment in CLI mode. This mode requires the user to provide the context for the Client, Driver, Load Balancer, Parser, and Transformer as JSON string flags.",
		Example: heredoc.Docf(
			`# Run Bombardment in CLI mode for a REST API, with a ROUND_ROBIN load balancer, CSV parser, and JSONATA transformer
			bombardment cli \
				--%s %s \
				--%s %s \
				--%s %s \
				--%s %s \
				--%s %s
			# or
			bombardment cli \
				-%s %s \
				-%s %s \
				-%s %s \
				-%s %s \
				-%s %s`,
			ClientContextLongKey,
			ClientContextExampleJSON,
			DriverContextLongKey,
			DriverContextExampleJSON,
			LoadBalancerContextLongKey,
			LoadBalancerContextExampleJSON,
			ParserContextLongKey,
			ParserContextExampleJSON,
			TransformerContextLongKey,
			TransformerContextExampleJSON,
			ClientContextShortKey,
			ClientContextExampleJSON,
			DriverContextShortKey,
			DriverContextExampleJSON,
			LoadBalancerContextShortKey,
			LoadBalancerContextExampleJSON,
			ParserContextShortKey,
			ParserContextExampleJSON,
			TransformerContextShortKey,
			TransformerContextExampleJSON,
		),
		Args: func(cmd *cobra.Command, args []string) error {
			// Validate that all required flags are provided
			requiredFlags := []string{
				ClientContextLongKey, DriverContextLongKey,
				LoadBalancerContextLongKey, ParserContextLongKey,
				TransformerContextLongKey,
			}
			for _, flag := range requiredFlags {
				val := cmd.Flag(flag).Value.String()
				if val == "" {
					return fmt.Errorf("--%s is required", flag)
				}
			}
			return nil
		},
		Version: "v0.0.1",
		Run: func(cmd *cobra.Command, args []string) {
			// Get the Flags
			clientContextCommand := cmd.Flag(ClientContextLongKey)
			driverContextCommand := cmd.Flag(DriverContextLongKey)
			loadBalancerContextCommand := cmd.Flag(LoadBalancerContextLongKey)
			parserContextCommand := cmd.Flag(ParserContextLongKey)
			transformerContextCommand := cmd.Flag(TransformerContextLongKey)

			// Parse the Client Context
			var clientContext modelsDtoClients.ClientContext
			err := json.Unmarshal([]byte(clientContextCommand.Value.String()), &clientContext)
			if err != nil {
				zap.L().Error("error parsing client context", zap.Error(err))
				return
			}

			// Parse the Driver Context
			var driverContext modelsDtoDriver.DriverContext
			err = json.Unmarshal([]byte(driverContextCommand.Value.String()), &driverContext)
			if err != nil {
				zap.L().Error("error parsing driver context", zap.Error(err))
				return
			}

			// Parse the Load Balancer Context
			var loadBalancerContext modelsDtoLoadBalancing.LoadBalancerContext
			err = json.Unmarshal([]byte(loadBalancerContextCommand.Value.String()), &loadBalancerContext)
			if err != nil {
				zap.L().Error("error parsing load balancer context", zap.Error(err))
				return
			}

			// Parse the Parser Context
			var parserContext modelsDtoParsing.ParserContext
			err = json.Unmarshal([]byte(parserContextCommand.Value.String()), &parserContext)
			if err != nil {
				zap.L().Error("error parsing parser context", zap.Error(err))
				return
			}

			// Parse the Transformer Context
			var transformerContext modelsDtoTransforming.TransformerContext
			err = json.Unmarshal([]byte(transformerContextCommand.Value.String()), &transformerContext)
			if err != nil {
				zap.L().Error("error parsing transformer context", zap.Error(err))
				return
			}

			// Run the Bombardment
			err = runCliCallback(
				clientContext,
				driverContext,
				loadBalancerContext,
				parserContext,
				transformerContext,
			)
			if err != nil {
				zap.L().Error("error running cli bombardment", zap.Error(err))
				return
			}
		},
	}
	cliCommand.Flags().StringP(
		ClientContextLongKey,
		ClientContextShortKey,
		"",
		heredoc.Docf(
			`The Context to use for the Client that will make the calls.
			Client Context is a JSON string that contains the following keys:
				- channel: The channel to use for the client. Possible values are {REST, GRPC}
				- dial_keep_alive: The number of nanoseconds for which to keep connections alive. Eg. 10000000000 (10 seconds)
				- dial_timeout: The number of nanoseconds for which to wait for a connection to complete. Eg. 5000000000 (5 seconds)
				- tls_handshake_timeout: The duration for which to wait for the TLS handshake to complete. Post the timeout, the connection will be closed. Eg. 5000000000 (5 seconds)
				- response_header_timeout: The duration for which to wait for the response headers. Eg. 5000000000 (5 seconds)
				- expect_continue_timeout: The duration for which to wait for a server's FIRST response headers after fully writing the request headers. Post the timeout, the request will be sent without the Expect: 100-continue header. Eg. 500000 (500 milliseconds)
				- request_timeout: Overall timeout for the entire request, from dial to response reading. Eg. 30000000000 (30 seconds)
				- insecure_skip_verify: Whether to skip TLS certificate verification. Eg. false
			Eg. %s`,
			ClientContextExampleJSON,
		),
	)
	cliCommand.Flags().StringP(
		DriverContextLongKey,
		DriverContextShortKey,
		"",
		heredoc.Docf(
			`The Context to use for the Driver that will orchestrate the Bombardment.
			Driver Context is a JSON string that contains the following keys:
				- batch_size: The number of records to send in a single batch. Eg. 1000
				- should_store_responses: A boolean flag to indicate if the responses should be stored. Default is false.
				- responses_storage_path: Optional path where response files will be stored if should_store_responses is true. Default is "./responses".
			Eg. %s`,
			DriverContextExampleJSON,
		),
	)
	cliCommand.Flags().StringP(
		LoadBalancerContextLongKey,
		LoadBalancerContextShortKey,
		"",
		heredoc.Docf(
			`The Context to use for Load Balancing requests across servers. All servers must support the same API contract. Eg. Different pods of the same service.
			Load Balancer Context is a JSON string that contains the following keys:
				- strategy: The strategy to use for the load balancer. Possible values are {ROUND_ROBIN}
				- urls: The list of URLs to use for the load balancer.
			Eg. %s`,
			LoadBalancerContextExampleJSON,
		),
	)
	cliCommand.Flags().StringP(
		ParserContextLongKey,
		ParserContextShortKey,
		"",
		heredoc.Docf(
			`The Context to use for Parsing the input data.
			Parser Context is a JSON string that contains the following keys:
				- strategy: The strategy to use for parsing the file. Possible values are {CSV, JSON, NDJSON, EXCEL, PARQUET}
				- file_path: The path to the file to parse.
				- on_error: Error handling behavior. "SKIP" (default) skips malformed records, "STOP" halts on first error.
				- options: Strategy-specific settings (JSON object).

			Examples:
				CSV:    -P '{"strategy":"CSV","file_path":"data.csv"}'
				TSV:    -P '{"strategy":"CSV","file_path":"data.tsv","options":{"delimiter":"\\t"}}'
				JSON:   -P '{"strategy":"JSON","file_path":"data.json"}'
				NDJSON: -P '{"strategy":"NDJSON","file_path":"events.jsonl"}'
				Excel:  -P '{"strategy":"EXCEL","file_path":"report.xlsx","options":{"sheet_name":"Sheet2"}}'
				Parquet: -P '{"strategy":"PARQUET","file_path":"export.parquet"}'
				Stop on error: add "on_error":"STOP" to any parser context
			Eg. %s`,
			ParserContextExampleJSON,
		),
	)
	cliCommand.Flags().StringP(
		TransformerContextLongKey,
		TransformerContextShortKey,
		"",
		heredoc.Docf(
			`The Context to use for Transforming the parsed data.
			Transformer Context is a JSON string that contains the following keys:
				- strategy: The strategy to use for transforming the data. Possible values are {JSONATA}
				- endpoint_expression: The expression to use for the endpoint of the request. Eg. "/insight/v1/event/ingest" (if the strategy is JSONATA)
				- headers_expression: The expression to use for the headers of the request. Eg. { "Content-Type": "application/json" } (if the strategy is JSONATA)
				- method_expression: The expression to use for the method of the request. Eg. "POST" (if the strategy is JSONATA)
				- body_expression: The expression to use for the body of the request. Eg. 
			Eg. %s`,
			TransformerContextExampleJSON,
		),
	)
	c.rootCmd.AddCommand(&cliCommand)
	return c
}

func (c *cobraCliHooks) AttachServerRunCommand(
	runServerCallback func(bindAddr string, port int),
) CliHook {
	var serverCommand = cobra.Command{
		Use:     "server",
		Short:   "Run Bombardment in Server mode",
		GroupID: RunModeGroupId,
		Long:    heredoc.Doc(`Run Bombardment in Server mode. This mode starts an HTTP server that listens on the specified address and port for incoming data and processes it in batches.`),
		Example: heredoc.Docf(
			`# Run server on default %s:%d
			bombardment server

			# Run server on custom address and port
			bombardment server 
				--%s %s 
				--%s %d
			# or
			bombardment server -%s %s -%s %d`,
			DefaultServerBindAddr,
			DefaultServerPort,
			BindAddrLongKey,
			DefaultServerBindAddr,
			PortLongKey,
			DefaultServerPort,
			BindAddrShortKey,
			DefaultServerBindAddr,
			PortShortKey,
			DefaultServerPort,
		),
		Args: func(cmd *cobra.Command, args []string) error {
			port, err := cmd.Flags().GetInt(PortLongKey)
			if err != nil {
				return err
			}
			if port < 1 || port > 65535 {
				return fmt.Errorf("invalid port: %d", port)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			bindAddr, _ := cmd.Flags().GetString(BindAddrLongKey)
			port, _ := cmd.Flags().GetInt(PortLongKey)

			runServerCallback(bindAddr, port)
		},
		Version: "v0.0.1",
	}
	// Define flags for server command
	serverCommand.Flags().StringP(
		BindAddrLongKey,
		"b",
		DefaultServerBindAddr,
		heredoc.Docf("Bind address for the server (default %s)", DefaultServerBindAddr),
	)
	serverCommand.Flags().IntP(
		PortLongKey,
		"p",
		DefaultServerPort,
		heredoc.Docf("Port for the server (default %d)", DefaultServerPort),
	)
	c.rootCmd.AddCommand(&serverCommand)
	return c
}

func (c *cobraCliHooks) Execute() error {
	err := c.rootCmd.Execute()
	if err != nil {
		return err
	}
	return nil
}
