// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// Copyright 2018 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package metadata

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	. "openpitrix.io/metad/pkg/assert"
	"openpitrix.io/metad/pkg/backends"
	"openpitrix.io/metad/pkg/flatmap"
	"openpitrix.io/metad/pkg/logger"
	"openpitrix.io/metad/pkg/store"
)

func init() {
	logger.SetLevelByString("error")
	rand.Seed(int64(time.Now().Nanosecond()))
}

var (
	backend   = "local"
	maxNode   = 5
	sleepTime = 200 * time.Millisecond
)

func TestMetarepoData(t *testing.T) {

	metarepo := NewTestMetarepo()
	metarepo.DeleteData("/")

	clientIP := "192.168.0.1"
	accessRule := map[string][]store.AccessRule{
		clientIP: {
			{Path: "/", Mode: store.AccessModeRead},
		},
	}
	metarepo.PutAccessRule(accessRule)

	metarepo.StartSync()

	testData := FillTestData(metarepo)
	time.Sleep(sleepTime)
	ValidTestData(t, testData, metarepo.data)

	_, val := metarepo.Root(clientIP, "/nodes/0")
	Assert(t, nil != val)

	mapVal, mok := val.(map[string]interface{})
	Assert(t, mok)

	_, mok = mapVal["name"]
	Assert(t, mok)

	metarepo.DeleteData("/nodes/0")

	time.Sleep(sleepTime)
	val = metarepo.GetData("/nodes/0")
	Assert(t, nil == val)

	subs := []string{"1", "3", "noexistkey"}
	//test batch delete
	err := metarepo.DeleteData("nodes", subs...)
	time.Sleep(sleepTime)
	Assert(t, nil == err)

	for _, sub := range subs {
		val = metarepo.GetData("/nodes/" + sub)
		Assert(t, nil == val)
	}

	val = metarepo.GetData("/nodes/2")
	Assert(t, nil != val)

	metarepo.DeleteData("/")
	metarepo.StopSync()
}

func TestMetarepoMapping(t *testing.T) {

	metarepo := NewTestMetarepo()
	metarepo.DeleteData("/")
	metarepo.DeleteMapping("/")

	metarepo.StartSync()

	key := "node"
	mappings := make(map[string]interface{})
	for i := 0; i < maxNode; i++ {
		ip := fmt.Sprintf("192.168.1.%v", i)
		mapping := map[string]interface{}{
			key:     fmt.Sprintf("/nodes/%v", i),
			"nodes": "/",
		}
		mappings[ip] = mapping
	}
	// batch update
	err := metarepo.PutMapping("/", mappings, true)
	Assert(t, nil == err)
	time.Sleep(sleepTime)

	metarepo.DeleteMapping("/192.168.1.0")

	time.Sleep(sleepTime)
	val := metarepo.GetMapping("/192.168.1.0")
	Assert(t, nil == val)

	subs := []string{"192.168.1.1", "192.168.1.3", "noexistkey"}
	//test batch delete
	err = metarepo.DeleteMapping("/", subs...)
	time.Sleep(sleepTime)
	Assert(t, nil == err)

	for _, sub := range subs {
		val = metarepo.GetMapping("/" + sub)
		Assert(t, nil == val)
	}

	val = metarepo.GetMapping("/192.168.1.2")
	Assert(t, nil != val)

	p := 4
	ip := fmt.Sprintf("192.168.1.%v", p)

	expectMapping0 := map[string]interface{}{
		"node":  fmt.Sprintf("/nodes/%v", p),
		"nodes": "/",
	}

	// test update replace(false)
	err = metarepo.PutMapping(ip, map[string]interface{}{"node2": "/nodes/2"}, false)
	Assert(t, nil == err)

	expectMapping1 := map[string]interface{}{
		"node":  fmt.Sprintf("/nodes/%v", p),
		"nodes": "/",
		"node2": "/nodes/2",
	}
	time.Sleep(sleepTime)
	mapping := metarepo.GetMapping(fmt.Sprintf("/%s", ip))
	Assert(t, reflect.DeepEqual(expectMapping1, mapping))

	// test update key
	err = metarepo.PutMapping(ip+"/node3", "/nodes/3", false)
	Assert(t, nil == err)

	expectMapping2 := map[string]interface{}{
		"node":  fmt.Sprintf("/nodes/%v", p),
		"nodes": "/",
		"node2": "/nodes/2",
		"node3": "/nodes/3",
	}
	time.Sleep(sleepTime)
	mapping = metarepo.GetMapping(fmt.Sprintf("/%s", ip))
	Assert(t, reflect.DeepEqual(expectMapping2, mapping))

	// test delete mapping
	metarepo.DeleteMapping(ip + "/node3")
	time.Sleep(sleepTime)
	mapping = metarepo.GetMapping(fmt.Sprintf("/%s", ip))
	Assert(t, reflect.DeepEqual(expectMapping1, mapping))

	// test update replace(true)
	err = metarepo.PutMapping(ip, expectMapping0, true)
	Assert(t, nil == err)
	time.Sleep(sleepTime)
	mapping = metarepo.GetMapping(fmt.Sprintf("/%s", ip))
	Assert(t, reflect.DeepEqual(expectMapping0, mapping))

	metarepo.DeleteData("/")
	metarepo.DeleteMapping("/")
	metarepo.StopSync()
}

