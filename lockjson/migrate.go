package lockjson

import (
	"encoding/json"

	"github.com/vim-volt/volt/logger"
)

func migrate(rawJSON []byte, lockJSON *LockJSON) error {
	// lockJSON.Version must be greater than 0 because it was validated
	var err error
	max := int64(len(migrateFunc))
	for lockJSON.Version-1 < max {
		logger.Infof("Migrating lock.json v%d to v%d ...", lockJSON.Version, lockJSON.Version+1)
		err = migrateFunc[lockJSON.Version-1](rawJSON, lockJSON)
		if err != nil {
			return err
		}
	}
	return nil
}

var migrateFunc = []func([]byte, *LockJSON) error{
	migrate1To2,
}

// Rename 'active_profile' to 'current_profile_name'
func migrate1To2(rawJSON []byte, lockJSON *LockJSON) error {
	var j struct {
		ActiveProfile string `json:"active_profile"`
	}

	err := json.Unmarshal(rawJSON, &j)
	if err != nil {
		return err
	}
	lockJSON.CurrentProfileName = j.ActiveProfile
	lockJSON.Version++

	return nil
}
