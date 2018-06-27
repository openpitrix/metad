// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// Copyright 2018 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package store

import (
	"fmt"
	"reflect"
	"testing"

	. "openpitrix.io/metad/pkg/assert"
)

func TestTravellerStack(t *testing.T) {
	stack := &travellerStack{}

	Assert(t, stack.Pop() == nil)

	one := &stackElement{node: nil, mode: AccessModeNil}
	two := &stackElement{node: nil, mode: AccessModeForbidden}
	three := &stackElement{node: nil, mode: AccessModeRead}
	stack.Push(one)
	stack.Push(two)
	stack.Push(three)

	Assert(t, reflect.DeepEqual(three, stack.Pop()))
	Assert(t, reflect.DeepEqual(two, stack.Pop()))
	Assert(t, reflect.DeepEqual(one, stack.Pop()))

	Assert(t, stack.Pop() == nil)
}

func TestTravellerEnter(t *testing.T) {
	s := New()
	data := map[string]interface{}{
		"clusters": map[string]interface{}{
			"cl-1": map[string]interface{}{
				"env": map[string]interface{}{
					"name":   "app1",
					"secret": "123456",
				},
				"public_key": "public_key_val",
			},
			"cl-2": map[string]interface{}{
				"env": map[string]interface{}{
					"name":   "app2",
					"secret": "1234567",
				},
				"public_key": "public_key_val2",
			},
		},
	}
	s.Put("/", data)

	accessRules := []AccessRule{
		{
			Path: "/",
			Mode: AccessModeRead,
		},
	}

	traveller := s.Traveller(NewAccessTree(accessRules))
	defer traveller.Close()
	Assert(t, traveller.Enter("/clusters"))
	Assert(t, traveller.Enter("/cl-1/env"))
	Assert(t, traveller.Enter("name"))
	Assert(t, reflect.DeepEqual("app1", traveller.GetValue()))

	traveller.BackToRoot()
	Assert(t, traveller.Enter("/clusters/cl-1/env/secret"))
	traveller.BackStep(2)
	Assert(t, traveller.Enter("public_key"))
	Assert(t, reflect.DeepEqual("public_key_val", traveller.GetValue()))

	traveller.BackToRoot()
	Assert(t, traveller.Enter("/"))

}

func TestTraveller(t *testing.T) {
	s := New()
	data := map[string]interface{}{
		"clusters": map[string]interface{}{
			"cl-1": map[string]interface{}{
				"env": map[string]interface{}{
					"name":   "app1",
					"secret": "123456",
				},
				"public_key": "public_key_val",
			},
			"cl-2": map[string]interface{}{
				"env": map[string]interface{}{
					"name":   "app2",
					"secret": "1234567",
				},
				"public_key": "public_key_val2",
			},
		},
	}
	s.Put("/", data)

	accessRules := []AccessRule{
		{
			Path: "/",
			Mode: AccessModeForbidden,
		},
		{
			Path: "/clusters",
			Mode: AccessModeRead,
		},
		{
			Path: "/clusters/*/env",
			Mode: AccessModeForbidden,
		},
		{
			Path: "/clusters/cl-1",
			Mode: AccessModeRead,
		},
	}
	traveller := s.Traveller(NewAccessTree(accessRules))
	defer traveller.Close()

	nodeTraveller := traveller.(*nodeTraveller)
	fmt.Println(nodeTraveller.access.Json())

	Assert(t, traveller.Enter("/clusters/cl-1/env"))
	traveller.BackToRoot()

	Assert(t, false == traveller.Enter("/clusters/cl-2/env"))
	Assert(t, traveller.Enter("/clusters/cl-2/public_key"))

	traveller.BackToRoot()

	traveller.Enter("/clusters")
	//traveller.Enter("cl-2")
	v := traveller.GetValue()
	//j,_ := json.MarshalIndent(v, "", "  ")
	//fmt.Printf("%s", string(j))
	mVal, ok := v.(map[string]interface{})
	Assert(t, ok)
	cl1 := mVal["cl-1"].(map[string]interface{})
	cl2 := mVal["cl-2"].(map[string]interface{})

	envM := cl1["env"].(map[string]interface{})
	Assert(t, 2 == len(envM))
	Assert(t, cl2["env"] == nil)
}
