---
decision_date: 2026-03-16
pattern: strategy-pattern
repository: bombardment-runner
status: Discovered
tags: [adr, strategy-pattern, extensibility, generics]
title: "0001. Strategy Pattern for Pipeline Extensibility"
type: adr
---

# 0001. Strategy Pattern for Pipeline Extensibility

Date: 2026-03-16

## Status

Discovered (existing architecture)

## Context

The Bombardment Runner processing pipeline consists of four pluggable stages: file parsing, data transformation, load balancing, and HTTP client execution. Each stage needs to support multiple implementations (e.g., CSV vs JSON parsing, round-robin vs random load balancing) while keeping the driver orchestrator agnostic to concrete types.

Evidence:

- `BaseStrategy[T]` generic interface: `services/base_strategy.go:4-7`
- Factory pattern in each stage:
  - `CreateFileParser()`: `services/parsing/base_file_parser.go:37-48`
  - `CreateTransformer()`: `services/transforming/base_transformer.go:19-29`
  - `CreateLoadBalancer()`: `services/load_balancing/base_client_load_balancer.go:21-36`
  - `CreateChannelClient()`: `services/clients/base_channel_client.go:23-32`
- Enum-driven selection: `models/enums/parser_strategy.go`, `client_channels.go`, `load_balancer_strategy.go`, `transformer_strategy.go`

## Decision

The system uses the Strategy pattern with Go generics and enum-driven factory functions to achieve pipeline extensibility. Each stage defines:

1. A base interface embedding `BaseStrategy[T]` where `T` is the stage's enum type
2. A factory function (`Create*()`) that switches on the enum to instantiate concrete implementations
3. Concrete implementations in the same package

The driver (`BombardmentDriver`) interacts only with base interfaces, never with concrete types.

## Consequences

**Benefits**:

- Adding a new strategy requires only: implementing the interface, adding an enum value, and registering in the factory switch. No changes to the driver or controllers.
- Enum values defined but not yet implemented (GRPC, KAFKA, GOTEMPLATE, LEAST_CONNECTION) document the intended extension points.
- Go generics (`BatchProcessor[T, R]`, `BaseFileParser[T]`) provide type safety without code duplication.

**Trade-offs**:

- Factory switch statements must be kept in sync with enum definitions manually.
- The strategy enum is embedded in DTOs and serialized as strings, coupling API contracts to internal enum values.
- Each new strategy requires touching three locations (interface impl, enum, factory), though each change is small.
