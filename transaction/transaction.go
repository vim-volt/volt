package transaction

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
)

// Create creates $VOLTPATH/trx.lock file
func Create() error {
	ownPid := []byte(strconv.Itoa(os.Getpid()))
	trxLockFile := pathutil.TrxLock()

	// Create trx.lock parent directories
	err := os.MkdirAll(filepath.Dir(trxLockFile), 0755)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	// Return error if the file exists
	if pathutil.Exists(trxLockFile) {
		return errors.New("failed to begin transaction: " + pathutil.TrxLock() + " exists: if no other volt process is currently running, this probably means a volt process crashed earlier. Make sure no other volt process is running and remove the file manually to continue")
	}

	// Write pid to trx.lock file
	err = ioutil.WriteFile(trxLockFile, ownPid, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	// Read pid from trx.lock file
	pid, err := ioutil.ReadFile(trxLockFile)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	if string(pid) != string(ownPid) {
		return errors.New("transaction lock was taken by PID " + string(pid))
	}
	return nil
}

// Remove removes $VOLTPATH/trx.lock file
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
	err = os.Remove(trxLockFile)
	if err != nil {
		logger.Error("Cannot remove trx.lock: " + err.Error())
		return
	}
}
