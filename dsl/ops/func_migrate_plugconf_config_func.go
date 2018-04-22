package ops

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/dsl/dslctx"
	"github.com/vim-volt/volt/dsl/ops/util"
	"github.com/vim-volt/volt/dsl/types"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/pathutil"
	"github.com/vim-volt/volt/plugconf"
	"github.com/vim-volt/volt/subcmd/builder"
)

func init() {
	opsMap[MigratePlugconfConfigFuncOp.String()] = MigratePlugconfConfigFuncOp
}

type migratePlugconfConfigFuncOp struct {
	funcBase
}

// MigratePlugconfConfigFuncOp is "migrate/plugconf/config-func" operator
var MigratePlugconfConfigFuncOp = &migratePlugconfConfigFuncOp{
	funcBase("migrate/plugconf/config-func"),
}

func (*migratePlugconfConfigFuncOp) Bind(args ...types.Value) (types.Expr, error) {
	if err := util.Signature().Check(args); err != nil {
		return nil, err
	}
	retType := types.VoidType
	return types.NewExpr(MigratePlugconfConfigFuncOp, args, retType), nil
}

func (*migratePlugconfConfigFuncOp) InvertExpr(_ context.Context, args []types.Value) (types.Value, error) {
	return MigratePlugconfConfigFuncOp.Bind(args...)
}

func (*migratePlugconfConfigFuncOp) EvalExpr(ctx context.Context, args []types.Value) (_ types.Value, rollback func(), result error) {
	rollback = NoRollback
	lockJSON := ctx.Value(dslctx.LockJSONKey).(*lockjson.LockJSON)
	cfg := ctx.Value(dslctx.ConfigKey).(*config.Config)

	parseResults, parseErr := plugconf.ParseMultiPlugconf(lockJSON.Repos)
	if parseErr.HasErrs() {
		var errMsg bytes.Buffer
		errMsg.WriteString("Please fix the following errors before migration:")
		for _, err := range parseErr.Errors().Errors {
			for _, line := range strings.Split(err.Error(), "\n") {
				errMsg.WriteString("  ")
				errMsg.WriteString(line)
			}
		}
		result = errors.New(errMsg.String())
		return
	}

	type plugInfo struct {
		path    string
		content []byte
	}
	infoList := make([]plugInfo, 0, len(lockJSON.Repos))

	// Collects plugconf infomations and check errors
	parseResults.Each(func(reposPath pathutil.ReposPath, info *plugconf.ParsedInfo) {
		if !info.ConvertConfigToOnLoadPreFunc() {
			return // no s:config() function
		}
		content, err := info.GeneratePlugconf()
		if err != nil {
			result = errors.Wrap(err, "could not generate converted plugconf")
			return
		}
		infoList = append(infoList, plugInfo{
			path:    reposPath.Plugconf(),
			content: content,
		})
	})
	if result != nil {
		return
	}

	// After checking errors, write the content to files
	for _, info := range infoList {
		os.MkdirAll(filepath.Dir(info.path), 0755)
		err := ioutil.WriteFile(info.path, info.content, 0644)
		if err != nil {
			result = errors.Wrapf(err, "could not write to file %s", info.path)
			return
		}
	}

	// Build ~/.vim/pack/volt dir
	result = builder.Build(false, lockJSON, cfg)
	if result != nil {
		result = errors.Wrap(result, "could not build "+pathutil.VimVoltDir())
	}
	return
}
