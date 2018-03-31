package migrate

import (
	"errors"

	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/transaction"
)

func init() {
	migrateOps["lockjson"] = &lockjsonMigrater{}
}

type lockjsonMigrater struct{}

func (*lockjsonMigrater) Migrate() error {
	// Read lock.json
	lockJSON, err := lockjson.ReadNoMigrationMsg()
	if err != nil {
		return errors.New("could not read lock.json: " + err.Error())
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()

	// Write to lock.json
	err = lockJSON.Write()
	if err != nil {
		return errors.New("could not write to lock.json: " + err.Error())
	}
	return nil
}
