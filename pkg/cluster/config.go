package cluster

var (
	Host       string
	Port       int
	ApiServer  string
	AgentToken string
)

/*
	Cluster lifecycle
*/
var (
	HeartbeatIntervalSeconds = 3
	HeartbeatTimeoutSeconds  = 10
)
