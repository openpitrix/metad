// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// Copyright 2018 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package backends

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	. "openpitrix.io/metad/pkg/assert"
	"openpitrix.io/metad/pkg/flatmap"
	"openpitrix.io/metad/pkg/logger"
	"openpitrix.io/metad/pkg/store"
)

var (
	backendNodes = []string{
		"etcdv3",
		"local",
	}
)

func init() {
	logger.SetLevelByString("debug")
	rand.Seed(int64(time.Now().Nanosecond()))
}

func TestClientGetPut(t *testing.T) {
	for _, backend := range backendNodes {
		println("Test backend: ", backend)

		prefix := fmt.Sprintf("/prefix%v", rand.Intn(1000))

		nodes := GetDefaultBackends(backend)

		config := Config{
			Backend:      backend,
			BackendNodes: nodes,
			Prefix:       prefix,
		}
		storeClient, err := New(config)
		if err != nil {
			t.Fatal(err)
		}

		err = storeClient.Delete("/", true)
		if err != nil {
			t.Fatal(err)
		}

		err = storeClient.Put("testkey", "testvalue", false)
		if err != nil {
			t.Fatal(err)
		}

		val, err := storeClient.Get("testkey", false)
		if err != nil {
			t.Fatal(err)
		}
		if "testvalue" != val {
			t.Fatalf("equal failed: %v", val)
		}

		// test no exist key
		val, err = storeClient.Get("noexistkey", false)
		if err != nil {
			t.Fatal(err)
		}
		if "" != val {
			t.Fatalf("equal failed: %v", val)
		}

		storeClient.Delete("/", true)
	}
}

func TestClientGetsPuts(t *testing.T) {
	for _, backend := range backendNodes {
		println("Test backend: ", backend)

		prefix := fmt.Sprintf("/prefix%v", rand.Intn(1000))

		nodes := GetDefaultBackends(backend)

		config := Config{
			Backend:      backend,
			BackendNodes: nodes,
			Prefix:       prefix,
		}
		storeClient, err := New(config)
		Assert(t, nil == err)

		Assert(t, nil == storeClient.Delete("/", true))

		values := map[string]interface{}{
			"subkey1": map[string]interface{}{
				"subkey1sub1": "subsubvalue1",
				"subkey1sub2": "subsubvalue2",
			},
		}

		err = storeClient.Put("testkey", values, true)
		Assert(t, nil == err)

		val, getErr := storeClient.Get("testkey", true)
		Assert(t, nil == getErr)
		Assert(t, reflect.DeepEqual(values, val))

		//test update

		values2 := map[string]interface{}{
			"subkey1": map[string]interface{}{
				"subkey1sub3": "subsubvalue3",
			},
		}

		err = storeClient.Put("testkey", values2, false)
		Assert(t, nil == err)

		values3 := map[string]interface{}{
			"subkey1": map[string]interface{}{
				"subkey1sub1": "subsubvalue1",
				"subkey1sub2": "subsubvalue2",
				"subkey1sub3": "subsubvalue3",
			},
		}

		val, getErr = storeClient.Get("testkey", true)
		Assert(t, nil == getErr)
		Assert(t, reflect.DeepEqual(values3, val))

		//test replace

		err = storeClient.Put("testkey", values2, true)
		Assert(t, nil == err)

		val, getErr = storeClient.Get("testkey", true)
		Assert(t, nil == getErr)
		Assert(t, reflect.DeepEqual(values2, val))

		Assert(t, nil == storeClient.Delete("/", true))
	}
}

