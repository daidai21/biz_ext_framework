package ext_model

type CopyExtMapOption[V ExtObj] func(*copyExtMapOptions[V])

type copyExtMapOptions[V ExtObj] struct {
	deepCopy  func(V) V
	keyFilter func(string) bool
}

func WithDeepCopy[V ExtObj](fn func(V) V) CopyExtMapOption[V] {
	return func(options *copyExtMapOptions[V]) {
		options.deepCopy = fn
	}
}

func WithKeyFilter[V ExtObj](fn func(string) bool) CopyExtMapOption[V] {
	return func(options *copyExtMapOptions[V]) {
		options.keyFilter = fn
	}
}

func CopyExtMap[V ExtObj](src ExtModel[V], opts ...CopyExtMapOption[V]) *ExtMap[V] {
	var dst ExtMap[V]
	if src == nil {
		return &dst
	}

	options := copyExtMapOptions[V]{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	src.ForEach(func(value V) {
		key := value.Key()
		if options.keyFilter != nil && !options.keyFilter(key) {
			return
		}
		if options.deepCopy != nil {
			value = options.deepCopy(value)
		}
		dst.Set(value)
	})

	return &dst
}
