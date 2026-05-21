# traefik-correlation-id
Correlation id is a Traefik middleware that ensures every incoming HTTP request carries a unique Correlation-ID header. If the header is already present, it's passed through unchanged; if not, the middleware generates one automatically. This makes individual requests traceable end-to-end across services, simplifying debugging and observability.
