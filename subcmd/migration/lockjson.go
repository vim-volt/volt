package migration

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/dsl"
	"github.com/vim-volt/volt/dsl/dslctx"
	"github.com/vim-volt/volt/dsl/ops"
	"github.com/vim-volt/volt/lockjson"
)

func init() {
	m := &lockjsonMigrater{}
	migrateOps[m.Name()] = m
}

type lockjsonMigrater struct{}

func (*lockjsonMigrater) Name() string {
	return "lockjson"
}

func (m *lockjsonMigrater) Description(brief bool) string {
	if brief {
		return "converts old lock.json format to the latest format"
	}
	return `Usage
  volt migrate [-help] ` + m.Name() + `

Description
  Perform migration of $VOLTPATH/lock.json, which means volt converts old version lock.json structure into the latest version. This is always done automatically when reading lock.json content. For example, 'volt get <repos>' will install plugin, and migrate lock.json structure, and write it to lock.json after all. so the migrated content is written to lock.json automatically.
  But, for example, 'volt list' does not write to lock.json but does read, so every time when running 'volt list' shows warning about lock.json is old.
  To suppress this, running this command simply reads and writes migrated structure to lock.json.`
}

func (*lockjsonMigrater) Migrate(lockJSON *lockjson.LockJSON, cfg *config.Config) error {
	ctx := dslctx.WithDSLValues(context.Background(), lockJSON, cfg)
	expr, err := ops.LockJSONWriteOp.Bind()
	if err != nil {
		return errors.Wrapf(err, "cannot bind %s operator", ops.LockJSONWriteOp.String())
	}
	_, err = dsl.Execute(ctx, expr)
	return err
}
