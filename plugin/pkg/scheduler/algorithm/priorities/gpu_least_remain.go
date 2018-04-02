package priorities

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api/v1"
	schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// GPULeastRemainPriorityMap favors node with less available gpu resource remains.
// Calualation formula of socre as belows:
//	100 - (total allocated gpu num in this node + requested gpu num in current pod)
func GPULeastRemainPriorityMap(pod *v1.Pod, meta interface{}, nodeInfo *schedulercache.NodeInfo) (schedulerapi.HostPriority, error) {
	node := nodeInfo.Node()
	if node == nil {
		return schedulerapi.HostPriority{}, fmt.Errorf("node not found")
	}

	var requestGPU int64 = 0
	for _, container := range pod.Spec.Containers {
		for rName, rQuantity := range container.Resources.Requests {
			switch rName {
			case v1.ResourceNvidiaGPU:
				requestGPU += rQuantity.Value()
			}
		}
	}
	allocatedGPU := nodeInfo.RequestedResource().NvidiaGPU
	allocatableGPU := nodeInfo.AllocatableResource().NvidiaGPU // same as "total" gpu in this node

	return schedulerapi.HostPriority{
		Host:  node.Name,
		Score: getGPULeastRemainScore(allocatableGPU, allocatedGPU, requestGPU),
	}, nil
}

func getGPULeastRemainScore(allocatableGPU, allocatedGPU, requestGPU int64) (score int) {
	defer func() {
		glog.V(2).Infof("allocatable: %d, allocated: %d, request: %d, GPULeastReaminScore: %d",
			allocatableGPU, allocatedGPU, requestGPU, score)
	}()

	if requestGPU == 0 {
		score = 0
	} else {
		remainGPUAfter := int(allocatableGPU - allocatedGPU - requestGPU)
		if remainGPUAfter < 0 {
			score = 0
		} else {
			score = 100 - remainGPUAfter // mayby there is node with more than 100 GPU? TODO
		}
	}

	return score
}
