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
	m := &plugconfConfigMigrater{}
	migrateOps[m.Name()] = m
}

type plugconfConfigMigrater struct{}

func (*plugconfConfigMigrater) Name() string {
	return "plugconf/config-func"
}

func (m *plugconfConfigMigrater) Description(brief bool) string {
	if brief {
		return "converts s:config() function name to s:on_load_pre() in all plugconf files"
	}
	return `Usage
  volt migrate [-help] ` + m.Name() + `

Description
  Perform migration of the function name of s:config() functions in plugconf files of all plugins. All s:config() functions are renamed to s:on_load_pre().
  "s:config()" is a old function name (see https://github.com/vim-volt/volt/issues/196).
  All plugconf files are replaced with new contents.`
}

func (*plugconfConfigMigrater) Migrate(lockJSON *lockjson.LockJSON, cfg *config.Config) (result error) {
	ctx := dslctx.WithDSLValues(context.Background(), lockJSON, cfg)
	expr, err := ops.MigratePlugconfConfigFuncOp.Bind()
	if err != nil {
		return errors.Wrapf(err, "cannot bind %s operator", ops.LockJSONWriteOp.String())
	}
	_, err = dsl.Execute(ctx, expr)
	return err
}
