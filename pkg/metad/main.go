// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// Copyright 2018 Yunify Inc. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package metad

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/gops/agent"

	"openpitrix.io/metad/pkg/logger"
	"openpitrix.io/metad/pkg/version"
)

func Main() {
	flag.Parse()

	if printVersion {
		fmt.Println(version.GetVersionString())
		os.Exit(0)
	}

	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}

	var config *Config
	var err error
	if config, err = initConfig(); err != nil {
		logger.Fatal("%v", err)
	}

	logger.Info("Starting metad %s", version.ShortVersion)
	metad, err = New(config)
	if err != nil {
		logger.Fatal(err.Error())
	}

	metad.Init()
	metad.Serve()
}
