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

package artillery

import (
	"encoding/json"

	"k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Job struct {
	*v1.Job
}

// NewTestJob returns a configured Kubernetes Job wrapper for an Artillery  test.
func NewTestJob(testName, configMapName, testScriptFilename string, count int) *Job {
	var (
		parallelism  int32 = 1
		completions  int32 = 1
		backoffLimit int32 = 0
	)

	if count > 0 {
		parallelism = int32(count)
		completions = int32(count)
	}

	j := &Job{
		Job: &v1.Job{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Job",
				APIVersion: "batch/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      testName,
				Namespace: "default",
				Labels:    labels(testName, "test-worker-master"),
			},
			Spec: v1.JobSpec{
				Parallelism:  &parallelism,
				Completions:  &completions,
				BackoffLimit: &backoffLimit,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels(testName, "test-worker"),
					},

					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:            testName,
								Image:           WorkerImage,
								ImagePullPolicy: corev1.PullAlways,
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      JobTestScriptVol,
										MountPath: "/data",
									},
								},
								Args: []string{
									"run",
									"/data/" + testScriptFilename,
								},
								Env: append(
									[]corev1.EnvVar{
										// published metrics use WORKER_ID to connect the pod (worker) to a Pushgateway JobID
										// Uses the downward API:
										// https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/#the-downward-api
										{
											Name: "WORKER_ID",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "metadata.name",
												},
											},
										},
									},
								),
							},
						},
						// Provides access to the ConfigMap holding the test script config
						Volumes: []corev1.Volume{
							{
								Name: JobTestScriptVol,
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: configMapName,
										},
									},
								},
							},
						},
						RestartPolicy: "Never",
					},
				},
			},
		},
	}
	return j
}

// labels creates K8s labels used to organize
// and categorize (scope and select) test jobs.
func labels(name string, component string) map[string]string {
	return map[string]string{
		"artillery.io/test-name": name,
		"artillery.io/component": component,
		"artillery.io/part-of":   LabelPrefix,
	}
}

func (j *Job) MarshalWithIndent(indent int) ([]byte, error) {
	data, err := j.json()
	if err != nil {
		return nil, err
	}

	y, err := jsonToYaml(data, indent)
	if err != nil {

		return nil, err
	}

	return y, nil
}

func (j *Job) json() ([]byte, error) {
	data, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}
	delete(temp, "status")
	delete(temp["metadata"].(map[string]interface{}), "creationTimestamp")
	delete(temp["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{}), "creationTimestamp")

	return json.Marshal(temp)
}