func TestMetarepoSelf(t *testing.T) {
	metarepo := NewTestMetarepo()

	metarepo.DeleteMapping("/")
	metarepo.DeleteData("/")

	metarepo.StartSync()

	testData := FillTestData(metarepo)
	time.Sleep(sleepTime)
	ValidTestData(t, testData, metarepo.data)

	key := "node"
	mappings := make(map[string]interface{})
	for i := 0; i < maxNode; i++ {
		ip := fmt.Sprintf("192.168.1.%v", i)
		mapping := map[string]interface{}{
			key:     fmt.Sprintf("/nodes/%v", i),
			"nodes": "/",
		}
		mappings[ip] = mapping
	}
	// batch update
	err := metarepo.PutMapping("/", mappings, true)
	Assert(t, nil == err)
	time.Sleep(sleepTime)

	//test mapping get
	mappings2 := metarepo.GetMapping("/")
	Assert(t, reflect.DeepEqual(mappings, mappings2))

	// test GetSelf
	time.Sleep(sleepTime)
	p := rand.Intn(maxNode)
	ip := fmt.Sprintf("192.168.1.%v", p)

	val := metarepo.Self(ip, "/")
	mapVal, mok := val.(map[string]interface{})

	Assert(t, mok)
	Assert(t, nil != mapVal[key])

	val = metarepo.Self(ip, "/node/name")
	Assert(t, fmt.Sprintf("node%v", p) == fmt.Sprint(val))

	//test date delete
	metarepo.DeleteData(fmt.Sprintf("/nodes/%v/name", p))

	time.Sleep(sleepTime)
	val = metarepo.Self(ip, "/node/name")
	Assert(t, nil == val)

	metarepo.PutData(fmt.Sprintf("/nodes/%v/name", p), fmt.Sprintf("node%v", p), true)

	//test mapping dir

	err = metarepo.PutMapping(ip, map[string]interface{}{
		"dir": map[string]interface{}{
			"n1": "/nodes/1",
			"n2": "/nodes/2",
		},
	}, false)
	Assert(t, nil == err)

	time.Sleep(sleepTime)
	val = metarepo.Self(ip, "/dir/n1/name")
	if val != "node1" {
		logger.Error("except node1, but get %s, ip: %s, data: %s, mapping:%s", val, ip, metarepo.data.Json(), metarepo.mapping.Json())
		t.Fatal("except node1, but get", val)
	}
	Assert(t, reflect.DeepEqual("node1", val))

	metarepo.DeleteData("/")
	metarepo.DeleteMapping("/")
	metarepo.StopSync()
}

func TestMetarepoRoot(t *testing.T) {

	metarepo := NewTestMetarepo()

	metarepo.DeleteMapping("/")
	metarepo.DeleteData("/")

	FillTestData(metarepo)

	metarepo.StartSync()
	time.Sleep(sleepTime)

	ip := "192.168.1.0"

	accessRule := map[string][]store.AccessRule{
		ip: {
			{Path: "/", Mode: store.AccessModeRead},
		},
	}
	metarepo.PutAccessRule(accessRule)

	mapping := make(map[string]interface{})
	mapping["node"] = "/nodes/0"
	err := metarepo.PutMapping(ip, mapping, true)
	Assert(t, nil == err)

	time.Sleep(sleepTime)
	_, val := metarepo.Root(ip, "/")
	mapVal, mok := val.(map[string]interface{})
	Assert(t, mok)
	//println(fmt.Sprintf("%v", mapVal))
	Assert(t, nil != mapVal["nodes"])
	selfVal := mapVal["self"]
	Assert(t, nil != selfVal)
	Assert(t, len(mapVal) > 1)

	metarepo.DeleteData("/")
	metarepo.DeleteMapping("/")
	metarepo.StopSync()
}

