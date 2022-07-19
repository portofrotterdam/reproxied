# ReProxied

ReProxied is a middleware plugin for [Traefik](https://github.com/traefik/traefik) to route an incoming request through a proxy.
Be aware that this middleware initiates the call to the proxy and any middlewares after this one will be skipped. If the request to the proxy itself fails the middleware will respond with a 502 bad gateway response.

## Configuration

### Static

```yaml
pilot:
  token: "xxxxx"

experimental:
  plugins:
    reproxied:
      moduleName: "github.com/portofrotterdam/reproxied"
      version: "v0.0.3"
```

### Dynamic

```yaml
http:
  middlewares:
    reproxied-foo:
      reproxied:
        proxy: http://proxyHost:3128
        targetHost: https://example.com
```
