package api

import "errors"

// MigrateKeyValueDB copies all key/value pairs from src into dst.
// It stops at the first dst.Set error and returns that error.
func MigrateKeyValueDB(src KeyValueDB, dst KeyValueDB) error {
	if src == nil {
		return errors.New("MigrateKeyValueDB: src is nil")
	}
	if dst == nil {
		return errors.New("MigrateKeyValueDB: dst is nil")
	}

	var setErr error
	iterErr := src.Iterate(func(key, value string) bool {
		if err := dst.Set(key, value); err != nil {
			setErr = err
			return false
		}
		return true
	})
	if iterErr != nil {
		return iterErr
	}
	return setErr
}
