package eventbus

type Topic string

const (
	CustomClusterInitTopic    Topic = "custom-cluster.lifecycle.init"
	CustomClusterDeletedTopic       = "custom-cluster.lifecycle.deleted"
)
