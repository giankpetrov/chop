# chop - Project Instructions

## Tooling runs inside Docker - never on the host

All development tooling (Go, Python, or any other runtime) runs inside the Docker container. The host machine does not have language runtimes installed.

Always use `docker compose run --rm dev <command>`:

```bash
docker compose run --rm dev go build ./...
docker compose run --rm dev go test ./...
docker compose run --rm dev go vet ./...
docker compose run --rm dev go mod tidy
```

Never run `go`, `python`, `node`, or any runtime directly on the host - it will fail.
