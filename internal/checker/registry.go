package checker

// Factory creates a new Detector instance.
type Factory func() Detector

var registry = map[string]Factory{}

// Register adds a detector factory for the given errorType.
func Register(errorType string, f Factory) {
	registry[errorType] = f
}

// Lookup returns a detector for the given errorType.
func Lookup(errorType string) (Detector, bool) {
	f, ok := registry[errorType]
	if !ok {
		return nil, false
	}
	return f(), true
}
