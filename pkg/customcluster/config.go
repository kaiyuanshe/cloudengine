package customcluster

var (
	Host           string
	Port           int
	ApiServer      string
	AgentToken     string
	ClusterID      string
	ControllerMode bool
	AgentMode      bool
)

/*
	Cluster lifecycle
*/
var (
	HeartbeatIntervalSeconds = 10
	HeartbeatTimeoutSeconds  = 30
)
