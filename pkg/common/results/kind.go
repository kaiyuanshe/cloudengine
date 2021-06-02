package results

type resultKind int

const (
	noRequeue resultKind = iota
	delayRequeue
	normalRequeue
)
