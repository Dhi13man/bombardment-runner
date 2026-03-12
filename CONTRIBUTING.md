# Contributing to Bombardment

[![License](https://img.shields.io/github/license/dhi13man/bombardment-runner)](https://github.com/Dhi13man/bombardment-runner/blob/main/LICENSE)
[![Contributors](https://img.shields.io/github/contributors-anon/dhi13man/bombardment-runner?style=flat)](https://github.com/Dhi13man/bombardment-runner/graphs/contributors)
[![GitHub forks](https://img.shields.io/github/forks/dhi13man/bombardment-runner?style=social)](https://github.com/Dhi13man/bombardment-runner/network/members)
[![GitHub Repo stars](https://img.shields.io/github/stars/dhi13man/bombardment-runner?style=social)](https://github.com/Dhi13man/bombardment-runner/stargazers)
[![Last Commit](https://img.shields.io/github/last-commit/dhi13man/bombardment-runner)](https://github.com/Dhi13man/bombardment-runner/commits/main)

Thank you for investing your time in contributing to this project! Any contributions you make have a chance of improving the tool for everyone, and that brightens up my day. :)

## Contents

- [Contributing to Bombardment](#contributing-to-bombardment)
  - [Contents](#contents)
  - [Code of Conduct](#code-of-conduct)
  - [Prerequisites](#prerequisites)
  - [Development Setup](#development-setup)
  - [General Steps to Contribute](#general-steps-to-contribute)
  - [Recommended Development Workflow](#recommended-development-workflow)
  - [Adding New Strategies](#adding-new-strategies)
  - [Code Style](#code-style)
  - [Issue-Based Contributions](#issue-based-contributions)

## Code of Conduct

This project follows the [Contributor Covenant v2.1](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior via [GitHub Issues](https://github.com/Dhi13man/bombardment-runner/issues).

## Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) (optional, for containerized development)
- [golangci-lint](https://golangci-lint.run/welcome/install/) (optional, for linting)

## Development Setup

```bash
git clone https://github.com/Dhi13man/bombardment-runner.git
cd bombardment-runner/app
go mod download
go run main.go server
# Visit http://localhost:8080
```

Alternatively, use Docker:

```bash
docker compose up --build
```

## General Steps to Contribute

1. Ensure you have **Go 1.23+** installed.

2. **Fork** the [project repository](https://github.com/Dhi13man/bombardment-runner).

3. **Clone** your fork by running `git clone <forked-repository-git-url>`.

4. Navigate to the project directory: `cd bombardment-runner`.

5. Pull the latest changes from upstream: `git pull origin main`.

6. Create a new branch: `git checkout -b <new-branch-name>`.

7. Make your changes:
   - Controller and server code goes to `app/src/app/`.
   - Data models go to `app/src/models/`.
   - Service implementations go to `app/src/services/`.

8. Add relevant tests for your changes to the appropriate `*_test.go` file next to the code you changed.

9. **Run tests** and ensure they all pass before committing:

   ```bash
   cd app && go test -race ./...
   ```

10. **Run the linter** if you have golangci-lint installed:

    ```bash
    golangci-lint run ./...
    ```

11. **Use conventional commits** for your commit messages:
    - `feat:` for new features
    - `fix:` for bug fixes
    - `test:` for adding or updating tests
    - `docs:` for documentation changes
    - `refactor:` for code restructuring
    - `chore:` for maintenance tasks

12. Push your branch and create a pull request with a clear description of the problem your changes solve.

## Recommended Development Workflow

Fork Project **->** Create new Branch **->** For each contribution:

**Develop** -> **Write Tests** -> **Run Tests** -> **Lint** -> **Commit** -> **Create Pull Request**

## Adding New Strategies

Bombardment uses the strategy pattern extensively. Here is how to add new implementations:

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
- Write tests for all new functionality.

## Issue-Based Contributions

### Create a New Issue

If you spot a problem or bug, search if an [issue](https://github.com/Dhi13man/bombardment-runner/issues) already exists. If a related issue doesn't exist, open a new one.

### Solve an Issue

Scan through [existing issues](https://github.com/Dhi13man/bombardment-runner/issues) to find one that interests you. You can narrow down the search using labels as filters.

---

Contributions are welcome on [GitHub](https://github.com/Dhi13man/bombardment-runner). Please ensure all the tests are running before pushing your changes. Write your own tests too!

File any [issues or feature requests here](https://github.com/Dhi13man/bombardment-runner/issues), or help resolve existing ones. :)
