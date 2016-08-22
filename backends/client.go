package backends

import (
	"errors"
	"github.com/yunify/metadata-proxy/backends/etcd"
	"github.com/yunify/metadata-proxy/backends/etcdv3"
	"github.com/yunify/metadata-proxy/log"
	"github.com/yunify/metadata-proxy/store"
	"strings"
)

// The StoreClient interface is implemented by objects that can retrieve
// key/value pairs from a backend store.
type StoreClient interface {
	GetValues(key string) (map[string]interface{}, error)
	GetValue(key string) (string, error)
	Sync(store store.Store, stopChan chan bool)
	SetValues(key string, values map[string]interface{}, replace bool) error
	SetValue(key string, value string) error
	// Delete
	// if the 'key' represent a dir, 'dir' should be true.
	Delete(key string, dir bool) error
	SyncSelfMapping(mapping store.Store, stopChan chan bool)
	RegisterSelfMapping(clientIP string, mapping map[string]string, replace bool) error
	UnregisterSelfMapping(clientIP string) error
}

// New is used to create a storage client based on our configuration.
func New(config Config) (StoreClient, error) {
	if config.Backend == "" {
		config.Backend = "etcd"
	}
	backendNodes := config.BackendNodes
	log.Info("Backend nodes set to " + strings.Join(backendNodes, ", "))
	if len(backendNodes) == 0 {
		backendNodes = GetDefaultBackends(config.Backend)
	}
	switch config.Backend {
	case "etcd":
		// Create the etcd client upfront and use it for the life of the process.
		// The etcdClient is an http.Client and designed to be reused.
		return etcd.NewEtcdClient(config.Prefix, backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
	case "etcdv3":
		// Create the etcdv3 client upfront and use it for the life of the process.
		return etcdv3.NewEtcdClient(config.Prefix, backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys, config.BasicAuth, config.Username, config.Password)
	}

	return nil, errors.New("Invalid backend")
}

func GetDefaultBackends(backend string) []string {
	switch backend {
	case "etcd":
		return []string{"http://127.0.0.1:2379"}
	case "etcdv3":
		return []string{"http://127.0.0.1:2379"}
	default:
		return nil
	}
}