func TestClientPutJSON(t *testing.T) {
	for _, backend := range backendNodes {
		println("Test backend: ", backend)

		prefix := fmt.Sprintf("/prefix%v", rand.Intn(1000))

		nodes := GetDefaultBackends(backend)

		config := Config{
			Backend:      backend,
			BackendNodes: nodes,
			Prefix:       prefix,
		}
		storeClient, err := New(config)
		Assert(t, nil == err)

		Assert(t, nil == storeClient.Delete("/", true))

		jsonVal := []byte(`
			{"subkey1":
				{
					"subkey1sub1":"subsubvalue1",
					"subkey1sub2": "subsubvalue2"
				}
			}
		`)
		var values interface{}
		err = json.Unmarshal(jsonVal, &values)
		Assert(t, nil == err)

		err = storeClient.Put("testkey", values, true)
		Assert(t, nil == err)

		val, getErr := storeClient.Get("testkey", true)
		Assert(t, nil == getErr)
		Assert(t, reflect.DeepEqual(values, val))

		//test update

		jsonVal2 := []byte(`
			{"subkey1":
				{
					"subkey1sub3":"subsubvalue3"
				},
			 "int":9663676416,
			 "bool":true,
			 "float":1.1111111
			}
		`)

		var values2 interface{}
		err = json.Unmarshal(jsonVal2, &values2)
		Assert(t, nil == err)

		err = storeClient.Put("testkey", values2, false)
		Assert(t, nil == err)

		values3 := map[string]interface{}{
			"subkey1": map[string]interface{}{
				"subkey1sub1": "subsubvalue1",
				"subkey1sub2": "subsubvalue2",
				"subkey1sub3": "subsubvalue3",
			},
			"int":   "9663676416",
			"bool":  "true",
			"float": "1.1111111",
		}

		val, getErr = storeClient.Get("testkey", true)
		Assert(t, nil == getErr)
		Assert(t, reflect.DeepEqual(values3, val))

		//test replace

		err = storeClient.Put("testkey", values2, true)
		Assert(t, nil == err)

		values4 := map[string]interface{}{
			"subkey1": map[string]interface{}{
				"subkey1sub3": "subsubvalue3",
			},
			"int":   "9663676416",
			"bool":  "true",
			"float": "1.1111111",
		}

		val, getErr = storeClient.Get("testkey", true)
		Assert(t, nil == getErr)
		Assert(t, reflect.DeepEqual(values4, val))

		Assert(t, nil == storeClient.Delete("/", true))
	}
}

func TestClientNoPrefix(t *testing.T) {
	for _, backend := range backendNodes {
		println("Test backend: ", backend)

		prefix := ""

		stopChan := make(chan bool)
		defer func() {
			stopChan <- true
		}()

		nodes := GetDefaultBackends(backend)

		config := Config{
			Backend:      backend,
			BackendNodes: nodes,
			Prefix:       prefix,
		}
		storeClient, err := New(config)
		Assert(t, nil == err)

		Assert(t, nil == storeClient.Delete("/", true))

		metastore := store.New()
		storeClient.Sync(metastore, stopChan)

		values := map[string]interface{}{
			"subkey1": map[string]interface{}{
				"subkey1sub1": "subsubvalue1",
				"subkey1sub2": "subsubvalue2",
			},
		}

		err = storeClient.Put("testkey", values, true)
		Assert(t, nil == err)

		val, getErr := storeClient.Get("testkey", true)
		Assert(t, nil == getErr)
		Assert(t, reflect.DeepEqual(values, val))

		mappings := map[string]interface{}{
			"192.168.1.1": map[string]interface{}{
				"key": "/subkey1/subkey1sub1",
			},
		}
		err = storeClient.PutMapping("/", mappings, true)

		mappings2, merr := storeClient.GetMapping("/", true)
		Assert(t, nil == merr)
		Assert(t, reflect.DeepEqual(mappings, mappings2))

		time.Sleep(1000 * time.Millisecond)

		// mapping data should not sync to metadata
		_, val = metastore.Get("/_metad")
		Assert(t, nil == val)

		Assert(t, nil == storeClient.Delete("/", true))
		Assert(t, nil == getErr)

		val, getErr = storeClient.Get("testkey", true)
		Assert(t, nil == getErr)
		Assert(t, 0 == len(val.(map[string]interface{})))

		// delete data "/" should not delete mapping
		mappings2, merr = storeClient.GetMapping("/", true)
		Assert(t, nil == merr)
		Assert(t, reflect.DeepEqual(mappings, mappings2))
	}
}

func TestClientSetMaxOps(t *testing.T) {
	//TODO for etcd3 batch update max ops
}

func TestClientSync(t *testing.T) {

	for _, backend := range backendNodes {
		println("Test backend: ", backend)

		prefix := fmt.Sprintf("/prefix%v", rand.Intn(1000))

		stopChan := make(chan bool)
		defer func() {
			stopChan <- true
		}()

		nodes := GetDefaultBackends(backend)

		config := Config{
			Backend:      backend,
			BackendNodes: nodes,
			Prefix:       prefix,
		}
		storeClient, err := New(config)
		Assert(t, nil == err)

		storeClient.Delete("/", true)
		//Assert(t, nil==err)

		metastore := store.New()
		storeClient.Sync(metastore, stopChan)

		testData := FillTestData(storeClient)
		time.Sleep(2000 * time.Millisecond)
		ValidTestData(t, testData, metastore, backend)

		RandomUpdate(testData, storeClient, 10)
		time.Sleep(1000 * time.Millisecond)
		ValidTestData(t, testData, metastore, backend)

		deletedKey := RandomDelete(testData, storeClient)
		time.Sleep(1000 * time.Millisecond)
		ValidTestData(t, testData, metastore, backend)

		_, val := metastore.Get(deletedKey)
		Assert(t, nil == val)

		storeClient.Delete("/", true)
	}
}

