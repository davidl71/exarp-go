package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// QueryContextDB centralizes the context + timeout + database resolution
// used by read-only operations. It mirrors the ensureContext + withQueryTimeout
// pattern spread across the datastore and returns a cancel function that must
// be deferred by callers (even when the helper returns an error, the cancel
// function is invoked before the error is returned).
func QueryContextDB(ctx context.Context) (context.Context, context.CancelFunc, *sqlx.DB, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	queryCtx, cancel := withQueryTimeout(ctx)

	db, err := GetDBx()
	if err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("failed to get database: %w", err)
	}

	return queryCtx, cancel, db, nil
}
