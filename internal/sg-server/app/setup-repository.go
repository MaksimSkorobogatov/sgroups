package sgserver

import (
	"context"
	"strings"

	appdb "github.com/PRO-Robotech/sgroups/internal/sg-server/repository"
	"github.com/PRO-Robotech/sgroups/internal/shared/app"

	"github.com/H-BF/corlib/pkg/atomic"
	"github.com/pkg/errors"
)

type (
	registryConstrutor func(context.Context) (appdb.Repository, error)
)

func getAppRepository() appdb.Repository {
	var ret appdb.Repository
	if !storedAppRepo.Fetch(func(v appdb.Repository) { ret = v }) {
		panic(errors.New("need setup db repository"))
	}
	return ret
}

var (
	storedAppRepo atomic.Value[appdb.Repository]

	repositoryConstructors = map[string]registryConstrutor{
		"postgres": newPostgresDB,
	}
)

// SetupRepository - setup registry storage
func SetupRepository() error {
	ctx := app.Context()
	st, err := StorageType.Value(ctx)
	if err != nil {
		return err
	}
	f, ok := repositoryConstructors[strings.ToLower(strings.TrimSpace(st))]
	if !ok {
		return errors.Errorf("unknown repository storage type '%s'", st)
	}
	var db appdb.Repository
	if db, err = f(ctx); err != nil {
		return err
	}
	storedAppRepo.Store(db, func(old appdb.Repository) {
		_ = old.Close()
	})
	return nil
}

var _ = SetupRepository
