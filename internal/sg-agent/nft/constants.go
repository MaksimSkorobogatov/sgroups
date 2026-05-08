package nft

const (
	chnIngress     = "INGRESS"
	chnEgress      = "EGRESS"
	chnEgressMain  = chnEgress + "-MAIN"
	chnIngressMain = chnIngress + "-MAIN"
)

const (
	dirIN  direction = true
	dirOUT direction = false

	mainTablePrefix = "main"
)

type direction bool
