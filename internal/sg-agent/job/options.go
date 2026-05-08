package job

type (
	// Option -
	Option interface {
		isOption()
	}
	// WithNetNS -
	WithNetNS string
	// WithDefPolicyAccept -
	WithDefPolicyAccept bool
)

func (WithNetNS) isOption()           {}
func (WithDefPolicyAccept) isOption() {}
