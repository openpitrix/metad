// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// Copyright 2018 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package store

import (
	"encoding/json"
	"reflect"
	"testing"

	. "openpitrix.io/metad/pkg/assert"
)

func TestAccessStore(t *testing.T) {
	accessStore := NewAccessStore()
	ip := "192.168.1.1"
	rules := []AccessRule{
		{Path: "/", Mode: AccessModeForbidden},
		{Path: "/clusters", Mode: AccessModeRead},
		{Path: "/clusters/cl-1/env/secret", Mode: AccessModeForbidden},
	}
	accessStore.Put(ip, rules)

	tree := accessStore.Get(ip)
	Assert(t, tree != nil)

	ip2 := "192.168.1.2"
	accessStore.Puts(map[string][]AccessRule{
		ip2: rules,
	})
	rulesGet := accessStore.GetAccessRule([]string{ip})
	rules1 := rulesGet[ip]
	Assert(t, reflect.DeepEqual(rules, rules1))

	rulesGet2 := accessStore.GetAccessRule(nil)
	Assert(t, reflect.DeepEqual(rules, rulesGet2[ip]))
	Assert(t, reflect.DeepEqual(rules, rulesGet2[ip2]))

	accessStore.Delete(ip)
	rulesGet3 := accessStore.GetAccessRule([]string{})
	Assert(t, 1 == len(rulesGet3))
}

func TestAccessTree(t *testing.T) {
	rules := []AccessRule{
		{Path: "/", Mode: AccessModeForbidden},
		{Path: "/clusters", Mode: AccessModeRead},
		{Path: "/clusters/*/env", Mode: AccessModeForbidden},
		{Path: "/clusters/cl-1/env/secret", Mode: AccessModeRead},
	}
	tree := NewAccessTree(rules)
	jsonStr := tree.Json()
	jsonMap := map[string]interface{}{}
	err := json.Unmarshal([]byte(jsonStr), &jsonMap)
	Assert(t, err == nil)
	root := tree.GetRoot()
	Assert(t, AccessModeForbidden == root.Mode)
	Assert(t, AccessModeRead == root.GetChild("clusters", true).Mode)
	Assert(t, AccessModeForbidden == root.GetChild("clusters", true).
		GetChild("cl-2", false).GetChild("env", true).Mode)
	Assert(t, AccessModeRead == root.GetChild("clusters", true).
		GetChild("cl-1", false).GetChild("env", true).GetChild("secret", true).Mode)
}
