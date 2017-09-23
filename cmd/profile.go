package cmd

import (
	"fmt"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/transaction"
)

type profileCmd struct{}

func Profile(args []string) int {
	cmd := profileCmd{}
	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		fmt.Println("[ERROR] Failed to read lock.json: " + err.Error())
		return 10
	}

	if len(args) == 0 {
		// Show profile name
		fmt.Println(lockJSON.ActiveProfile)
	} else {
		// Set profile name
		err = cmd.setProfile(lockJSON, args[0])
		if err != nil {
			fmt.Println("[ERROR] Failed to set active profile: " + err.Error())
			return 11
		}
		fmt.Println("[INFO] Set active profile to '" + lockJSON.ActiveProfile + "'")
	}

	return 0
}

func (profileCmd) setProfile(lockJSON *lockjson.LockJSON, name string) error {
	// Begin transaction
	err := transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()
	lockJSON.TrxID++
	lockJSON.ActiveProfile = name

	// Write to lock.json
	err = lockjson.Write(lockJSON)
	if err != nil {
		return err
	}

	return nil
}
