package translate

// Translator defines the interface for converting text between different languages.
type Translator interface {
	Translate(source, fromLocal, toLocale string) (string, error)
}
