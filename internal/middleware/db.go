package middleware

import (
	"database/sql"
	"fmt"
	"net/http"

	rctx "github.com/reststore/restkit/internal/context"
)

// DBMiddleware injects database queries into every request context.
func DBMiddleware(queries any) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(rctx.WithQueries(r.Context(), queries)))
		})
	}
}

// NewQueriesFunc creates queries from a database connection.
type NewQueriesFunc func(*sql.DB) any

// WithTxFunc creates transactional queries.
type WithTxFunc func(queries any, tx *sql.Tx) any

// TransactionMiddleware wraps requests in a transaction.
// Commits on 2xx status codes, rolls back otherwise.
func TransactionMiddleware(
	database *sql.DB,
	newQueries NewQueriesFunc,
	withTx WithTxFunc,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tx, err := database.Begin()
			if err != nil {
				http.Error(
					w,
					fmt.Sprintf("failed to begin transaction: %v", err),
					http.StatusInternalServerError,
				)
				return
			}

			queries := withTx(newQueries(database), tx)
			ctx := rctx.WithQueries(r.Context(), queries)
			rr := newResponseRecorder(w)

			defer func() {
				if rec := recover(); rec != nil {
					_ = tx.Rollback()
					panic(rec)
				}
				if rr.statusCode >= 200 && rr.statusCode < 300 {
					_ = tx.Commit()
				} else {
					_ = tx.Rollback()
				}
			}()

			next.ServeHTTP(rr, r.WithContext(ctx))
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rr *responseRecorder) WriteHeader(code int) {
	if !rr.written {
		rr.statusCode = code
		rr.written = true
		rr.ResponseWriter.WriteHeader(code)
	}
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.written {
		rr.WriteHeader(http.StatusOK)
	}
	return rr.ResponseWriter.Write(b)
}
