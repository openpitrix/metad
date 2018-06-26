// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// Copyright 2018 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"openpitrix.io/metad/pkg/logger"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			// metad can run as a service, and enable the auto restart flag.
			// see docs/service.md for more information.
			logger.Fatal("Main Recover: %v, try restart.", r)
		}
	}()

	flag.Parse()

	if printVersion {
		fmt.Printf("Metad Version: %s\n", VERSION)
		fmt.Printf("Git Version: %s\n", GIT_VERSION)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	if pprof {
		fmt.Printf("Start pprof, 127.0.0.1:6060\n")
		go logger.Fatal("%v", http.ListenAndServe("127.0.0.1:6060", nil))
	}

	var config *Config
	var err error
	if config, err = initConfig(); err != nil {
		logger.Fatal(err.Error())
		os.Exit(-1)
	}

	logger.Info("Starting metad %s", VERSION)
	metad, err = New(config)
	if err != nil {
		logger.Fatal(err.Error())
	}

	metad.Init()
	metad.Serve()
}
