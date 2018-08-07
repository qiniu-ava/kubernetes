package priorities

import (
	"fmt"

	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

const maxGPUPriority = 100
const gpuResourceName = "nvidia.com/gpu"

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
	nGPU := podRequestsGPU(pod)
	if nGPU <= 0 {
		return zeroPriority, nil
	}

	allocatable, requested := nodeGPU(nodeInfo)
	if allocatable == 0 {
		return zeroPriority, nil
	}
	availableGPU := allocatable - requested
	if availableGPU <= 0 || availableGPU < nGPU {
		return zeroPriority, nil
	}

	score := maxGPUPriority - int(availableGPU-nGPU)
	if score <= 1 {
		score = 1
	}

	glog.V(7).Infof("%v -> %v: Least Remained GPU Priority, allocatable %d, available %d, requesting %d, score %d",
		pod.Name, node.Name, nodeInfo.AllocatableResource().NvidiaGPU, availableGPU, nGPU, score)

	return schedulerapi.HostPriority{
		Host:  node.Name,
		Score: score,
	}, nil
}

func podRequestsGPU(pod *v1.Pod) int64 {
	var res int64
	for _, c := range pod.Spec.Containers {
		q := c.Resources.Limits[gpuResourceName]
		res += q.Value()
	}
	return res
}

func nodeGPU(nodeInfo *schedulercache.NodeInfo) (allocatable, requested int64) {
	allocatable = nodeInfo.AllocatableResource().ScalarResources[gpuResourceName]
	for _, p := range nodeInfo.Pods() {
		requested += podRequestsGPU(p)
	}
	return
}
