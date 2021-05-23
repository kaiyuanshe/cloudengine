package types

type Heartbeat struct {
	Cluster       ClusterStatus    `json:"cluster"`
	Resources     []ResourceStatus `json:"resources,omitempty"`
	CommandResult *CommandResult   `json:"commandResult,omitempty"`
	Time          int64            `json:"time"`
}

type HeartbeatResponse struct {
	OK      bool          `json:"ok"`
	Message string        `json:"message"`
	Cluster ClusterStatus `json:"cluster"`
	Command *Command      `json:"command,omitempty"`
}
