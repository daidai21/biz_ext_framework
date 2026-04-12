package ext_model

type CopyExtMapOption func(*copyExtMapOptions)

type copyExtMapOptions struct {
	deepCopy  func(ExtObj) ExtObj
	keyFilter func(string) bool
}

func WithDeepCopy(fn func(ExtObj) ExtObj) CopyExtMapOption {
	return func(options *copyExtMapOptions) {
		options.deepCopy = fn
	}
}

func WithKeyFilter(fn func(string) bool) CopyExtMapOption {
	return func(options *copyExtMapOptions) {
		options.keyFilter = fn
	}
}

func CopyExtMap(src ExtModel, opts ...CopyExtMapOption) *ExtMap[ExtObj] {
	var dst ExtMap[ExtObj]
	if src == nil {
		return &dst
	}

	options := copyExtMapOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	src.ForEach(func(value ExtObj) {
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
