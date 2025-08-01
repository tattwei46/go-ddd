# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go-based payment service implementing Domain Driven Design (DDD) patterns with comprehensive audit functionality. The system demonstrates clean architecture principles with clear separation of concerns across domain, application, and infrastructure layers.

## Development Commands

- **Run the application**: `go run main.go`
- **Build the application**: `go build`
- **Run tests**: `go test ./...`
- **Format code**: `go fmt ./...`
- **Vet code**: `go vet ./...`
- **Add dependencies**: `go mod tidy`

## Architecture

The project follows DDD principles with a layered architecture:

### Domain Layer (`internal/domain/`)
- **Payment Domain**: Core payment entities, value objects, and business logic
  - `Payment` entity with status transitions (pending → processing → completed/failed/cancelled)
  - `Amount` value object with currency validation
  - `PaymentID` value object for type safety
  - Domain service for payment operations
- **Audit Domain**: Complete audit trail functionality
  - `AuditEntry` entity for tracking all changes
  - Support for filtering audit history by various criteria
  - Automatic capture of old/new data states

### Application Layer (`internal/application/`)
- `PaymentApplicationService`: Orchestrates payment operations with automatic audit logging
- Coordinates between payment and audit domains
- Handles cross-cutting concerns like transaction management

### Infrastructure Layer (`internal/infrastructure/`)
- In-memory repository implementations for both Payment and Audit domains
- Thread-safe operations using sync.RWMutex
- Ready for replacement with database implementations

### Key DDD Patterns Implemented
- **Entities**: Payment, AuditEntry with identity and lifecycle
- **Value Objects**: PaymentID, Amount, AuditID with immutability
- **Domain Services**: Business logic that doesn't belong to entities
- **Repository Pattern**: Abstract data access from domain logic
- **Application Services**: Use case orchestration

### Audit Features
- Automatic audit trail for all payment operations
- Captures complete before/after state changes
- User tracking for accountability
- Flexible filtering and querying capabilities
- Extensible metadata support