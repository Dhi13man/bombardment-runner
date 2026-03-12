# Contributing to Bombardment

Contributions are welcome! Whether it's bug reports, feature requests, or code contributions, we appreciate your help making Bombardment better. Please open an issue on [GitHub Issues](https://github.com/Dhi13man/bombardment/issues) to get started.

## Prerequisites

- **Go 1.23+**
- **Docker** (optional, for containerized development)
- **golangci-lint** (optional, for linting)

## Development Setup

```bash
git clone https://github.com/Dhi13man/bombardment.git
cd bombardment/app
go mod download
go run main.go server
# Visit http://localhost:8080
```

## Project Structure

```
bombardment/
├── app/                    # Go application
│   ├── main.go            # Entry point
│   ├── docs/              # Swagger docs
│   └── src/
│       ├── app/           # HTTP server, controllers, UI
│       ├── models/        # DTOs, entities, enums
│       ├── repositories/  # Data access
│       └── services/      # Business logic
│           ├── batching/      # Concurrent batch processor
│           ├── clients/       # Channel clients (REST, etc.)
│           ├── driver/        # Bombardment orchestrator
│           ├── load_balancing/ # Load balancer strategies
│           ├── parsing/       # File parsers (CSV, JSON)
│           └── transforming/  # Request transformers (JSONata)
├── landing-site/          # Marketing website
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## How to Contribute

1. **Fork** the repository and create a feature branch from `main`.
2. **Use conventional commits** for your commit messages:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `test:` for adding or updating tests
   - `docs:` for documentation changes
   - `chore:` for maintenance tasks
3. **Run tests** before submitting:

   ```bash
   go test -race ./...
   ```

4. **Submit a pull request** with a clear description of your changes and the problem they solve.

## Adding New Strategies

Bombardment uses the strategy pattern extensively. Here's how to add new implementations:

### Adding a New Parser

1. Implement the `BaseFileParser[T]` interface in a new file under `app/src/services/parsing/`.
2. Add a new enum value to `app/src/models/enums/parser_strategy.go`.
3. Add the corresponding case to the factory in `app/src/services/parsing/base_file_parser.go`.

### Adding a New Load Balancer

1. Implement the `BaseLoadBalancer` interface in a new file under `app/src/services/load_balancing/`.
2. Add a new enum value to `app/src/models/enums/load_balancer_strategy.go`.
3. Add the corresponding case to the factory in `app/src/services/load_balancing/base_client_load_balancer.go`.

### Adding a New Client Channel

1. Implement the `BaseChannelClient` interface in a new file under `app/src/services/clients/`.
2. Add a new enum value to `app/src/models/enums/client_channels.go`.
3. Add the corresponding case to the factory in `app/src/services/clients/base_channel_client.go`.

### Adding a New Transformer

1. Implement the `BaseTransformer` interface in a new file under `app/src/services/transforming/`.
2. Add a new enum value to `app/src/models/enums/transformer_strategy.go`.
3. Add the corresponding case to the factory in `app/src/services/transforming/base_transformer.go`.

## Code Style

- Run `gofmt` on all Go files before committing.
- Do not leave unused imports in your code.
- Use structured logging with `zap`.
- Define interfaces for testability and dependency injection.

## Code of Conduct

This project follows the [Contributor Covenant v2.1](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). Be respectful and constructive in all interactions.
