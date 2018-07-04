// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// Copyright 2018 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package metad

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"

	. "openpitrix.io/metad/pkg/assert"
)

func TestConfigFile(t *testing.T) {
	config := Config{
		Backend:      "etcd",
		LogLevel:     "debug",
		PIDFile:      "/var/run/metad.pid",
		EnableXff:    true,
		Prefix:       "/users/uid1",
		Group:        "default",
		Listen:       ":8080",
		ListenManage: "127.0.0.1:9611",
		BasicAuth:    true,
		ClientCaKeys: "/opt/metad/client_ca_keys",
		ClientCert:   "/opt/metad/client_cert",
		ClientKey:    "/opt/metad/client_key",
		BackendNodes: []string{"192.168.11.1:2379", "192.168.11.2:2379"},
		Username:     "username",
		Password:     "password",
	}

	data, err := yaml.Marshal(config)
	Assert(t, err == nil, err)
	configFile, fileErr := ioutil.TempFile("/tmp", "metad")

	fmt.Printf("configFile: %v \n", configFile.Name())

	Assert(t, fileErr == nil, fileErr)
	c, ioErr := configFile.Write(data)
	Assert(t, ioErr == nil, ioErr)
	Assert(t, len(data) == c)

	config2 := Config{}
	loadErr := loadConfigFile(configFile.Name(), &config2)
	Assert(t, loadErr == nil, loadErr)

	Assert(t, reflect.DeepEqual(config, config2))
}
