# Bombardment

Bombardment is a lightweight automation tool intended to pick up data, transform it using a set of rules and then send it to a target system. It is designed to perform small repetitive migrations of data from one system to another. It supports concurrent processing of data, client-side load balancing strategies, and is designed to be extensible and reusable.

## Why Bombardment?

- Bombardment is written in Golang to be fast, lightweight and scalable enough to process large amounts of data
- Bombardment is designed to be easily extensible
- Bombardment supports batched concurrent processing of data
- Bombardment supports different channels (REST / GRPC etc) to send data to target systems
- Bombardment supports different client-side load balancing strategies

## Features

- **Modular Architecture**: Built with clean, layered architecture for easy extension and maintenance
- **Data Processing**: Supports various file formats through extensible parsers
- **Transformation Rules**: Transform data using powerful rule engines like JSONata
- **Multiple Client Channels**: REST API support with more channels planned
- **Load Balancing**: Client-side load balancing with Round Robin strategy
- **Concurrency**: Process data in batches concurrently for higher throughput
- **Job Management**: Track and manage data migration jobs

## Installation

```bash
go get dhi13man.github.io/credit_card_bombardment
```

## Usage

### CLI Mode

```bash
bombardment run [options]
```

## Project Structure

The project follows a clean architecture with:

- **Core**: CLI and bootstrap functionality
- **Domain**: Business logic, repositories, and services
- **Models**: DTOs, entities, and enums
- **App**: Application bootstrap and configuration

## To Do

- [x] Initial setup with scalable architecture
- [x] Define basic DTOs and Data Models
- [x] Implement workgroup and worker pool
- [x] Implement basic clients: REST
- [x] Implement basic load balancing strategy: Round Robin
- [x] Implement concurrent batch processing using channels
- [ ] Implement basic data transformation rules: JSONata
- [ ] Implement a state machine for Start, Pause, Resume, Stop
- [ ] Implement advanced progress tracking
- [ ] Basic UI for monitoring and control
