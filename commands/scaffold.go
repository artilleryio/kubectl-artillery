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
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/artilleryio/kubectl-artillery/internal/artillery"
	"github.com/artilleryio/kubectl-artillery/internal/kube"
	"github.com/artilleryio/kubectl-artillery/internal/telemetry"
	"github.com/posthog/posthog-go"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const scaffoldExample = `- $ %[1]s scaffold <k8s-Service-name> 
- $ %[1]s scaffold <k8s-service1> <k8s-service2>
- $ %[1]s scaffold <k8s-Service-name> [--namespace ] [--out ]`

// newCmdScaffold creates the test script scaffold command
func newCmdScaffold(
	workingDir string,
	io genericclioptions.IOStreams,
	cliName string,
	tClient posthog.Client,
	tCfg telemetry.Config,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "scaffold [OPTIONS]",
		Short:   "Scaffolds test scripts from K8s services using liveness probe HTTP endpoints",
		Example: fmt.Sprintf(scaffoldExample, cliName),
		RunE:    makeRunScaffold(workingDir, io),
		PostRunE: func(cmd *cobra.Command, args []string) error {
			ns, _ := cmd.Flags().GetString("namespace")
			outPath, _ := cmd.Flags().GetString("out")

			logger := artillery.NewIOLogger(io.Out, io.ErrOut)
			telemetry.TelemeterServicesScaffold(args, ns, outPath, tClient, tCfg, logger)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringP(
		"namespace",
		"n",
		"default",
		"Optional. Specify a namespace for your services",
	)

	flags.StringP(
		"out",
		"o",
		"",
		"Optional. Specify output path to write the test script files",
	)

	return cmd
}

// makeRunScaffold creates the RunE function used to scaffold a test script
func makeRunScaffold(workingDir string, io genericclioptions.IOStreams) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := validateScaffold(args); err != nil {
			return err
		}

		ns, err := cmd.Flags().GetString("namespace")
		if err != nil {
			return err
		}

		outPath, err := cmd.Flags().GetString("out")
		if err != nil {
			return err
		}

		targetDir, err := artillery.MkdirAllTargetOrDefault(workingDir, outPath, artillery.DefaultScriptsDir)
		if err != nil {
			return err
		}

		ctl, err := kube.NewClient(genericclioptions.NewConfigFlags(true))
		if err != nil {
			return err
		}

		if len(ns) == 0 {
			ns = ctl.CfgNamespace
		}

		queryResults, err := kube.DoQuery(context.TODO(), args, ns, ctl)
		if err != nil {
			return err
		}

		for _, qr := range queryResults.QueryMisses() {
			_, _ = io.Out.Write([]byte(fmt.Sprintf("services \"%s\" not found\n", qr.QueriedServiceName())))
		}

		if !queryResults.HasQueryHits() {
			return nil
		}

		for _, qr := range queryResults.LivenessMisses() {
			svc := qr.SelectionServiceName()
			_, _ = io.Out.Write([]byte(fmt.Sprintf("services \"%s\" has no liveness probe endpoints, or ports mapping to endpoints\n", svc)))
		}

		if !queryResults.HasLivenessHits() {
			return nil
		}

		var scripts artillery.Generatables
		for _, result := range queryResults.LivenessHits() {
			ts := artillery.NewTestScript(result.ServiceProbes())
			scripts = append(scripts, artillery.Generatable{
				Path:      filepath.Join(targetDir, fmt.Sprintf("test-script_%s.yaml", result.SelectionServiceName())),
				Marshaler: ts,
			})
		}
		msg, err := scripts.Generate(2)
		if err != nil {
			return err
		}

		_, _ = io.Out.Write([]byte(msg + "\n"))

		return nil
	}
}

// validateScaffold validates scaffold command arguments
func validateScaffold(args []string) error {
	if len(args) == 0 {
		return errors.New("missing service name or names")
	}

	return nil
}
