// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// Copyright 2018 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package local

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"openpitrix.io/metad/pkg/logger"
	"openpitrix.io/metad/pkg/store"
)

func init() {
	logger.SetLevelByString("debug")
	rand.Seed(int64(time.Now().Nanosecond()))
}

func TestClientSyncStop(t *testing.T) {

	stopChan := make(chan bool)

	storeClient, err := NewLocalClient()
	assert.NoError(t, err)

	go func() {
		time.Sleep(3000 * time.Millisecond)
		stopChan <- true
	}()

	metastore := store.New()
	// expect internalSync not block after stopChan has signal
	storeClient.internalSync("data", storeClient.data, metastore, stopChan)
}
