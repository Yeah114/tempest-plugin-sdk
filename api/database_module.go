package api

const NameDatabaseModule = "database"

const (
	DBTypeDefault = ""
	DBTypeTextLog = "text_log"
	DBTypeLevel   = "level"
	DBTypeJSON    = "json"
)

// DatabaseModule provides access to persistent key-value databases.
type DatabaseModule interface {
	Name() string

	// KeyValueDB opens (or creates) a database at the given logical name.
	//
	// dbType can be one of:
	// - "" / "text_log": text log backend (human readable, append-only log)
	// - "level": leveldb backend
	// - "json": json file backend
	KeyValueDB(name string, dbType string) (KeyValueDB, error)
}
