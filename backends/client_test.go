package backends

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/yunify/metadata-proxy/log"
	"github.com/yunify/metadata-proxy/store"
	"github.com/yunify/metadata-proxy/util/flatmap"
	"math/rand"
	"testing"
	"time"
)

var (
	backendNodes = []string{
		"etcd",
		"etcdv3",
	}
)

func init() {
	log.SetLevel("debug")
}

func TestClientGetSet(t *testing.T) {
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
		assert.NoError(t, err)

		storeClient.Delete("/", true)

		err = storeClient.SetValue("testkey", "testvalue")
		assert.NoError(t, err)

		val, getErr := storeClient.GetValue("testkey")
		assert.NoError(t, getErr)
		assert.Equal(t, "testvalue", val)

		// test no exist key
		val, getErr = storeClient.GetValue("noexistkey")
		assert.NoError(t, getErr)
		assert.Equal(t, "", val)

		storeClient.Delete("/", true)
	}
}

func TestClientGetsSets(t *testing.T) {
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
		assert.NoError(t, err)

		storeClient.Delete("/", true)

		values := map[string]interface{}{
			"subkey1": map[string]interface{}{
				"subkey1sub1": "subsubvalue1",
				"subkey1sub2": "subsubvalue2",
			},
		}

		err = storeClient.SetValues("testkey", values, true)
		assert.NoError(t, err)

		val, getErr := storeClient.GetValues("testkey")
		assert.NoError(t, getErr)
		assert.Equal(t, values, val)

		//test update

		values2 := map[string]interface{}{
			"subkey1": map[string]interface{}{
				"subkey1sub3": "subsubvalue3",
			},
		}

		err = storeClient.SetValues("testkey", values2, false)
		assert.NoError(t, err)

		values3 := map[string]interface{}{
			"subkey1": map[string]interface{}{
				"subkey1sub1": "subsubvalue1",
				"subkey1sub2": "subsubvalue2",
				"subkey1sub3": "subsubvalue3",
			},
		}

		val, getErr = storeClient.GetValues("testkey")
		assert.NoError(t, getErr)
		assert.Equal(t, values3, val)

		//test replace

		err = storeClient.SetValues("testkey", values2, true)
		assert.NoError(t, err)

		val, getErr = storeClient.GetValues("testkey")
		assert.NoError(t, getErr)
		assert.Equal(t, values2, val)

		assert.NoError(t, storeClient.Delete("/", true))
	}
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
		assert.NoError(t, err)

		storeClient.Delete("/", true)
		//assert.NoError(t, err)

		metastore := store.New()
		storeClient.Sync(metastore, stopChan)

		testData := FillTestData(storeClient)
		time.Sleep(1000 * time.Millisecond)
		ValidTestData(t, testData, metastore)

		RandomUpdate(testData, storeClient, 10)
		time.Sleep(1000 * time.Millisecond)
		ValidTestData(t, testData, metastore)

		deletedKey := RandomDelete(testData, storeClient)
		time.Sleep(1000 * time.Millisecond)
		ValidTestData(t, testData, metastore)

		val, ok := metastore.Get(deletedKey)
		assert.False(t, ok)
		assert.Nil(t, val)

		storeClient.Delete("/", true)
	}
}

func TestSelfMapping(t *testing.T) {

	prefix := fmt.Sprintf("/prefix%v", rand.Intn(1000))

	for _, backend := range backendNodes {
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
		}
		storeClient, err := New(config)
		assert.NoError(t, err)

		metastore := store.New()
		storeClient.SyncSelfMapping(metastore, stopChan)

		for i := 0; i < 10; i++ {
			ip := fmt.Sprintf("192.168.1.%v", i)
			mapping := map[string]string{
				"instance": fmt.Sprintf("/instances/%v", i),
			}
			storeClient.RegisterSelfMapping(ip, mapping, true)
		}
		time.Sleep(1000 * time.Millisecond)
		for i := 0; i < 10; i++ {
			ip := fmt.Sprintf("192.168.1.%v", i)
			val, ok := metastore.Get(ip)
			assert.True(t, ok)
			mapVal, mok := val.(map[string]interface{})
			assert.True(t, mok)
			path := mapVal["instance"]
			assert.Equal(t, path, fmt.Sprintf("/instances/%v", i))
		}
	}
}

func FillTestData(storeClient StoreClient) map[string]string {
	testData := make(map[string]interface{})
	for i := 0; i < 10; i++ {
		ci := make(map[string]string)
		for j := 0; j < 10; j++ {
			ci[fmt.Sprintf("%v", j)] = fmt.Sprintf("%v-%v", i, j)
		}
		testData[fmt.Sprintf("%v", i)] = ci
	}
	err := storeClient.SetValues("/", testData, true)
	if err != nil {
		log.Error("SetValues error", err.Error())
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
		storeClient.SetValue(key, newVal)
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

func ValidTestData(t *testing.T, testData map[string]string, metastore store.Store) {
	for k, v := range testData {
		storeVal, _ := metastore.Get(k)
		assert.Equal(t, v, storeVal)
	}
}
