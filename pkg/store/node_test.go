// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// Copyright 2018 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package store

import (
	"testing"

	. "openpitrix.io/metad/pkg/assert"
)

func TestRelativePath(t *testing.T) {

	s := newStore()
	root := s.Root

	s.Put("/1/2/3", "v")
	n1 := s.internalGet("/1")
	n2 := s.internalGet("/1/2")
	n3 := s.internalGet("/1/2/3")

	Assert(t, "/1/2/3" == n3.RelativePath(root))
	Assert(t, "/2/3" == n3.RelativePath(n1))
	Assert(t, "/3" == n3.RelativePath(n2))
	Assert(t, "/" == n3.RelativePath(n3))
}
