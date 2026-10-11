package web

import (
	"crypto/rand"
	"database/sql"
	"log/slog"
	"net"
	"net/http"
	"time"

	"vote/sql/sqlcgen"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func resultsPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := resultsTempl("").Render(r.Context(), w); err != nil {
			slog.Debug("render error", "component", "NatsSqlite", "err", err)
		}
	}
}

func sseResults(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sse := datastar.NewSSE(w, r, datastar.WithCompression(datastar.WithBrotli()))
		q := sqlcgen.New(db)

		// Tick every 100ms, batching any votes that arrived since the last frame
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		lastCount := int64(-1)
		// render sends a new frame if the vote count changed since the last one.
		render := func() error {
			count, err := q.CountVotes(r.Context())
			if err != nil || count == lastCount {
				return err
			}
			votes, err := q.GetAllVotes(r.Context())
			if err != nil {
				return err
			}
			lastCount = count
			return sse.PatchElementTempl(voteLineGraph(votes))
		}

		for {
			// First frame right away instead of waiting for the first tick
			if err := render(); err != nil {
				slog.Debug("sse results ended", "err", err)
				return
			}
			select {
			case <-ticker.C: // waits here before the next render
			case <-r.Context().Done():
				return
			}
		}
	}
}

func votePage(w http.ResponseWriter, r *http.Request) {
	if err := voteTempl("").Render(r.Context(), w); err != nil {
		slog.Debug("render error", "component", "NotFound", "err", err)
	}
}

func votePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// The route only matches {type:cheer|boo}, so anything else 404s before reaching here.
		voteType := chi.URLParam(r, "type")

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		row, err := sqlcgen.New(db).AddVote(r.Context(), sqlcgen.AddVoteParams{
			Vote:    voteType == "cheer",
			VoterID: voterID(w, r),
			Ip:      ip,
		})

		if err != nil {
			slog.Error("query error", "component", "AddVote", "err", err)
			http.Error(w, "could not save vote", http.StatusInternalServerError)
			return
		}

		slog.Debug("vote saved", "id", row.ID, "type", voteType, "voter", row.VoterID)
		w.WriteHeader(http.StatusNoContent)
	}
}

// voterID returns the visitor's voter_id cookie, creating one if it's missing.
func voterID(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie("voter_id"); err == nil && c.Value != "" {
		return c.Value
	}
	id := rand.Text()
	http.SetCookie(w, &http.Cookie{
		Name:     "voter_id",
		Value:    id,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return id
}
