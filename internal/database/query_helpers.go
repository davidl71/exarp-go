package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// QueryContextDB is the default setup path for read-only database helpers:
// normalize the incoming context, apply the query timeout, and resolve the
// shared sqlx handle in one place so callers can immediately use GetContext or
// SelectContext without repeating boilerplate.
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
