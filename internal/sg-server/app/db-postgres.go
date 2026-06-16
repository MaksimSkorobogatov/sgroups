package app

import (
	"context"
	"net/url"

	appdb "github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
)

func newPostgresDB(ctx context.Context) (r appdb.Repository, err error) {
	var u string
	if u, err = PostgresURL.Value(ctx); err != nil {
		return nil, err
	}
	var dbURL *url.URL
	if dbURL, err = url.Parse(u); err != nil {
		return nil, err
	}
	r, err = appdb.NewRepositoryFromPG(ctx, *dbURL)
	return r, err
}
