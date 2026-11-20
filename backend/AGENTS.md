# Backend Guidelines

## Project Structure & Module Organization

This is a Go module; `cmd/server/main.go` is its server entry point. Keep backend packages under this directory.

## Build, Test, and Development Commands

- `go run ./cmd/server` runs the Go entry point, which currently logs a startup message.
- `go build ./...` builds all Go packages.
- `go test ./...` runs their tests.

## Coding Style & Naming Conventions

Use Go's standard formatting (`gofmt`) and idiomatic package and exported identifier names.

## Testing Guidelines

Add Go tests alongside the code they exercise, using `*_test.go` files and the standard `testing` package. Run `go test ./...`.

