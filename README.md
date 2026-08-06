# gateway-proxy

The authentication gateway in front of the weeb.vip GraphQL router — known as
`smokey` on the command line.

Every request to the API arrives here first. It validates the bearer token,
resolves who the caller is, and passes the request on to the router with that
identity attached as headers, so the services behind it never have to parse a
token themselves.

## What it does to a request

- Parses the JWT from the `Authorization` header and pulls out the subject
  (user id) and the token's purpose.
- Forwards to the configured backend (the Apollo router) with `x-user-id`,
  `x-token-purpose`, `x-raw-token`, `x-remote-ip` and `x-user-agent` set, and
  the client's own `User-Agent` and `x-forwarded-for` stripped.
- Applies CORS, response caching, logging, metrics and tracing on the way
  through — each a middleware under `http/middlewares`.

Signing keys are not held locally. `internal/keys` fetches them from the key
management service over GraphQL, and `internal/poller` re-fetches them on an
interval (15 minutes by default) so a rotated key is picked up without a
restart.

## Running it

Requires Go, and something to proxy to.

```sh
docker-compose up                  # postgres, key-management-service, echo server
go run ./cmd/cli/ server start     # the proxy itself, on :8080
```

`server start text-formatter` swaps the JSON log output for something readable
at a terminal.

```sh
go test ./...                      # tests
go build -o main ./cmd/cli/        # build
```

## Configuration

[configor](https://github.com/jinzhu/configor) reads `config/config.json` and
lets environment variables override it. `config/config.go` has the full shape;
the settings that matter most are the proxy target, the port, the GraphQL
endpoint the keys come from, the allowed CORS origins, and the polling
interval.

`auth_mode` decides how strict it is about tokens — `both` accepts
authenticated and anonymous callers, leaving the services behind it to decide
what an anonymous one may do.

## Docker

```sh
docker build --build-arg VERSION=local -t gateway-proxy .
docker run -p 8080:8080 gateway-proxy
```

The image runs `./main server start` by default.
