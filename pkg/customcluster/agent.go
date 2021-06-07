package customcluster

import (
	"cloudengine/pkg/types"
	"cloudengine/pkg/utils/clients"
	"fmt"
	"k8s.io/klog"
	"net/url"
	"sync"
	"time"
)

type Agent struct {
	cluster      types.ClusterStatus
	resources    []types.ResourceStatus
	result       types.CommandResult
	mux          sync.Mutex
	serverClient clients.HttpClient
}

func (a *Agent) Run(stopCh chan struct{}) error {
	klog.V(4).Infof("agent starting, heartbeat interval %ds", HeartbeatIntervalSeconds)
	ticker := time.NewTicker(time.Duration(HeartbeatIntervalSeconds) * time.Second)
MAINLOOP:
	for {
		select {
		case <-stopCh:
			break MAINLOOP
		case <-ticker.C:
			a.heartbeat()
		}
	}
	klog.V(4).Info("agent stopped")
	return nil
}

func (a *Agent) heartbeat() {
	var resources []types.ResourceStatus
	func() {
		a.mux.Lock()
		defer a.mux.Unlock()
		resources = a.resources[:]
		a.resources = a.resources[0:0]
	}()
	body := types.Heartbeat{
		Cluster:       a.cluster,
		Resources:     resources,
		CommandResult: &a.result,
		Time:          time.Now().Unix(),
	}
	// TODO
	_ = body
}

func (a *Agent) commandHandler(cmd types.Command) {

}

func (a *Agent) commandResultCollector(result types.CommandResult) {
	a.result = result
}

func NewAgent() (*Agent, error) {
	_, err := url.Parse(ApiServer)
	if err != nil {
		return nil, fmt.Errorf("api server config %s invalid: %s", ApiServer, err.Error())
	}

	cli := clients.NewDefaultHttpClient(ApiServer)
	agent := &Agent{
		serverClient: cli,
	}

	return agent, nil
}
