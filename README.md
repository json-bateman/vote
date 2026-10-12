# Vote!

Live audience voting. People open `/` on their phones and tap **Cheer** or **Boo**; `/results` streams the voting chart.

### Stack
- Go 
- chi 
- templ 
- Datastar 
- ECharts 
- SQLite (Goose + SQLC) 
- Bootstrap

### Installation
```sh
task setup   # install pinned tools
task         # dev server with hot reload on http://localhost:59876
```

Config via `.env`: `PORT` (default `59876`), `DB_PATH` (default `./data/app.db`). Migrations run on startup.
