# vote

Live audience voting for presentations. People open `/` on their phones and tap **Cheer** or **Boo** as often as they like; `/results` streams the tally to the projector in near real time.

## How it works

- `POST /vote/{cheer|boo}` saves a row in SQLite. Voters are identified by a random `voter_id` cookie (no login).
- `/results` opens a Datastar SSE stream (`/sse/results`) that checks for new votes every 100ms. When something changed it patches the totals and sends new [ECharts](https://echarts.apache.org) options (built with [go-echarts](https://github.com/go-echarts/go-echarts)) in a signal, which `static/js/vote-chart.js` draws: 5s chunks over the last 5 minutes.

## Stack

- **Go** + [chi](https://github.com/go-chi/chi) router
- [templ](https://templ.guide) views, [Datastar](https://data-star.dev) for SSE/hypermedia, Bootstrap 5
- **SQLite** with [goose](https://github.com/pressly/goose) migrations (`sql/migrations`) and [sqlc](https://sqlc.dev) queries (`sql/queries` → `sql/sqlcgen`)
- [air](https://github.com/air-verse/air) hot reload, [Task](https://taskfile.dev) runner; tools pinned via `go.mod`'s `tool` block

## Running

```sh
task setup   # install pinned tools
task         # dev server with hot reload on http://localhost:59876
```

Config via `.env` or environment: `PORT` (default `59876`), `DB_PATH` (default `./data/app.db`).

Migrations run on startup. Editing an already-applied migration does nothing; add a new `00N_*.sql`, or in dev delete `data/app.db*` to rebuild.

## Layout

- `web/p_*.templ`: pages; `web/pc_*.templ`: shared components
- `Base` is the bare HTML shell (full-screen pages like voting); `Layout` adds the header and footer
