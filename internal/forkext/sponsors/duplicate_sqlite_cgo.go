//go:build cgo

package sponsors

import (
	"errors"

	"github.com/mattn/go-sqlite3"
)

func isSQLiteSponsorDuplicateError(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique || sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey
	}
	var sqliteErrPtr *sqlite3.Error
	if errors.As(err, &sqliteErrPtr) && sqliteErrPtr != nil {
		return sqliteErrPtr.ExtendedCode == sqlite3.ErrConstraintUnique || sqliteErrPtr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey
	}
	return false
}
