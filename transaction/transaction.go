package transaction

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"github.com/vim-volt/volt/pathutil"
)

// Start creates $VOLTPATH/trx/lock directory
func Start() (*Transaction, error) {
	os.MkdirAll(pathutil.TrxDir(), 0755)
	lockDir := filepath.Join(pathutil.TrxDir(), "lock")
	if err := os.Mkdir(lockDir, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction: "+lockDir+" exists: if no other volt process is currently running, this probably means a volt process crashed earlier. Make sure no other volt process is running and remove the file manually to continue")
	}
	trxID, err := genNewTrxID()
	if err != nil {
		return nil, errors.Wrap(err, "could not allocate a new transaction ID")
	}
	return &Transaction{trxID: trxID}, nil
}

// genNewTrxID gets unallocated transaction ID looking $VOLTPATH/trx/ directory
func genNewTrxID() (_ TrxID, result error) {
	trxDir, err := os.Open(pathutil.TrxDir())
	if err != nil {
		return nil, errors.Wrap(err, "could not open $VOLTPATH/trx directory")
	}
	defer func() { result = trxDir.Close() }()
	names, err := trxDir.Readdirnames(0)
	if err != nil {
		return nil, errors.Wrap(err, "could not readdir of $VOLTPATH/trx directory")
	}
	var maxID TrxID
	for i := range names {
		if !isTrxDirName(names[i]) {
			continue
		}
		if maxID == nil {
			maxID = TrxID(names[i])
			continue
		}
		if greaterThan(names[i], string(maxID)) {
			maxID = TrxID(names[i])
		}
	}
	if maxID == nil {
		return TrxID("1"), nil // no transaction ID directory
	}
	return maxID.Add(1)
}

func greaterThan(a, b string) bool {
	d := len(a) - len(b)
	if d > 0 {
		b = strings.Repeat("0", d) + b
	} else if d < 0 {
		a = strings.Repeat("0", -d) + a
	}
	return strings.Compare(a, b) > 0
}

func isTrxDirName(name string) bool {
	for _, r := range name {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// TrxID is a transaction ID, which is a serial number and directory name of
// transaction log file.
type TrxID []byte

// Add adds n to transaction ID
func (tid *TrxID) Add(n int) (TrxID, error) {
	newID, err := strconv.ParseUint(string(*tid), 10, 32)
	if err != nil {
		return nil, err
	}
	if newID+uint64(n) < newID {
		// TODO: compute in string?
		return nil, errors.Errorf("%d + %d causes overflow", newID, n)
	}
	return TrxID(strconv.FormatUint(newID+uint64(n), 10)), nil
}

// Transaction provides transaction methods
type Transaction struct {
	trxID TrxID
}

// Done rename "lock" directory to "{trxid}" directory
func (trx *Transaction) Done() error {
	lockDir := filepath.Join(pathutil.TrxDir(), "lock")
	trxIDDir := filepath.Join(pathutil.TrxDir(), string(trx.trxID))
	return os.Rename(lockDir, trxIDDir)
}