func TestMapping(t *testing.T) {
	for _, backend := range backendNodes {
		println("Test backend: ", backend)
		prefix := fmt.Sprintf("/prefix%v", rand.Intn(1000))
		group := fmt.Sprintf("/group%v", rand.Intn(1000))
		nodes := GetDefaultBackends(backend)

		config := Config{
			Backend:      backend,
			BackendNodes: nodes,
			Prefix:       prefix,
			Group:        group,
		}
		storeClient, err := New(config)
		Assert(t, nil == err)
		mappings := make(map[string]interface{})
		for i := 0; i < 10; i++ {
			ip := fmt.Sprintf("192.168.1.%v", i)
			mapping := map[string]string{
				"instance": fmt.Sprintf("/instances/%v", i),
				"config":   fmt.Sprintf("/configs/%v", i),
			}
			mappings[ip] = mapping
		}
		storeClient.PutMapping("/", mappings, true)

		val, err := storeClient.GetMapping("/", true)
		Assert(t, nil == err)
		m, mok := val.(map[string]interface{})
		Assert(t, mok)
		Assert(t, m["192.168.1.0"] != nil)

		ip := fmt.Sprintf("192.168.1.%v", 1)
		nodePath := "/" + ip + "/" + "instance"
		storeClient.PutMapping(nodePath, "/instances/new1", true)
		time.Sleep(1000 * time.Millisecond)
		val, err = storeClient.GetMapping(nodePath, false)
		Assert(t, nil == err)
		Assert(t, reflect.DeepEqual("/instances/new1", val))
		storeClient.Delete("/", true)
		storeClient.DeleteMapping("/", true)
	}
}

func TestMappingSync(t *testing.T) {

	for _, backend := range backendNodes {
		prefix := fmt.Sprintf("/prefix%v", rand.Intn(1000))
		group := fmt.Sprintf("/group%v", rand.Intn(1000))
		println("Test backend: ", backend)
		stopChan := make(chan bool)
		defer func() {
			stopChan <- true
		}()
		nodes := GetDefaultBackends(backend)

		config := Config{
			Backend:      backend,
			BackendNodes: nodes,
			Prefix:       prefix,
			Group:        group,
		}
		storeClient, err := New(config)
		Assert(t, nil == err)

		mappingstore := store.New()

		//for test init sync.

		for i := 0; i < 5; i++ {
			ip := fmt.Sprintf("192.168.1.%v", i)
			mapping := map[string]string{
				"instance": fmt.Sprintf("/instances/%v", i),
				"config":   fmt.Sprintf("/configs/%v", i),
			}
			storeClient.PutMapping(ip, mapping, true)
		}
		time.Sleep(1000 * time.Millisecond)
		storeClient.SyncMapping(mappingstore, stopChan)
		time.Sleep(1000 * time.Millisecond)

		for i := 0; i < 5; i++ {
			ip := fmt.Sprintf("192.168.1.%v", i)
			_, val := mappingstore.Get(ip)
			mapVal, mok := val.(map[string]interface{})
			Assert(t, mok)
			path := mapVal["instance"]
			Assert(t, reflect.DeepEqual(path, fmt.Sprintf("/instances/%v", i)))
		}

		for i := 5; i < 10; i++ {
			ip := fmt.Sprintf("192.168.1.%v", i)
			mapping := map[string]string{
				"instance": fmt.Sprintf("/instances/%v", i),
				"config":   fmt.Sprintf("/configs/%v", i),
			}
			storeClient.PutMapping(ip, mapping, true)
		}
		time.Sleep(1000 * time.Millisecond)

		for i := 5; i < 10; i++ {
			ip := fmt.Sprintf("192.168.1.%v", i)
			_, val := mappingstore.Get(ip)
			mapVal, mok := val.(map[string]interface{})
			Assert(t, mok)
			path := mapVal["instance"]
			Assert(t, reflect.DeepEqual(path, fmt.Sprintf("/instances/%v", i)))
		}
		ip := fmt.Sprintf("192.168.1.%v", 1)
		nodePath := ip + "/" + "instance"
		storeClient.PutMapping(nodePath, "/instances/new1", true)
		time.Sleep(1000 * time.Millisecond)
		_, val := mappingstore.Get(nodePath)
		Assert(t, reflect.DeepEqual("/instances/new1", val))
		storeClient.Delete("/", true)
		storeClient.DeleteMapping("/", true)
	}
}

