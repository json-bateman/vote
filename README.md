# vote

A minimal starter for [Datastar](https://data-star.dev) apps in Go: [chi](https://github.com/go-chi/chi) router, [templ](https://templ.guide) views, self-hosted fonts, Bootstrap 5.

## Stack

- **Router:** chi
- **Views:** templ
- **Frontend:** Datastar (client lib in `web/static/js`), 
- **Styling** Bootstrap 5 and IBMPlexMono for the font because it's awesome
- **Static assets:** embedded in the binary, served content-hashed (`hashfs`) for aggressive caching
- **Dev loop:** `air` (rebuild + reload on save)
- **Task runner:** [Task](https://taskfile.dev) (`Taskfile.yml`)

## Running

```sh
task setup      # install pinned tool deps (air, templ) and tidy go.mod
task            # run the dev server with hot reload (default task)
```
Server listens on `:59876` by default (override with `PORT` in `.env` or the environment).

## File naming convention

Inside `web/`:

- `p_*.templ` — a **page**: a full route's top-level template (e.g. `p_home.templ`, `p_404.templ`)
- `pc_*.templ` — a **page component**: shared/reusable pieces pages are built from (e.g. `pc_layout.templ`)

## Building for production

```sh
task go:build
```

## Adding a page

1. Add `web/p_<name>.templ`
2. Register its route in `web/httpServer.go` (`setupRoutes`).
3. Reference any new static assets via `StaticPath("css/foo.css")` / `StaticPath("js/foo.js")` — never link static files directly, so caching stays correct.

## Based on

The `nats-sqlite` branch of [datastar-template](https://github.com/json-bateman/datastar-template), using [NATS](https://github.com/nats-io/nats.go) and [sqlite3](https://www.sqlite.org/lang.html)