func TestWatch(t *testing.T) {
	metarepo := NewTestMetarepo()
	metarepo.DeleteMapping("/")
	metarepo.DeleteData("/")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ip := "192.168.1.1"

	ch := make(chan interface{})
	defer close(ch)

	go func() {
		ch <- metarepo.Watch(ctx, ip, "/")
	}()

	FillTestData(metarepo)

	time.Sleep(sleepTime)

	metarepo.StartSync()

	time.Sleep(sleepTime)

	//println(metarepo.data.Json())
	var result interface{}
	select {
	case result = <-ch:
	case <-time.Tick(sleepTime):
		logger.Error("TestWatch wait timeout, key: /, ip: %s, data: %s, mapping: %s", ip, metarepo.data.Json(), metarepo.mapping.Json())
		t.Fatal("TestWatch wait timeout")
	}

	m, mok := result.(map[string]interface{})
	Assert(t, mok)
	//println(fmt.Sprintf("%v", m))
	Assert(t, 1 == len(m))
	Assert(t, maxNode*2 == len(flatmap.Flatten(m)))

	//test watch leaf node

	go func() {
		ch <- metarepo.Watch(ctx, ip, "/nodes/1/name")
	}()
	time.Sleep(sleepTime)

	metarepo.PutData("/nodes/1/name", "n1", false)

	select {
	case result = <-ch:
	case <-time.Tick(sleepTime):
		logger.Error("TestWatch wait timeout for key /nodes/1/name , ip: %s, data: %s, mapping: %s", ip, metarepo.data.Json(), metarepo.mapping.Json())
		t.Fatal("TestWatch wait timeout")
	}

	Assert(t, reflect.DeepEqual("UPDATE|n1", result))
	metarepo.StopSync()
}

func TestWatchSelf(t *testing.T) {
	metarepo := NewTestMetarepo()
	metarepo.DeleteMapping("/")
	metarepo.DeleteData("/")

	FillTestData(metarepo)
	metarepo.StartSync()

	ip := "192.168.1.1"

	err := metarepo.PutMapping(ip, map[string]interface{}{
		"node": "/nodes/1",
	}, true)
	Assert(t, nil == err)

	time.Sleep(sleepTime)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan interface{})
	defer close(ch)

	for i := 0; i <= 10; i++ {
		go func() {
			ch <- metarepo.WatchSelf(ctx, "192.168.1.1", "/")
		}()
		time.Sleep(sleepTime)
		//test data change

		name := fmt.Sprintf("n1_%v", i)
		ip := fmt.Sprintf("192.168.1.%v", i)
		err = metarepo.PutData("/nodes/1", map[string]interface{}{
			"name": name,
			"ip":   ip,
		}, false)
		Assert(t, nil == err)

		//println(metarepo.data.Json())

		result := <-ch

		m, mok := result.(map[string]interface{})
		Assert(t, mok)
		//println(fmt.Sprintf("%v", m))
		fmap := flatmap.Flatten(m)
		Assert(t, fmt.Sprintf("UPDATE|%s", name) == fmap["/node/name"])
		Assert(t, fmt.Sprintf("UPDATE|%s", ip) == fmap["/node/ip"])
	}

	// test watch self subdir
	go func() {
		ch <- metarepo.WatchSelf(ctx, "192.168.1.1", "/node")
	}()

	time.Sleep(sleepTime)

	err = metarepo.DeleteData("/nodes/1/name")
	Assert(t, nil == err)

	result := <-ch

	m, mok := result.(map[string]interface{})
	Assert(t, mok)
	//println(fmt.Sprintf("%v", m))
	Assert(t, "DELETE|n1_10" == fmt.Sprint(m["name"]))

	//logger.Debug("TimerPool stat,total New:%v, Get:%v", metarepo.timerPool.TotalNew.Get(), metarepo.timerPool.TotalGet.Get())
	metarepo.StopSync()
}

