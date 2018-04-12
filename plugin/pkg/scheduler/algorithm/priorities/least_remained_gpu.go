package priorities

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api/v1"
	schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// LeastRemainedGPUPriorityMap prefer nodes with less remained GPUs if the pod is scheduled on.
// score = (100 - remained GPU after scheduled)
func LeastRemainedGPUPriorityMap(pod *v1.Pod, meta interface{}, nodeInfo *schedulercache.NodeInfo) (schedulerapi.HostPriority, error) {
	node := nodeInfo.Node()
	if node == nil {
		return schedulerapi.HostPriority{}, fmt.Errorf("node not found")
	}
	noGPUPriority := schedulerapi.HostPriority{
		Host: node.Name,
	}
	// return fast
	if nodeInfo.AllocatableResource().NvidiaGPU == 0 {
		return noGPUPriority, nil
	}
	capableGPU := nodeInfo.AllocatableResource().NvidiaGPU - nodeInfo.RequestedResource().NvidiaGPU
	if capableGPU <= 0 {
		return noGPUPriority, nil
	}

	var limitedGPU int64
	for _, container := range pod.Spec.Containers {
		for rName, rQuantity := range container.Resources.Limits { // for GPU, only limits is required to be specified.
			switch rName {
			case v1.ResourceNvidiaGPU:
				limitedGPU += rQuantity.Value()
			}
		}
	}
	if limitedGPU == 0 {
		return noGPUPriority, nil
	}
	remained := capableGPU - limitedGPU
	if remained < 0 {
		return noGPUPriority, nil
	}
	if remained >= 100 {
		remained = 99
	}

	glog.V(7).Infof("%v -> %v: Least Remained GPU Priority, capacity %d, limits %d, remained %d, score %d",
		pod.Name, node.Name, capableGPU, limitedGPU, remained, 100-int(remained))

	return schedulerapi.HostPriority{
		Host:  node.Name,
		Score: 100 - int(remained),
	}, nil
}
