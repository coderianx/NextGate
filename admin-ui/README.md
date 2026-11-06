# NextGate Admin UI

The NextGate administrator interface is a React and Vite application. It handles
administrator login against `/api/v1/login`, stores the returned JWT in session
storage, and redirects to `/dashboard`. The dashboard uses that JWT to load the
configured providers from `/api/v1/providers`. It also manages the complete
admin provider workflow: adding and deleting providers, loading a provider's
models, adding and deleting models, and enabling or disabling them.

The production build in `dist/` is embedded into the Go binary by `embed.go`
and served at the root of the admin server.

## Development

Start the Go admin server on its default port, then run Vite:

```sh
npm install
npm run dev
```

Vite proxies `/api` requests to `http://127.0.0.1:9090` during development.

## Build

```sh
npm run build
```

The generated `dist/` directory is embedded into the Go application. From the
repository root, the same build can be triggered with:

```sh
go generate ./...
```

Commit the updated `dist/` output whenever the interface changes so a plain
`go build ./...` produces a complete binary.
