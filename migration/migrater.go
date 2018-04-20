package migration

import (
	"errors"
	"sort"

	"github.com/vim-volt/volt/config"
	"github.com/vim-volt/volt/lockjson"
)

// Migrater migrates many kinds of data.
type Migrater interface {
	Migrate(*lockjson.LockJSON, *config.Config) error
	Name() string
	Description(brief bool) string
}

var migrateOps = make(map[string]Migrater)

// GetMigrater gets Migrater of specified name.
func GetMigrater(name string) (Migrater, error) {
	m, exists := migrateOps[name]
	if !exists {
		return nil, errors.New("no such migration operation: " + name)
	}
	return m, nil
}

// ListMigraters lists all migraters.
func ListMigraters() []Migrater {
	migraters := make([]Migrater, 0, len(migrateOps))
	for _, m := range migrateOps {
		migraters = append(migraters, m)
	}
	sort.Slice(migraters, func(i, j int) bool {
		return migraters[i].Name() < migraters[j].Name()
	})
	return migraters
}
