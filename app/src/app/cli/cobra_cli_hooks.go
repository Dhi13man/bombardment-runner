package core_cli

import (
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	models_dto_clients "github.dhi13man.com/bombardment-runner/src/models/dto/clients"
	models_dto_driver "github.dhi13man.com/bombardment-runner/src/models/dto/driver"
	models_dto_load_balancing "github.dhi13man.com/bombardment-runner/src/models/dto/load_balancing"
	models_dto_parsing "github.dhi13man.com/bombardment-runner/src/models/dto/parsing"
	models_dto_transforming "github.dhi13man.com/bombardment-runner/src/models/dto/transforming"
)

var (
	RUN_MODE_GROUP_ID string = "run-mode"

	BIND_ADDR_LONG_KEY              string = "bind-addr"
	BIND_ADDR_SHORT_KEY             string = "b"
	CLIENT_CONTEXT_LONG_KEY         string = "client-context"
	CLIENT_CONTEXT_SHORT_KEY        string = "C"
	DRIVER_CONTEXT_LONG_KEY         string = "driver-context"
	DRIVER_CONTEXT_SHORT_KEY        string = "D"
	LOAD_BALANCER_CONTEXT_LONG_KEY  string = "load-balancer-context"
	LOAD_BALANCER_CONTEXT_SHORT_KEY string = "L"
	PARSER_CONTEXT_LONG_KEY         string = "parser-context"
	PARSER_CONTEXT_SHORT_KEY        string = "P"
	PORT_LONG_KEY                   string = "port"
	PORT_SHORT_KEY                  string = "p"
	TRANSFORMER_CONTEXT_LONG_KEY    string = "transformer-context"
	TRANSFORMER_CONTEXT_SHORT_KEY   string = "T"

	DEFAULT_SERVER_BIND_ADDR string = "127.0.0.1"
	DEFAULT_SERVER_PORT      int    = 8080
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

			Options:
			cli     Run Bombardment in CLI mode. This mode reads data from a file, transforms it, and sends it to the server in batches.
			server  Run Bombardment in Server mode. This mode starts a server that listens for incoming data and sends it to the server in batches.`,
		),
		Example: heredoc.Doc(
			`# Run Bombardment in CLI mode
			bombardment cli <flags>

			# Run Bombardment in Server mode
			bombardment server <flags>`,
		),
		Version: "v0.0.1",
	}
	rootCmd.AddGroup(&cobra.Group{ID: RUN_MODE_GROUP_ID, Title: "Run Mode"})
	return &cobraCliHooks{rootCmd: rootCmd}
}

func (c *cobraCliHooks) AttachCliRunCommand(
	runCliCallback func(
		clientContext models_dto_clients.ClientContext,
		driverContext models_dto_driver.DriverContext,
		loadBalancerContext models_dto_load_balancing.LoadBalancerContext,
		parserContext models_dto_parsing.ParserContext,
		transformerContext models_dto_transforming.TransformerContext,
	) error,
) CliHook {
	var cliCommand = cobra.Command{
		Use:     fmt.Sprintf("cli {-C|--%s} {-D|--%s} {-L|--%s} {-P|--%s} {-T|--%s}", CLIENT_CONTEXT_LONG_KEY, DRIVER_CONTEXT_LONG_KEY, LOAD_BALANCER_CONTEXT_LONG_KEY, PARSER_CONTEXT_LONG_KEY, TRANSFORMER_CONTEXT_LONG_KEY),
		Short:   "Run Bombardment in CLI mode",
		GroupID: RUN_MODE_GROUP_ID,
		Long:    "Run Bombardment in CLI mode. This mode requires the user to provide the context for the Client, Driver, Load Balancer, Parser, and Transformer as JSON string flags.",
		Example: heredoc.Docf(
			`# Run Bombardment in CLI mode for a REST API, with a ROUND_ROBIN load balancer, CSV parser, and JSONATA transformer
			bombardment cli \
				--%s "{\"channel\":\"REST\",\"dial_keep_alive\":10000000000,\"dial_timeout\":5000000000,\"expect_continue_timeout\":500000,\"response_header_timeout\":5000000000,\"tls_handshake_timeout\":5000000000}" \
				--%s "{\"batch_size\":1000,\"should_store_responses\":false}" \
				--%s "{\"strategy\":\"ROUND_ROBIN\",\"urls\":[\"http://api.bombardment.org\",\"http://mirror-1.bombardment.org\",\"http://mirror-2.bombardment.org\"]}" \
				--%s "{\"file_path\":\"./private/file_path.csv\",\"strategy\":\"CSV\"}" \
				--%s "{\"body_expression\":\"{\\n\\t\\t\\\"request_id\\\": \\\"bulk-create-\\\" & $number(row_id),\\n\\t\\t\\\"event_ts\\\": $millis(),\\n\\t\\\"user_account_id\\\": user_account_id,\\n\\t\\\"template_id\\\": \\\"4066f10464763823cc3e70c2ebd973fbd72cc5b1b450ccd31c0e87d9405e9dd6\\\",\\n\\t\\\"sms_date\\\": $millis(),\\n\\t\\\"insights\\\": $string({\\n\\t\\t\\\"billerName\\\": biller_name,\\n\\t\\\"last_four_dig_cc\\\": last_4_digits,\\n\\t\\\"mobile__number\\\": $floor($number(mobile_number))\\n\\t})\\n\\t}\",\"endpoint_expression\":\"\\\"/insight/v1/event/ingest\\\"\",\"headers_expression\":\"{ \\\"Content-Type\\\": \\\"application/json\\\" }\",\"method_expression\":\"\\\"POST\\\"\",\"strategy\":\"JSONATA\"}"
			# or
			bombardment cli \
				-%s "{\"channel\":\"REST\",\"dial_keep_alive\":10000000000,\"dial_timeout\":5000000000,\"expect_continue_timeout\":500000,\"response_header_timeout\":5000000000,\"tls_handshake_timeout\":5000000000}" \
				-%s "{\"batch_size\":1000,\"should_store_responses\":false}" \
				-%s "{\"strategy\":\"ROUND_ROBIN\",\"urls\":[\"http://api.bombardment.org\",\"http://mirror-1.bombardment.org\",\"http://mirror-2.bombardment.org\"]}" \
				-%s "{\"file_path\":\"./private/file_path.csv\",\"strategy\":\"CSV\"}" \
				-%s "{\"body_expression\":\"{\\n\\t\\t\\\"request_id\\\": \\\"bulk-create-\\\" & $number(row_id),\\n\\t\\t\\\"event_ts\\\": $millis(),\\n\\t\\\"user_account_id\\\": user_account_id,\\n\\t\\\"template_id\\\": \\\"4066f10464763823cc3e70c2ebd973fbd72cc5b1b450ccd31c0e87d9405e9dd6\\\",\\n\\t\\\"sms_date\\\": $millis(),\\n\\t\\\"insights\\\": $string({\\n\\t\\t\\\"billerName\\\": biller_name,\\n\\t\\\"last_four_dig_cc\\\": last_4_digits,\\n\\t\\\"mobile__number\\\": $floor($number(mobile_number))\\n\\t})\\n\\t}\",\"endpoint_expression\":\"\\\"/insight/v1/event/ingest\\\"\",\"headers_expression\":\"{ \\\"Content-Type\\\": \\\"application/json\\\" }\",\"method_expression\":\"\\\"POST\\\"\",\"strategy\":\"JSONATA\"}"`,
			CLIENT_CONTEXT_LONG_KEY,
			DRIVER_CONTEXT_LONG_KEY,
			LOAD_BALANCER_CONTEXT_LONG_KEY,
			PARSER_CONTEXT_LONG_KEY,
			TRANSFORMER_CONTEXT_LONG_KEY,
			CLIENT_CONTEXT_SHORT_KEY,
			DRIVER_CONTEXT_SHORT_KEY,
			LOAD_BALANCER_CONTEXT_SHORT_KEY,
			PARSER_CONTEXT_SHORT_KEY,
			TRANSFORMER_CONTEXT_SHORT_KEY,
		),
		Args: func(cmd *cobra.Command, args []string) error {
			// Get the Flags
			clientContextCommand := cmd.Flag(CLIENT_CONTEXT_LONG_KEY)
			driverContextCommand := cmd.Flag(DRIVER_CONTEXT_LONG_KEY)
			loadBalancerContextCommand := cmd.Flag(LOAD_BALANCER_CONTEXT_LONG_KEY)
			parserContextCommand := cmd.Flag(PARSER_CONTEXT_LONG_KEY)
			transformerContextCommand := cmd.Flag(TRANSFORMER_CONTEXT_LONG_KEY)

			// Check if the Flags are set properly
			var clientContext models_dto_clients.ClientContext
			var driverContext models_dto_driver.DriverContext
			var loadBalancerContext models_dto_load_balancing.LoadBalancerContext
			var parserContext models_dto_parsing.ParserContext
			var transformerContext models_dto_transforming.TransformerContext

			// Try Parsing the Client Context
			err := json.Unmarshal([]byte(clientContextCommand.Value.String()), &clientContext)
			if err != nil {
				return fmt.Errorf("error parsing client context: %v", err)
			}

			// Try Parsing the Driver Context
			err = json.Unmarshal([]byte(driverContextCommand.Value.String()), &driverContext)
			if err != nil {
				return fmt.Errorf("error parsing driver context: %v", err)
			}

			// Try Parsing the Load Balancer Context
			err = json.Unmarshal([]byte(loadBalancerContextCommand.Value.String()), &loadBalancerContext)
			if err != nil {
				return fmt.Errorf("error parsing load balancer context: %v", err)
			}

			// Try Parsing the Parser Context
			err = json.Unmarshal([]byte(parserContextCommand.Value.String()), &parserContext)
			if err != nil {
				return fmt.Errorf("error parsing parser context: %v", err)
			}

			// Try Parsing the Transformer Context
			err = json.Unmarshal([]byte(transformerContextCommand.Value.String()), &transformerContext)
			if err != nil {
				return fmt.Errorf("error parsing transformer context: %v", err)
			}
			return nil
		},
		Version: "v0.0.1",
		Run: func(cmd *cobra.Command, args []string) {
			// Get the Flags
			clientContextCommand := cmd.Flag(CLIENT_CONTEXT_LONG_KEY)
			driverContextCommand := cmd.Flag(DRIVER_CONTEXT_LONG_KEY)
			loadBalancerContextCommand := cmd.Flag(LOAD_BALANCER_CONTEXT_LONG_KEY)
			parserContextCommand := cmd.Flag(PARSER_CONTEXT_LONG_KEY)
			transformerContextCommand := cmd.Flag(TRANSFORMER_CONTEXT_LONG_KEY)

			// Parse the Client Context
			var clientContext models_dto_clients.ClientContext
			json.Unmarshal([]byte(clientContextCommand.Value.String()), &clientContext)

			// Parse the Driver Context
			var driverContext models_dto_driver.DriverContext
			json.Unmarshal([]byte(driverContextCommand.Value.String()), &driverContext)

			// Parse the Load Balancer Context
			var loadBalancerContext models_dto_load_balancing.LoadBalancerContext
			json.Unmarshal([]byte(loadBalancerContextCommand.Value.String()), &loadBalancerContext)

			// Parse the Parser Context
			var parserContext models_dto_parsing.ParserContext
			json.Unmarshal([]byte(parserContextCommand.Value.String()), &parserContext)

			// Parse the Transformer Context
			var transformerContext models_dto_transforming.TransformerContext
			json.Unmarshal([]byte(transformerContextCommand.Value.String()), &transformerContext)

			// Run the Bombardment
			runCliCallback(
				clientContext,
				driverContext,
				loadBalancerContext,
				parserContext,
				transformerContext,
			)
		},
	}
	cliCommand.Flags().StringP(
		CLIENT_CONTEXT_LONG_KEY,
		CLIENT_CONTEXT_SHORT_KEY,
		"",
		heredoc.Doc(
			`The Context to use for the Client that will make the calls.
			Client Context is a JSON string that contains the following keys:
				- channel: The channel to use for the client. Possible values are {REST, GRPC}
				- dial_keep_alive: The number of nanoseconds for which to keep connections alive. Eg. 10000000000 (10 seconds)
				- dial_timeout: The number of nanoseconds for which to wait for a connection to complete. Eg. 5000000000 (5 seconds)
				- expect_continue_timeout: The duration for which to wait for a server's FIRST response headers after fully writing the request headers. Post the timeout, the request will be sent without the Expect: 100-continue header. Eg. 500000 (500 milliseconds)
				- response_header_timeout: The duration for which to wait for the response headers. Eg. 5000000000 (5 seconds)
				- tls_handshake_timeout: The duration for the TLS handshake to complete. Post the timeout, the connection will be closed. Eg. 5000000000 (5 seconds)
			Eg. "{\"channel\":\"REST\",\"dial_keep_alive\":10000000000,\"dial_timeout\":5000000000,\"expect_continue_timeout\":500000,\"response_header_timeout\":5000000000,\"tls_handshake_timeout\":5000000000}"`,
		),
	)
	cliCommand.Flags().StringP(
		DRIVER_CONTEXT_LONG_KEY,
		DRIVER_CONTEXT_SHORT_KEY,
		"",
		heredoc.Doc(
			`The Context to use for the Driver that will orchestrate the Bombardment.
			Driver Context is a JSON string that contains the following keys:
				- batch_size: The number of records to send in a single batch. Eg. 1000
				- should_store_responses: A boolean flag to indicate if the responses should be stored. Default is false.
			Eg. "{\"batch_size\":1000,\"should_store_responses\":false}"`,
		),
	)
	cliCommand.Flags().StringP(
		LOAD_BALANCER_CONTEXT_LONG_KEY,
		LOAD_BALANCER_CONTEXT_SHORT_KEY,
		"",
		heredoc.Doc(
			`The Context to use for Load Balancing requests across servers. All servers must support the same API contract. Eg. Different pods of the same service.
			Load Balancer Context is a JSON string that contains the following keys:
				- strategy: The strategy to use for the load balancer. Possible values are {ROUND_ROBIN}
				- urls: The list of URLs to use for the load balancer.
			Eg. "{\"strategy\":\"ROUND_ROBIN\",\"urls\":[\"http://api.bombardment.org\",\"http://mirror-1.bombardment.org\",\"http://mirror-2.bombardment.org\"]}"`,
		),
	)
	cliCommand.Flags().StringP(
		PARSER_CONTEXT_LONG_KEY,
		PARSER_CONTEXT_SHORT_KEY,
		"",
		heredoc.Doc(
			`The Context to use for Parsing the input data.
			Parser Context is a JSON string that contains the following keys:
				- strategy: The strategy to use for parsing the file. Possible values are {CSV, JSON}
				- file_path: The path to the file to parse. Eg. ./private/gupi_sms_credit_card.csv (if the strategy is CSV)
			Eg. "{\"file_path\":\"./private/file_path.csv\",\"strategy\":\"CSV\"}"`,
		),
	)
	cliCommand.Flags().StringP(
		TRANSFORMER_CONTEXT_LONG_KEY,
		TRANSFORMER_CONTEXT_SHORT_KEY,
		"",
		heredoc.Doc(
			`The Context to use for Transforming the parsed data.
			Transformer Context is a JSON string that contains the following keys:
				- strategy: The strategy to use for transforming the data. Possible values are {JSONATA}
				- endpoint_expression: The expression to use for the endpoint of the request. Eg. "/insight/v1/event/ingest" (if the strategy is JSONATA)
				- headers_expression: The expression to use for the headers of the request. Eg. { "Content-Type": "application/json" } (if the strategy is JSONATA)
				- method_expression: The expression to use for the method of the request. Eg. "POST" (if the strategy is JSONATA)
				- body_expression: The expression to use for the body of the request. Eg. "{\"request_id\": \"bulk-create-\" & $number(row_id),\"event_ts\": $millis(),\"user_account_id\": user_account_id,\"template_id\": \"T123\",\"sms_date\": $millis(),\"insights\": $string({\"billerName\": biller_name,\"last_four_dig_cc\": last_4_digits,\"mobile__number\": $floor($number(mobile_number))})\"}" (if the strategy is JSONATA)
			Eg. "--transformation_context '{"strategy":"JSONATA","method_expression":"\"POST\"","endpoint_expression":"\"/api/v1/\" & resource","headers_expression":"{ \"Content-Type\": \"application/json\", \"X-Request-ID\": request_id }","body_expression":"{ \"id\": $number(id), \"timestamp\": $millis() }"}"`,
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
		GroupID: RUN_MODE_GROUP_ID,
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
			DEFAULT_SERVER_BIND_ADDR,
			DEFAULT_SERVER_PORT,
			BIND_ADDR_LONG_KEY,
			DEFAULT_SERVER_BIND_ADDR,
			PORT_LONG_KEY,
			DEFAULT_SERVER_PORT,
			BIND_ADDR_SHORT_KEY,
			DEFAULT_SERVER_BIND_ADDR,
			PORT_SHORT_KEY,
			DEFAULT_SERVER_PORT,
		),
		Args: func(cmd *cobra.Command, args []string) error {
			port, err := cmd.Flags().GetInt(PORT_LONG_KEY)
			if err != nil {
				return err
			}
			if port < 1 || port > 65535 {
				return fmt.Errorf("invalid port: %d", port)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			bindAddr, _ := cmd.Flags().GetString(BIND_ADDR_LONG_KEY)
			port, _ := cmd.Flags().GetInt(PORT_LONG_KEY)

			runServerCallback(bindAddr, port)
		},
		Version: "v0.0.1",
	}
	// Define flags for server command
	serverCommand.Flags().StringP(
		BIND_ADDR_LONG_KEY,
		"b",
		DEFAULT_SERVER_BIND_ADDR,
		heredoc.Docf("Bind address for the server (default %s)", DEFAULT_SERVER_BIND_ADDR),
	)
	serverCommand.Flags().IntP(
		PORT_LONG_KEY,
		"p",
		DEFAULT_SERVER_PORT,
		heredoc.Docf("Port for the server (default %d)", DEFAULT_SERVER_PORT),
	)
	c.rootCmd.AddCommand(&serverCommand)
	return c
}

func (c *cobraCliHooks) Execute() {
	c.rootCmd.Execute()
}
