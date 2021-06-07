package customcluster

var (
	Host           string
	Port           int
	ApiServer      string
	AgentToken     string
	ClusterID      string
	ControllerMode bool
)

/*
	Cluster lifecycle
*/
var (
	HeartbeatIntervalSeconds = 3
	HeartbeatTimeoutSeconds  = 10
)
