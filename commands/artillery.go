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

package commands

import (
	"fmt"
	"io"

	"github.com/artilleryio/kubectl-artillery/internal/artillery"
	"github.com/artilleryio/kubectl-artillery/internal/telemetry"
	"github.com/posthog/posthog-go"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var Version = "v0.2.0"

//cliName, the name of the CLI including kubectl as a root
const cliName = "kubectl artillery"

// NewCmdArtillery creates the kubectl-artillery CLI root command
func NewCmdArtillery(
	workingDir string,
	io genericclioptions.IOStreams,
	tClient posthog.Client,
	tCfg telemetry.Config,
) *cobra.Command {

	cmd := &cobra.Command{
		Short:        "Bootstrap artillery.io testing on Kubernetes",
		Use:          "artillery",
		Version:      Version,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return highlightTelemetryIfRequired(io.Out)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("%q is not a %[1]s command\nSee '%[1]s --help'", args[0], cliName)
			}
			cmd.HelpFunc()(cmd, args)
			return nil
		},
	}

	cmd.AddCommand(newCmdScaffold(workingDir, io, cliName, tClient, tCfg))
	cmd.AddCommand(newCmdGenerate(workingDir, io, cliName, tClient, tCfg))

	return cmd
}

// highlightTelemetryIfRequired informs CLI users whether telemetry is on or not.
// Will only display telemetry status on first run
func highlightTelemetryIfRequired(out io.Writer) error {
	settings, err := artillery.GetOrCreateCLISettings()
	if err != nil {
		return err
	}

	if !settings.GetFirstRun() {
		return nil
	}

	_, _ = out.Write([]byte("Telemetry is on. Learn more: https://artillery.io/docs/resources/core/telemetry.html\n"))

	return settings.SetFirstRun(false).Save()
}
