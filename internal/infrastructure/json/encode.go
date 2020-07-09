package json

// MarshallCallback represents function type for json marshaling.
type MarshallCallback func(interface{}) ([]byte, error)
