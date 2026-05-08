package service

// Option configures.
type Option interface {
	apply(*CommonOptions)
}

// CommonOptions holds common options for service configuration.
type CommonOptions struct {
	PathPrefixes           []string
	AdditionalServiceNames []string
}

// ApplyOptions applies options to CommonOptions.
func (c *CommonOptions) ApplyOptions(opts ...Option) {
	for _, o := range opts {
		o.apply(c)
	}
}

type optionFunc func(*CommonOptions)

func (f optionFunc) apply(opts *CommonOptions) {
	f(opts)
}

// WithAPIpathPrefixes adds additional API path prefixes.
func WithAPIpathPrefixes(pp ...string) Option {
	return optionFunc(func(ss *CommonOptions) {
		ss.PathPrefixes = append(ss.PathPrefixes, pp...)
	})
}

// WithAdditionalServiceNames -
func WithAdditionalServiceNames(names ...string) Option {
	return optionFunc(func(ss *CommonOptions) {
		ss.AdditionalServiceNames = append(ss.AdditionalServiceNames, names...)
	})
}
