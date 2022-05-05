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

package telemetry

import (
	"github.com/go-logr/logr"
	"github.com/posthog/posthog-go"
)

// TelemeterServicesScaffold enqueues a kubectl-artillery scaffold command event.
func TelemeterServicesScaffold(
	serviceNames []string,
	namespace, outPath string,
	tClient posthog.Client,
	tConfig Config,
	logger logr.Logger,
) {
	if err := enqueue(
		tClient,
		tConfig,
		event{
			Name: "kubectl-artillery scaffold",
			Properties: map[string]interface{}{
				"source":           "kubectl-artillery-plugin",
				"serviceCount":     len(serviceNames),
				"namespace":        hashEncode(namespace),
				"defaultOutputDir": len(outPath) == 0,
			},
		},
		logger,
	); err != nil {
		logger.Error(err,
			"could not broadcast telemetry",
			"telemetry disable", tConfig.Disable,
			"telemetry debug", tConfig.Debug,
			"event", "kubectl-artillery scaffold",
		)
	}
}

// TelemeterGenerateManifests enqueues a kubectl-artillery generate command event.
func TelemeterGenerateManifests(
	name, testScriptPath, env, outPath string,
	count int,
	tClient posthog.Client,
	tConfig Config,
	logger logr.Logger,
) {
	if err := enqueue(
		tClient,
		tConfig,
		event{
			Name: "kubectl-artillery generate",
			Properties: map[string]interface{}{
				"source":           "kubectl-artillery-plugin",
				"name":             hashEncode(name),
				"testScript":       hashEncode(testScriptPath),
				"count":            count,
				"environment":      hashEncode(env),
				"defaultOutputDir": len(outPath) == 0,
			},
		},
		logger,
	); err != nil {
		logger.Error(err,
			"could not broadcast telemetry",
			"telemetry disable", tConfig.Disable,
			"telemetry debug", tConfig.Debug,
			"event", "kubectl-artillery generate",
		)
	}
}
