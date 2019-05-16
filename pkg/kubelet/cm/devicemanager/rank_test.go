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
	"flag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestRank(t *testing.T) {
	var available sets.String
	var ret []string
	var logLevel, v string

	klog.InitFlags(flag.CommandLine)
	flag.StringVar(&logLevel, "logLevel", "3", "test")
	flag.Lookup("v").Value.Set(logLevel)

	mgr := &ManagerImpl{gpuBitmap: make(map[string]int, 0)}
	for i := 0; i < 8; i++ {
		mgr.gpuBitmap[strconv.Itoa(i)] = i
	}

	klog.V(2).Infoln("================= Begin tests ==================")
	// 0-7, need 1
	available = make(sets.String, 0)
	available.Insert("0", "1", "2", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 1, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0" {
		t.FailNow()
	}

	// 0-7, need 2
	available = make(sets.String, 0)
	available.Insert("0", "1", "2", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 2, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,1" {
		t.FailNow()
	}

	// 0-7, need 3
	available = make(sets.String, 0)
	available.Insert("0", "1", "2", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 3, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,1,2" {
		t.FailNow()
	}

	// 0-7, need 4
	available = make(sets.String, 0)
	available.Insert("0", "1", "2", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 4, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,1,2,3" {
		t.FailNow()
	}

	// 0-7, need 5
	available = make(sets.String, 0)
	available.Insert("0", "1", "2", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 5, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,1,2,3,4" {
		t.FailNow()
	}

	// 0-7, need 6
	available = make(sets.String, 0)
	available.Insert("0", "1", "2", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 6, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,1,2,3,4,5" {
		t.FailNow()
	}

	// 0-7, need 7
	available = make(sets.String, 0)
	available.Insert("0", "1", "2", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 7, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,1,2,3,4,5,6" {
		t.FailNow()
	}

	// 0-7, need 8
	available = make(sets.String, 0)
	available.Insert("0", "1", "2", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 8, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,1,2,3,4,5,6,7" {
		t.FailNow()
	}

	// 03 left, 4-7 right, need 4
	available = make(sets.String, 0)
	available.Insert("0", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 4, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "4,5,6,7" {
		t.FailNow()
	}

	// 03 left, 4-7 right, need 1
	available = make(sets.String, 0)
	available.Insert("0", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 1, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0" {
		t.FailNow()
	}

	// 0,2 left, 4-7 right, need 1
	available = make(sets.String, 0)
	available.Insert("0", "2", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 1, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0" {
		t.FailNow()
	}

	// 0,3 left, 4-7 right, need 2
	available = make(sets.String, 0)
	available.Insert("0", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 2, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "4,5" {
		t.FailNow()
	}

	// 0,2 left, 4-7 right, need 2
	available = make(sets.String, 0)
	available.Insert("0", "2", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 2, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "4,5" {
		t.FailNow()
	}

	// 023 left, 4-7 right, need 1
	available = make(sets.String, 0)
	available.Insert("0", "2", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 1, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0" {
		t.FailNow()
	}

	// 023 left, 4-7 right, need 2
	available = make(sets.String, 0)
	available.Insert("0", "2", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 2, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "2,3" {
		t.FailNow()
	}

	// 023 left, 4-7 right, need 3
	available = make(sets.String, 0)
	available.Insert("0", "2", "3", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 3, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,2,3" {
		t.FailNow()
	}

	// 0 left, 4-7 right, need 1
	available = make(sets.String, 0)
	available = make(sets.String, 0)
	available.Insert("0", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 1, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0" {
		t.FailNow()
	}

	// 0 left, 4-7 right, need 3
	available = make(sets.String, 0)
	available.Insert("0", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 3, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "4,5,6" {
		t.FailNow()
	}

	// 0 left, 4-7 right, need 4
	available = make(sets.String, 0)
	available.Insert("0", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 4, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "4,5,6,7" {
		t.FailNow()
	}

	// 0 left, 4-7 right, need 5
	available = make(sets.String, 0)
	available.Insert("0", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 5, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,4,5,6,7" {
		t.FailNow()
	}

	// 02 left, 67 right, need 4
	available = make(sets.String, 0)
	available.Insert("0", "2", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 4, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,2,6,7" {
		t.FailNow()
	}

	// 02 left, 67 right, need 2
	available = make(sets.String, 0)
	available.Insert("0", "2", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 2, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "6,7" {
		t.FailNow()
	}

	// 01 left, 67 right, need 2
	available = make(sets.String, 0)
	available.Insert("0", "1", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 2, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,1" {
		t.FailNow()
	}

	// 02 left, 57 right, need 4
	available = make(sets.String, 0)
	available.Insert("0", "2", "5", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 4, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,2,5,7" {
		t.FailNow()
	}

	// 023 left, 7 right, need 3
	available = make(sets.String, 0)
	available.Insert("0", "2", "3", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 3, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "0,2,3" {
		t.FailNow()
	}

	// 0 left, 567 right, need 3
	available = make(sets.String, 0)
	available.Insert("0", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 3, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "5,6,7" {
		t.FailNow()
	}

	// ERR 1 holes on left, need 6
	available = make(sets.String, 0)
	available.Insert("0", "4", "5", "6", "7")
	ret = mgr.calcAllocated("nvidia.com/gpu", 6, sets.String{}, available)
	sort.Sort(sort.StringSlice(ret))
	v = strings.Join(ret, ",")
	klog.V(2).Infoln("========== result", v)
	if v != "" {
		t.FailNow()
	}

	klog.V(2).Infoln("================= All tests done ==================")
}
