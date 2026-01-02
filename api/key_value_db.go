package api

// KeyValueDB is a simple string key/value database.
type KeyValueDB interface {
	Get(key string) (value string, ok bool, err error)
	Set(key, value string) error
	Delete(key string) error
	Iterate(fn func(key, value string) bool) error
	Close() error
}
