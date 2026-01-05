# Repository Guidelines

## Project Structure & Module Organization

- `cmd/gimedic/` contains the CLI entrypoint and subcommands.
- `app/` holds core app logic and shared helpers.
- `gen.go` defines `go:generate` tasks (manual/usage generation via `gimedic man`).
- `testdata/` and `usage/` are reserved for test fixtures and generated docs; both are currently empty.
- Root assets include `hinshi.png` and sample `user_dictionary*.db` files.

## Build, Test, and Development Commands

- `make generate` runs `go generate -x ./...` after clearing `*_gen.go` files.
- `make generate-clear` removes generated `*_gen.go` files.
- `make lint` runs `golangci-lint` with the repo config.
- `make test` runs `go test -tags man -v --race ./...`.
- `make install` builds and installs `./cmd/...` with version metadata.

## Coding Style & Naming Conventions

- Go code is formatted with `gofmt` and organized with `goimports` (see `.golangci.yml`).
- Follow standard Go naming: exported identifiers use `CamelCase`, packages use short, lower-case names.
- Keep generated files suffixed with `_gen.go` to match the cleanup target.

## Testing Guidelines

- No `_test.go` files exist yet; add tests alongside packages using standard Go naming (`*_test.go`).
- Prefer table-driven tests and name test functions `TestXxx`.
- Run `make test` to include the `man` build tag used by generators.

## Commit & Pull Request Guidelines

- Git history is minimal; no commit convention is established. Use clear, imperative summaries (e.g., "Add dictionary decoder").
- PRs should include: summary of changes, rationale, and any manual test output (e.g., `make test`).
- If changes touch generated files, mention the generator command used.

## Configuration Notes

- `golangci-lint` is expected for linting; install it locally or via your CI tools.
- `go generate` invokes the CLI with `-tags man`, so keep `cmd/gimedic` runnable.
