package validators

import "buf.build/go/protovalidate"

// ProtoValidateOption configures protovalidate interceptors.
type ProtoValidateOption func(*protoValidateCfg)

type protoValidateCfg struct {
	v    protovalidate.Validator
	opts []protovalidate.ValidationOption
}

// WithValidator sets a custom Validator instance (constructed via protovalidate.New(...)).
// If not set, the global validator (protovalidate.Validate) is used.
func WithValidator(v protovalidate.Validator) ProtoValidateOption {
	return func(c *protoValidateCfg) {
		c.v = v
	}
}

// WithValidationOptions appends per-call validation options.
func WithValidationOptions(opts ...protovalidate.ValidationOption) ProtoValidateOption {
	return func(c *protoValidateCfg) {
		c.opts = append(c.opts, opts...)
	}
}
