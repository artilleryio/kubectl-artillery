/*
 * Copyright (c) 2022.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0.
 *
 * If a copy of the MPL was not distributed with
 * this file, You can obtain one at
 *
 *   http://mozilla.org/MPL/2.0/
 */

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/artilleryio/kubectl-artillery/commands"
	"github.com/artilleryio/kubectl-artillery/internal/telemetry"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	// AppName the controller and CLI app name.
	AppName = "kubectl-artillery-plugin"

	// Version controller version.
	Version = "alpha"

	// WorkerImage the Artillery image used by workers to run load tests.
	WorkerImage = "artilleryio/artillery:latest"
)

// kubectl-artillery CLI entrypoint
func main() {
	wd := "."
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	tCfg := telemetry.NewConfig(AppName, Version, WorkerImage, nil)
	tClient, err := telemetry.NewClient(tCfg)
	if err != nil {
		_, _ = ioStreams.ErrOut.Write([]byte(fmt.Sprintf("unable to create telemetry client: %s", err.Error())))
		os.Exit(1)
	}
	defer doClose(tClient, "could not close Posthog telemetry client", ioStreams)

	root := commands.NewCmdArtillery(wd, ioStreams, tClient, tCfg)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func doClose(closer io.Closer, msg string, ioStreams genericclioptions.IOStreams) {
	if err := closer.Close(); err != nil {
		_, _ = ioStreams.ErrOut.Write([]byte(fmt.Sprintf("%s: %s", msg, err.Error())))
	}
}
