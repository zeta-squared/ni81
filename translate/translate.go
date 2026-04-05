package translate

// Translator defines a translation service capable of converting text
// from one locale to another.
type Translator interface {
	Translate(source, fromLocal, toLocale string) (string, error)
}
