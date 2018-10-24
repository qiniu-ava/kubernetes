package priorities

import (
	"fmt"

	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

const maxGPUPriority = 100
const nvidiaGPUResourceName v1.ResourceName = "nvidia.com/gpu"

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

	nLimits := podGPU(pod)
	nAvailable := nodeGPU(nodeInfo)
	if nLimits <= 0 {
		return zeroPriority, nil
	}
	if nAvailable <= 0 || nAvailable < nLimits {
		return zeroPriority, nil
	}

	score := maxGPUPriority - int(nAvailable-nLimits)
	if score <= 1 {
		score = 1
	}

	glog.V(7).Infof("%v -> %v: Least Remained GPU Priority, available %d, requesting %d, score %d",
		pod.Name, node.Name, nodeInfo.AllocatableResource().NvidiaGPU, nAvailable, nLimits, score)

	return schedulerapi.HostPriority{
		Host:  node.Name,
		Score: score,
	}, nil
}

func podGPU(pod *v1.Pod) int64 {
	var nGPU, nInit int64
	for _, c := range pod.Spec.Containers {
		nGPU += containerGPU(&c)
	}
	for _, c := range pod.Spec.InitContainers {
		nInit += containerGPU(&c)
	}
	if nInit > nGPU {
		return nInit
	}
	return nGPU
}

func containerGPU(c *v1.Container) int64 {
	if res, ok := c.Resources.Limits[nvidiaGPUResourceName]; ok {
		return res.Value()
	}
	return 0
}

func nodeGPU(ni *schedulercache.NodeInfo) int64 {
	if alloc, ok := ni.AllocatableResource().ScalarResources[nvidiaGPUResourceName]; ok {
		if req, ok := ni.RequestedResource().ScalarResources[nvidiaGPUResourceName]; ok {
			if req <= alloc {
				return alloc - req
			}
		}
	}
	return 0
}
