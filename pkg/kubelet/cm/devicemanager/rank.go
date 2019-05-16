/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package devicemanager

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
)

const (
	MAXCOST = 10000
)

var costs [][]int = costs_()

func costs_() [][]int { // communication cost between gpus
	return [][]int{
		{0, 1, 10, 10, 100, 100, 100, 100},
		{1, 0, 10, 10, 100, 100, 100, 100},
		{10, 10, 0, 1, 100, 100, 100, 100},
		{10, 10, 1, 0, 100, 100, 100, 100},
		{100, 100, 100, 100, 0, 1, 10, 10},
		{100, 100, 100, 100, 1, 0, 10, 10},
		{100, 100, 100, 100, 10, 10, 0, 1},
		{100, 100, 100, 100, 10, 10, 1, 0},
	}
}

// SM align to exponential of 2
func align2(need, max int) int {
	for need < max {
		max = max / 2
	}
	return max
}

type node struct {
	parent *node
	left   *node
	right  *node
	bm     []bool // globally shared, updated while ranking
	used   []bool // globally shared, currently allocated or pre-installed resource, updated while ranking
	start  int    // [start index in bm of this node
	end    int    // end) index in bm of this node
}

func mkNode(parent *node, bm []bool, used []bool, start, end int) *node {
	if start >= end {
		return nil
	}
	var n *node = &node{parent: parent, bm: bm, used: used, start: start, end: end}
	if end > start+1 {
		middle := (end + start) / 2
		n.left = mkNode(n, bm, used, start, middle)
		n.right = mkNode(n, bm, used, middle, end)
	}
	return n
}

func (n *node) length() int {
	return n.end - n.start
}

// available resource in this node
func (n *node) available() int {
	c := 0
	for i := n.start; i < n.end; i++ {
		if n.bm[i] {
			c++
		}
	}
	return c
}

// sum of communication cost between all available resources in this node,
// including the one just allocated during ranking
func (n *node) cost() (int, []int) {
	idx, c := make([]int, 0), 0
	gbm := make([]bool, len(n.bm))
	for i := 0; i < len(n.bm); i++ { // include just allocated resources
		if n.used[i] || (i >= n.start && i < n.end && n.bm[i]) {
			gbm[i] = true
		}
	}
	for i := 0; i < len(gbm); i++ {
		if gbm[i] {
			for j := i + 1; j < len(gbm); j++ { // cost include nodes just allocated
				if gbm[j] {
					c += costs[i][j]
				}
			}
			if i >= n.start && i < n.end { // only count in this node
				idx = append(idx, i)
			}
		}
	}
	return c, idx
}

// step1: if only have *num* resources, rank this node and return
// step2: rank left & right child, select which one having less cost
// step3: in case left cost = right cost, checking parent's cost recursively
// step4: left & right can not satisfy, go back and come again with half the resources
func doRanking(n *node, num int) (int, []int, *node) {
	if n == nil || num > n.available() {
		return MAXCOST, nil, nil
	}

	if n.available() == num && n.length() == num { // all available
		c, cs := n.cost()
		return c, cs, n
	}

	lcost, lidx, ln := doRanking(n.left, num)
	rcost, ridx, rn := doRanking(n.right, num)
	klog.V(3).Infoln("SMAFFINITY ranking left", num, lcost, lidx, ln)
	klog.V(3).Infoln("SMAFFINITY ranking right", num, rcost, ridx, rn)

	if ln == nil && rn == nil { // neither satisfied
		return MAXCOST, nil, nil
	} else if ln != nil && rn != nil && ln != rn { // both satisfied
		if lcost == rcost { // check parent recursively
			for true {
				ln, rn = ln.parent, rn.parent
				if ln != nil && rn != nil && ln != rn {
					lcost, _ = ln.cost()
					rcost, _ = rn.cost()
					klog.V(3).Infoln("SMAFFINITY ranking up", lcost, rcost, ln, rn)
					if lcost == rcost { // continue to loop to parents
						continue
					} else if lcost < rcost { // left parent has smaller cost
						return lcost, lidx, ln
					} else { // right parent has smaller cost
						return rcost, ridx, rn
					}
				} else if ln != nil { // prefer left one
					return lcost, lidx, ln
				} else {
					return rcost, ridx, rn
				}
			}
			return 0, nil, nil // never reach, keep compiler happy
		} else if lcost < rcost {
			return lcost, lidx, ln
		} else {
			return rcost, ridx, rn
		}
	} else if ln != nil { // left satisfied, also prefer left one
		return lcost, lidx, ln
	} else { // right satisfied
		return rcost, ridx, rn
	}
}

// break into integral parts and do ranking separately.
// for example, for needed=7, rank 4 + 2 + 1 separately
func rank(root *node, needed int) bool {
	if root.available() < needed {
		klog.Error("SMAFFINITY wrong: no resource for", needed, root)
		return false
	}
	num := align2(needed, root.end-root.start)
	for needed > 0 {
		if num < 1 { // should never get here
			klog.Error("SMAFFINITY badly wrong ", needed, root)
			return false
		}
		cost, idx, node := doRanking(root, num)
		if idx == nil {
			klog.V(3).Infoln("SMAFFINITY will cont with half for", num)
			num = num / 2
			continue // half again
		}
		for _, v := range idx { // remember just allocated resource
			root.bm[v] = false
			root.used[v] = true
		}
		klog.V(3).Infoln("SMAFFINITY satisfied", num, "cost:", cost, "idx:", idx, "node:", node)
		needed = needed - num
		num = align2(needed, root.end-root.start)
	}
	return true
}

func (m *ManagerImpl) calcAllocated(resource string, needed int, inuse, available sets.String) []string {
	if resource != "nvidia.com/gpu" || len(m.gpuBitmap) == 0 {
		return available.UnsortedList()[:needed]
	}

	bm := make([]bool, len(m.gpuBitmap))
	for id, _ := range available {
		if _, ok := m.gpuBitmap[id]; ok {
			bm[m.gpuBitmap[id]] = true
		} else {
			klog.Error("SMAFFINITY not bitmap for available", resource, id)
			return available.UnsortedList()[:needed]
		}
	}

	used := make([]bool, len(m.gpuBitmap))
	for id, _ := range inuse {
		if _, ok := m.gpuBitmap[id]; ok {
			used[m.gpuBitmap[id]] = true
		} else {
			klog.Error("SMAFFINITY not bitmap for available", resource, id)
			return available.UnsortedList()[:needed]
		}
	}
	klog.V(3).Infoln("SMAFFINITY bm", bm, "used", used)
	bmout := append([]bool(nil), bm...)

	root := mkNode(nil, bmout, used, 0, len(bmout))
	if root.available() < needed {
		klog.Error("SMAFFINITY wrong: no resource for", needed, bm)
		return nil
	}
	if !rank(root, needed) { // should never get here
		return available.UnsortedList()[:needed]
	}
	klog.V(3).Infoln("SMAFFINITY bmout", bmout)
	ret := make([]string, 0)

	for k, v := range m.gpuBitmap {
		if bm[v] && !bmout[v] {
			ret = append(ret, k)
			klog.V(2).Infoln("SMAFFINITY allocate ", k, "idx", v)
		}
	}
	return ret
}