func TestAccessRule(t *testing.T) {
	for _, backend := range backendNodes {
		stopChan := make(chan bool)
		defer func() {
			stopChan <- true
		}()
		storeClient := NewTestClient(backend)

		accessStore := store.NewAccessStore()
		storeClient.SyncAccessRule(accessStore, stopChan)

		rules := map[string][]store.AccessRule{
			"192.168.1.1": {
				{Path: "/clusters", Mode: store.AccessModeForbidden},
				{Path: "/clusters/cl-1", Mode: store.AccessModeRead},
			},
			"192.168.1.2": {
				{Path: "/clusters", Mode: store.AccessModeForbidden},
				{Path: "/clusters/cl-2", Mode: store.AccessModeRead},
			},
		}
		var rulesGet map[string][]store.AccessRule
		err := storeClient.PutAccessRule(rules)
		Assert(t, nil == err)

		rulesGet, err = storeClient.GetAccessRule()
		Assert(t, nil == err)
		Assert(t, reflect.DeepEqual(rules, rulesGet))

		time.Sleep(1000 * time.Millisecond)
		Assert(t, accessStore.Get("192.168.1.1") != nil)

		err = storeClient.DeleteAccessRule([]string{"192.168.1.2"})
		Assert(t, nil == err)

		rulesGet, err = storeClient.GetAccessRule()
		Assert(t, nil == err)
		_, ok := rulesGet["192.168.1.2"]
		Assert(t, ok == false)

		rules2 := map[string][]store.AccessRule{
			"192.168.1.3": {
				{Path: "/clusters", Mode: store.AccessModeForbidden},
				{Path: "/clusters/cl-3", Mode: store.AccessModeRead},
			},
		}
		err = storeClient.PutAccessRule(rules2)
		Assert(t, nil == err)

		time.Sleep(1000 * time.Millisecond)

		Assert(t, nil == accessStore.Get("192.168.1.2"))
		Assert(t, accessStore.Get("192.168.1.3") != nil)

		err = storeClient.DeleteAccessRule([]string{"192.168.1.1", "192.168.1.2", "192.168.1.3"})
		Assert(t, nil == err)
	}
}

func NewTestClient(backend string) StoreClient {
	prefix := fmt.Sprintf("/prefix%v", rand.Intn(1000))
	group := fmt.Sprintf("/group%v", rand.Intn(1000))
	println("Test backend: ", backend)
	nodes := GetDefaultBackends(backend)

	config := Config{
		Backend:      backend,
		BackendNodes: nodes,
		Prefix:       prefix,
		Group:        group,
	}
	storeClient, err := New(config)
	if err != nil {
		panic(err)
	}
	return storeClient
}

func FillTestData(storeClient StoreClient) map[string]string {
	testData := make(map[string]interface{})
	for i := 0; i < 5; i++ {
		ci := make(map[string]string)
		for j := 0; j < 5; j++ {
			ci[fmt.Sprintf("%v", j)] = fmt.Sprintf("%v-%v", i, j)
		}
		testData[fmt.Sprintf("%v", i)] = ci
	}
	err := storeClient.Put("/", testData, true)
	if err != nil {
		logger.Error("SetValues error", err.Error())
	}
	return flatmap.Flatten(testData)
}

func RandomUpdate(testData map[string]string, storeClient StoreClient, times int) {
	length := len(testData)
	keys := make([]string, 0, length)
	for k := range testData {
		keys = append(keys, k)
	}
	for i := 0; i < times; i++ {
		idx := rand.Intn(length)
		key := keys[idx]
		val := testData[key]
		newVal := fmt.Sprintf("%s-%v", val, 0)
		storeClient.Put(key, newVal, true)
		testData[key] = newVal
	}
}

func RandomDelete(testData map[string]string, storeClient StoreClient) string {
	length := len(testData)
	keys := make([]string, 0, length)
	for k := range testData {
		keys = append(keys, k)
	}
	idx := rand.Intn(length)
	key := keys[idx]
	storeClient.Delete(key, false)
	delete(testData, key)
	return key
}

func ValidTestData(t *testing.T, testData map[string]string, metastore store.Store, backend string) {
	for k, v := range testData {
		_, storeVal := metastore.Get(k)
		Assertf(t, reflect.DeepEqual(v, storeVal), "valid data fail for backend %s", backend)
	}
}
