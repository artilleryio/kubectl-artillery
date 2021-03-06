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
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/artilleryio/kubectl-artillery/internal/artillery"
	"github.com/artilleryio/kubectl-artillery/internal/telemetry"
	"github.com/posthog/posthog-go"
	"github.com/spf13/cobra"
	k8sValidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const generatetestExample = `- $ %[1]s generate <job-name> --script path/to/test-script
- $ %[1]s generate <job-name> -s path/to/test-script
- $ %[1]s generate <job-name> -s path/to/test-script [--namespace] [--out ] [--count ]`

// newCmdGenerate creates the "generate" test command
func newCmdGenerate(
	workingDir string,
	io genericclioptions.IOStreams,
	cliName string,
	tClient posthog.Client,
	tCfg telemetry.Config,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate [OPTIONS]",
		Aliases: []string{"gen"},
		Short:   "Generates a k8s Job packaged with Kustomize to execute a test",
		Example: fmt.Sprintf(generatetestExample, cliName),
		RunE:    makeRunGenTest(workingDir, io, tCfg),
		PostRunE: func(cmd *cobra.Command, args []string) error {
			testScriptPath, _ := cmd.Flags().GetString("script")
			ns, _ := cmd.Flags().GetString("namespace")
			outPath, _ := cmd.Flags().GetString("out")
			count, _ := cmd.Flags().GetInt("count")

			logger := artillery.NewIOLogger(io.Out, io.ErrOut)
			telemetry.TelemeterGenerateManifests(args[0], testScriptPath, ns, outPath, count, tClient, tCfg, logger)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringP(
		"script",
		"s",
		"",
		"Specify path to artillery test-script file",
	)

	flags.StringP(
		"namespace",
		"n",
		"default",
		"Optional. Specify a namespace to apply your Job and related manifests",
	)

	flags.StringP(
		"out",
		"o",
		"",
		"Optional. Specify output path to write Job and related manifests",
	)

	flags.IntP(
		"count",
		"c",
		1,
		"Optional. Specify how many test workers the created Job should run",
	)

	if err := cmd.MarkFlagRequired("script"); err != nil {
		return nil
	}

	return cmd
}

// makeRunGenTest creates the RunE function used to generate a test
func makeRunGenTest(workingDir string, io genericclioptions.IOStreams, cfg telemetry.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := validateTest(args); err != nil {
			return err
		}

		testScriptPath, err := cmd.Flags().GetString("script")
		if err != nil {
			return err
		}

		if err := validateTestScriptExists(testScriptPath); err != nil {
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

		count, err := cmd.Flags().GetInt("count")
		if err != nil {
			return err
		}

		testName := args[0]
		configMapName := fmt.Sprintf("%s-test-script", testName)

		targetDir, err := artillery.MkdirAllTargetOrDefault(workingDir, outPath, artillery.DefaultManifestDir)
		if err != nil {
			return err
		}

		if err := artillery.CopyFileTo(targetDir, testScriptPath); err != nil {
			return err
		}

		job := artillery.NewTestJob(testName, ns, configMapName, filepath.Base(testScriptPath), count, cfg)
		kustomization := artillery.NewKustomization(artillery.TestFilename, ns, configMapName, testScriptPath, artillery.LabelPrefix)

		msg, err := artillery.Generatables{
			{
				Path:      filepath.Join(targetDir, artillery.TestFilename),
				Marshaler: job,
			},
			{
				Path:      filepath.Join(targetDir, "kustomization.yaml"),
				Marshaler: kustomization,
			},
		}.Generate(2)
		if err != nil {
			return err
		}

		_, _ = io.Out.Write([]byte(msg))
		_, _ = io.Out.Write([]byte("\n"))
		return nil
	}
}

// validateTest validates test RunE arguments.
// Including,
// - Extra supplied arguments
// - Missing test name
// - Invalid named test
func validateTest(args []string) error {
	if len(args) == 0 {
		return errors.New("missing test name")
	}
	if len(args) > 1 {
		return errors.New("unknown arguments detected")
	}

	testName := args[0]
	invalids := k8sValidation.NameIsDNSSubdomain(testName, false)
	if len(invalids) > 0 {
		return fmt.Errorf("test name %s must be a valid DNS subdomain name, \n%s", testName, strings.Join(invalids, "\n- "))
	}

	return nil
}

// validateTestScriptExists validates the test script file exists.
func validateTestScriptExists(s string) error {
	absPath, err := filepath.Abs(s)
	if err != nil {
		return err
	}

	if !artillery.DirOrFileExists(absPath) {
		return fmt.Errorf("cannot find script file %s ", s)
	}

	return nil
}
