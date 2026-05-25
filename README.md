# Traefik Correlation ID

A [Traefik](https://traefik.io) middleware plugin that ensures every incoming HTTP request carries a unique Correlation ID header. If the header is already present it is passed through unchanged; if not, the middleware generates a UUID v4 automatically. This makes individual requests traceable end-to-end across services, simplifying debugging and observability.

## Features

- Generates a cryptographically random UUID v4 when the header is absent
- Passes existing headers through unchanged, preserving IDs set by upstream callers
- Configurable header name (default: `correlation-id`)
- Zero external dependencies

## Installation

### Traefik Pilot / Plugin Catalog

Add the plugin to your Traefik static configuration:

```yaml
# traefik.yml
experimental:
  plugins:
    correlation:
      moduleName: github.com/surfe/traefik-correlation-id
      version: v0.1.0
```

### Local development

Mount the plugin source and register it as a local plugin:

```yaml
# traefik.yml
experimental:
  localPlugins:
    correlation:
      moduleName: github.com/surfe/traefik-correlation-id
```

The included `docker-compose.yml` sets this up automatically.

## Configuration

| Field | Type | Default | Description |
|---|---|---|---|
| `headerName` | `string` | `correlation-id` | Name of the header to read and inject |

### Static configuration (traefik.yml)

```yaml
experimental:
  plugins:
    correlation:
      moduleName: github.com/surfe/traefik-correlation-id
      version: v0.1.0
```

### Dynamic configuration

**File provider:**

```yaml
http:
  middlewares:
    correlation:
      plugin:
        correlation:
          headerName: correlation-id

  routers:
    my-router:
      middlewares:
        - correlation
```

**Docker labels (per service):**

```yaml
labels:
  - "traefik.http.middlewares.correlation.plugin.correlation.headerName=correlation-id"
  - "traefik.http.routers.my-service.middlewares=correlation@docker"
```

**Global middleware (all routes on an entrypoint):**

```yaml
# traefik.yml
entryPoints:
  web:
    address: ":80"
    http:
      middlewares:
        - correlation@docker
```

## Local development with Docker

The `docker-compose.yml` starts Traefik with the plugin loaded from the local source tree and a `whoami` service to test against.

```bash
docker compose up
```

Then verify the header is injected:

```bash
curl -H "Host: whoami.localhost" http://localhost
# Response will contain: correlation-id: <uuid>

# Existing header is preserved:
curl -H "Host: whoami.localhost" -H "correlation-id: my-trace-id" http://localhost
# Response will contain: correlation-id: my-trace-id
```

The Traefik dashboard is available at [http://localhost:8080](http://localhost:8080).

## Running tests

```bash
go test -v -cover ./...
```

## License

[MIT](LICENSE)
