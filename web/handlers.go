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
		if err := resultsTempl("", Timeline{}).Render(r.Context(), w); err != nil {
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

		// render sends a new frame, and signals every 100ms
		render := func() error {
			rows, err := q.GetVotesPerSecond(r.Context())
			if err != nil {
				return err
			}

			t := buildTimeline(rows, time.Now())
			chartJSON, err := t.EChartsOptions()
			if err != nil {
				return err
			}
			if err := sse.PatchElementTempl(resultsTempl("", t)); err != nil {
				return err
			}
			// The chart is drawn client-side; vote-chart.js picks this signal up
			return sse.MarshalAndPatchSignals(map[string]string{"_chartJson": string(chartJSON)})
		}

		for {
			select {
			case <-ticker.C:
				if err := render(); err != nil {
					slog.Debug("sse results ended", "err", err)
					return
				}
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
