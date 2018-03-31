package migrate

import (
	"errors"

	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/transaction"
)

func init() {
	m := &lockjsonMigrater{}
	migrateOps[m.Name()] = m
}

type lockjsonMigrater struct{}

func (*lockjsonMigrater) Name() string {
	return "lockjson"
}

func (*lockjsonMigrater) Description() string {
	return "converts old lock.json format to the latest format"
}

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
