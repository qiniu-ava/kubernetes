package priorities

import (
	"fmt"

	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

const maxGPUPriority = 100

// LeastRemainedGPUPriorityMap prefer nodes with less remained GPUs if the pod is scheduled on.
// score = (100 - remained GPU after scheduled)
func LeastRemainedGPUPriorityMap(pod *v1.Pod, meta interface{}, nodeInfo *schedulercache.NodeInfo) (schedulerapi.HostPriority, error) {
	node := nodeInfo.Node()
	if node == nil {
		return schedulerapi.HostPriority{}, fmt.Errorf("node not found")
	}
	zeroPriority := schedulerapi.HostPriority{
		Host: node.Name,
	}
	// return fast
	var requestingGPU int64
	for _, container := range pod.Spec.Containers {
		requestingGPU += container.Resources.Limits.NvidiaGPU().Value()
	}
	if requestingGPU <= 0 {
		return zeroPriority, nil
	}

	if nodeInfo.AllocatableResource().NvidiaGPU == 0 {
		return zeroPriority, nil
	}
	capableGPU := nodeInfo.AllocatableResource().NvidiaGPU - nodeInfo.RequestedResource().NvidiaGPU
	if capableGPU <= 0 {
		return zeroPriority, nil
	}

	score := maxGPUPriority - int(capableGPU-requestingGPU)
	if score <= 1 {
		score = 1
	}

	glog.V(7).Infof("%v -> %v: Least Remained GPU Priority, allocatable %d, capable %d, requesting %d, score %d",
		pod.Name, node.Name, nodeInfo.AllocatableResource().NvidiaGPU, capableGPU, requestingGPU, score)

	return schedulerapi.HostPriority{
		Host:  node.Name,
		Score: score,
	}, nil
}
