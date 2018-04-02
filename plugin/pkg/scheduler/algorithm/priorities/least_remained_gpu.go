package priorities

import (
	"fmt"
	"sort"

	"k8s.io/kubernetes/pkg/api/v1"
	schedulerapi "k8s.io/kubernetes/plugin/pkg/scheduler/api"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

// LeastRemainedGPUPriorityMap favors node with less available gpu resource remains.
// Score of a node equals to the number of remained GPUs if pod is scheduled to this node:
// 'score = (allocatableGPU - allocatedGPU - limitedGPU)'
//
// Notice: Scores should be in the range of [0, 10] to collaborate with other schedulers,
// and a higher priority should get a higher final score,
// so we will need a reducer to reverses the scores and constrains them to [0, 10]
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
	if nodeInfo.AllocatableResource().NvidiaGPU-nodeInfo.RequestedResource().NvidiaGPU == 0 {
		return noGPUPriority, nil
	}

	var limitedGPU int64
	for _, container := range pod.Spec.Containers {
		for rName, rQuantity := range container.Resources.Limits { // for GPUs, only limits is required to be specified.
			switch rName {
			case v1.ResourceNvidiaGPU:
				limitedGPU += rQuantity.Value()
			}
		}
	}
	if limitedGPU == 0 {
		return noGPUPriority, nil
	}

	allocatedGPU := nodeInfo.RequestedResource().NvidiaGPU
	allocatableGPU := nodeInfo.AllocatableResource().NvidiaGPU // same as "total" gpu in this node
	remained := allocatableGPU - allocatedGPU - limitedGPU
	if remained < 0 {
		glog.V(10).Infof("Combined requested %d GPUs from existing pods exceeds capacity %d on node %s",
			limitedGPU, allocatableGPU, node)
	}

	return schedulerapi.HostPriority{
		Host:  node.Name,
		Score: int(remained),
	}, nil
}

// LeastRemainedGPUPriorityReduce aggregates remained GPU numbers of all possible nodes,
// and reassign scores according the order of this number among all numbers.
//
// say, we have 6 GPU nodes in our cluster, each of them will remian [0, 0, 2, 3, 3, 7] GPU if some pod is scheduled on.
// and the scores of these nodes to this pods are [10, 10, 9, 8, 8, 7]
func LeastRemainedGPUPriorityReduce(pod *v1.Pod, meta interface{}, nodeNameToInfo map[string]*schedulercache.NodeInfo, result schedulerapi.HostPriorityList) error {
	scoreMap := make(map[int]int) // remained GPU -> priority score
	var hasInavailabelNodes bool
	for i, r := range result {
		if r.Score < 0 {
			result[i].Score = -1
			hasInavailabelNodes = true
		}
		scoreMap[result[i].Score] = 0
	}
	if len(scoreMap) <= 1 { // if all nodes have same score (e.g. pod does not use GPU), return fast
		for i := range result {
			glog.Infof("%v -> %v: LeastRemainedGPUPriority, Score: (%d)", pod.Name, result[i].Host, 0)
			result[i].Score = 0
		}
		return nil
	}

	numbers := make([]int, 0, len(scoreMap))
	for number := range scoreMap {
		numbers = append(numbers, number)
	}
	sort.Ints(numbers)

	// final score = 10 - (order of remained number among all appeared remained numbers)
	for order, number := range numbers {
		if number < 0 {
			scoreMap[number] = 0
		}
		if order >= 10 {
			scoreMap[number] = 1
		}
		if hasInavailabelNodes {
			scoreMap[number] = 10 + 1 - order
		} else {
			scoreMap[number] = 10 - order
		}
	}

	for i, r := range result {
		result[i].Score = scoreMap[r.Score]
		if glog.V(10) {
			// We explicitly don't do glog.V(10).Infof() to avoid computing all the parameters if this is
			// not logged. There is visible performance gain from it.
			glog.Infof("%v -> %v: LeastRemainedGPUPriority, Score: (%d)", pod.Name, result[i].Host, result[i].Score)
		}
	}

	return nil
}
