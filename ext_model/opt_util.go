package ext_model

// GetAs gets a value from ExtModel by key and converts it to target generic type.
// It returns false when the key is missing or the concrete type does not match T.
func GetAs[T ExtObj](model ExtModel, key string) (T, bool) {
	var zero T
	if model == nil {
		return zero, false
	}

	value, ok := model.Get(key)
	if !ok {
		return zero, false
	}

	typed, ok := value.(T)
	if !ok {
		return zero, false
	}
	return typed, true
}
