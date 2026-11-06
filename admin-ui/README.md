# NextGate Admin UI

The NextGate administrator interface is a React and Vite application. The
current screen handles administrator login against the admin API at
`/api/v1/login`.

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