func TestWatchCloseChan(t *testing.T) {
	metarepo := NewTestMetarepo()

	metarepo.StartSync()

	ip := "192.168.1.1"

	err := metarepo.PutMapping(ip, map[string]interface{}{
		"node": "/nodes/1",
	}, true)
	Assert(t, nil == err)

	time.Sleep(sleepTime)

	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())

	ch := make(chan interface{})
	defer close(ch)

	ch2 := make(chan interface{})
	defer close(ch2)

	go func() {
		ch <- metarepo.Watch(ctx1, "192.168.1.1", "/")
	}()
	go func() {
		ch2 <- metarepo.WatchSelf(ctx2, ip, "/")
	}()

	time.Sleep(sleepTime)

	cancel1()
	result := <-ch
	Assert(t, nil != result)
	cancel2()
	result2 := <-ch2
	Assert(t, nil != result2)
	metarepo.StopSync()
}

// TestSelfWatchNodeNotExist
// create a mapping with a node not exist, then update the node
func TestSelfWatchNodeNotExist(t *testing.T) {
	metarepo := NewTestMetarepo()

	metarepo.StartSync()

	//fmt.Printf("data:%p mapping:%p \n", metarepo.data.)

	ip := "192.168.1.1"

	err := metarepo.PutMapping(ip, map[string]interface{}{
		"host": "/hosts/i-local",
		"cmd":  "/cmd/i-local",
	}, true)
	Assert(t, nil == err)

	//err = metarepo.PutData("/hosts/i-local", ip , true)
	//Assert(t, nil == err)

	time.Sleep(sleepTime)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan interface{})
	defer close(ch)
	go func() {
		ch <- metarepo.WatchSelf(ctx, ip, "/")
	}()

	time.Sleep(sleepTime)

	err = metarepo.PutData("/cmd/i-local", "start", true)
	Assert(t, nil == err)
	result := <-ch
	println(fmt.Sprintf("%s", result))
	Assert(t, nil != result)
	//closeChan <- true
	metarepo.StopSync()
}

func TestAccessRule(t *testing.T) {
	metarepo := NewTestMetarepo()
	metarepo.StartSync()

	data := map[string]interface{}{
		"clusters": map[string]interface{}{
			"cl-1": map[string]interface{}{
				"name": "cl-1",
				"env": map[string]interface{}{
					"username": "user1",
					"secret":   "123456",
				},
				"public_key": "public_key_val",
			},
			"cl-2": map[string]interface{}{
				"name": "cl-2",
				"env": map[string]interface{}{
					"username": "user2",
					"secret":   "1234567",
				},
				"public_key": "public_key_val2",
			},
		},
	}

	err := metarepo.PutData("/", data, true)
	Assert(t, nil == err)

	ip := "192.168.1.1"
	rules := map[string][]store.AccessRule{
		ip: {
			{Path: "/", Mode: store.AccessModeRead},
		},
	}

	err = metarepo.PutAccessRule(rules)
	Assert(t, nil == err)

	time.Sleep(sleepTime)

	rulesGet := metarepo.GetAccessRule([]string{ip})
	Assert(t, reflect.DeepEqual(rules, rulesGet))

	_, dataGet := metarepo.Root(ip, "/")
	Assert(t, reflect.DeepEqual(data, dataGet))

	metarepo.StopSync()
}

func NewTestMetarepo() *MetadataRepo {
	prefix := fmt.Sprintf("/prefix%v", rand.Intn(10000))
	group := fmt.Sprintf("/group%v", rand.Intn(10000))
	nodes := backends.GetDefaultBackends(backend)

	config := backends.Config{
		Backend:      backend,
		BackendNodes: nodes,
		Prefix:       prefix,
		Group:        group,
	}
	storeClient, err := backends.New(config)
	if err != nil {
		panic(err)
	}

	return New(storeClient)
}

func FillTestData(metarepo *MetadataRepo) map[string]string {
	nodes := make(map[string]interface{})
	for i := 0; i < maxNode; i++ {
		node := make(map[string]interface{})
		node["name"] = fmt.Sprintf("node%v", i)
		node["ip"] = fmt.Sprintf("192.168.1.%v", i)
		nodes[fmt.Sprintf("%v", i)] = node
	}
	testData := map[string]interface{}{
		"nodes": nodes,
	}
	err := metarepo.PutData("/", testData, true)
	if err != nil {
		logger.Error("SetValues error", err.Error())
		panic(err)
	}
	return flatmap.Flatten(testData)
}

func ValidTestData(t *testing.T, testData map[string]string, metastore store.Store) {
	for k, v := range testData {
		_, storeVal := metastore.Get(k)
		Assert(t, reflect.DeepEqual(v, storeVal))
	}
}
