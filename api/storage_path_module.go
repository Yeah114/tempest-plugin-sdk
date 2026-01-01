package api

const NameStoragePathModule = "storage_path"

type StoragePathModule interface {
	Name() string

	ConfigPath(parts ...string) string
	CodePath(parts ...string) string
	DataFilePath(parts ...string) string
	CachePath(parts ...string) string
}
