package cache

// Cache defines the interface for reading and writing the most recent translations in a flat format.
type Cache interface {
	Read() (map[string]string, error)
	Write(obj map[string]string) error
}
