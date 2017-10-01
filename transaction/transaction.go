package transaction

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
)

// Create $VOLTPATH/trx.lock file
func Create() error {
	ownPid := []byte(strconv.Itoa(os.Getpid()))
	trxLockFile := pathutil.TrxLock()

	// Create trx.lock parent directories
	err := os.MkdirAll(filepath.Dir(trxLockFile), 0755)
	if err != nil {
		return err
	}

	// Write pid to trx.lock file
	err = ioutil.WriteFile(trxLockFile, ownPid, 0644)
	if err != nil {
		return err
	}

	// Read pid from trx.lock file
	pid, err := ioutil.ReadFile(trxLockFile)
	if err != nil {
		return err
	}

	if string(pid) != string(ownPid) {
		return errors.New("transaction lock was taken by PID " + string(pid))
	}
	return nil
}

// Remove $VOLTPATH/trx.lock file
func Remove() {
	// Read pid from trx.lock file
	trxLockFile := pathutil.TrxLock()
	pid, err := ioutil.ReadFile(trxLockFile)
	if err != nil {
		logger.Error("trx.lock was already removed")
		return
	}

	// Remove trx.lock if pid is same
	if string(pid) != strconv.Itoa(os.Getpid()) {
		logger.Error("Cannot remove another process's trx.lock")
		return
	}
	os.Remove(trxLockFile)
}
