# Kubectl Artillery plugin

> Bootstrap [Artillery](https://www.artillery.io) testing on Kubernetes with this kubectl plugin.

## Use cases

- Scaffold test scripts from existing Kubernetes [Services](https://kubernetes.io/docs/concepts/services-networking/service/).

- Generate testing Jobs to run on Kubernetes using existing or scaffolded test scripts.

## Installation

### Download the binary

Download the binary from [GitHub Releases](https://github.com/artilleryio/kubectl-artillery/releases) and drop it in
your `$PATH`.

#### Linux

```shell
curl -L -o kubectl-artillery.tar.gz https://github.com/artilleryio/kubectl-artillery/releases/download/v0.2.0/kubectl-artillery_0.2.0_linux_amd64_2022-04-04T15.07.18Z.tar.gz
tar -xvf kubectl-artillery.tar.gz
sudo mv kubectl-artillery /usr/local/bin
```

#### Darwin(amd64)

```shell
curl -L -o kubectl-artillery.tar.gz https://github.com/artilleryio/kubectl-artillery/releases/download/v0.2.0/kubectl-artillery_v0.2.0_darwin_amd64_2022-04-04T15.07.18Z.tar.gz
tar -xvf kubectl-artillery.tar.gz
sudo mv kubectl-artillery /usr/local/bin
```

#### Darwin(arm64)

```shell
curl -L -o kubectl-artillery.tar.gz https://github.com/artilleryio/kubectl-artillery/releases/download/v0.2.0/kubectl-artillery_v0.2.0_darwin_arm64_2022-04-04T15.07.18Z.tar.gz
tar -xvf kubectl-artillery.tar.gz
sudo mv kubectl-artillery /usr/local/bin
```

### Verify installation

You can verify its installation using `kubectl`:

```shell
$ kubectl plugin list
#The following kubectl-compatible plugins are available:

# /usr/local/bin/kubectl-artillery
```

Validate if `kubectl artillery` can be executed.

```bash
$ kubectl artillery --help
# Bootstrap artillery.io testing on Kubernetes

# Usage:
#   artillery [flags]
#   artillery [command]

# Available Commands:
#   ...
#   generate    Generates a k8s Job packaged with Kustomize to execute a test
#   ...
#   scaffold    Scaffolds test scripts from K8s services using liveness probe HTTP endpoints

# Flags:
#   -h, --help      help for artillery
#   -v, --version   version for artillery

# Use "artillery [command] --help" for more information about a command.
```

## How it works

The plugin provides two sub commands:

- [scaffold](#scaffold)
- [generate](#generate)

### scaffold

Use the `scaffold` subcommand to
scaffold [test scripts](https://www.artillery.io/docs/guides/guides/test-script-reference)
from existing K8s [Services](https://kubernetes.io/docs/concepts/services-networking/service/).

Created test scripts use
the [expect plugin](https://www.artillery.io/docs/guides/plugins/plugin-expectations-assertions)
to functionally
test [HTTP liveness probe endpoints](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-http-request)
in Pods proxied by the supplied services.

Once created, update the tests to match your requirements. You can also use created test scripts to generate tests.

See [Example: scaffold test scripts](#example-scaffold-test-scripts).

#### Output is configurable

By default, all test scripts will be written to an `artillery-scripts` directory. This will be located in the same path
where the `scaffold` subcommand was run.

Use the `--out/-o` flag to specify a different directory path to write the test scripts.

#### A target url for every test

A Kubernetes Service may reference multiple ports, requiring multiple `target` urls. Created test scripts work around
this by using a full urls for every test endpoint.

For example,

```yaml
...
scenarios:
  - flow:
      - get:
          url: http://nginx-probes-mapped:80/
          expect:
            - statusCode: 200
...
```

#### Some liveness probes cannot be tested

A Pod may define a liveness probe on a port not accessible to the proxying Service. Such a liveness a probe cannot be
tested.

The plugin cannot scaffold a test script for a Service that has no access to the proxied Pod's liveness probes.

### Example: scaffold test scripts

This example will test an Nginx server running on K8s. The related deployment will be configured with an HTTP 
liveness probe running on port 80.

All the related manifests are [found here](https://raw.githubusercontent.com/artilleryio/kubectl-artillery/main/examples/has-probes/has-probes.yaml).

```shell
$ kubectl apply -f https://raw.githubusercontent.com/artilleryio/kubectl-artillery/es-generate-rework/examples/has-probes/has-probes.yaml
# deployment.apps/k8s-probes-mapped created
# service/nginx-probes-mapped created
```

Let's check that our service `nginx-probes-mapped` has related Pods with defined HTTP liveness probes. A
service's `Selector` field helps us identify the correct Pods.

```shell
kubectl describe service nginx-probes-mapped
# Name:                     nginx-probes-mapped
# Namespace:                default
# ...
# Selector:                 app=nginx-probes-mapped <<<<<
# ...
# Port:                     nginx-http-port  80/TCP
# TargetPort:               80/TCP
# ...
```

```shell
kubectl get pods --selector=app=nginx-probes-mapped
# NAME                                 READY   STATUS    RESTARTS   AGE
# k8s-probes-mapped-64998cbdf5-7cqzg   1/1     Running   0          11m
```

```shell
kubectl get pods <found-pod-object> -o yaml
# apiVersion: v1
# kind: Pod
# ...
# spec:
#  containers:
#  - image: nginx
#    imagePullPolicy: Always
#    livenessProbe:
#      failureThreshold: 1
#      httpGet:
#        path: /
#        port: 80
#        scheme: HTTP
#      initialDelaySeconds: 1
...
```

The answer is YES. `nginx-probes-mapped` can access the Pod's HTTP liveness probe.

```shell
kubectl artillery scaffold nginx-probes-mapped
# artillery-scripts/test-script_nginx-probes-mapped.yaml generated
```

Looking into the `artillery-scripts` directory reveals the generated test script YAML file.

```shell
ls -alh ./artillery-scripts
# total 8
# drwx------   3 xxx  xxx    96B  1 Apr 15:22 .
# drwxr-xr-x@ 34 xxx  xxx   1.1K  1 Apr 15:22 ..
# -rw-r--r--   1 xxx  xxx   305B  4 Apr 13:20 test-script_nginx-probes-mapped.yaml
````

You can edit the files as you please. Then use it to generate a test.

### generate

Use the `generate` subcommand to generate 
- A K8s Job that will run Artillery test workers.
- Related Kubernetes manifests (e.g. ConfigMap).

All packaged with a kustomization.yaml file.

See [Example: generate and apply Test](#example-generate-and-apply-test).

#### Output is configurable

By default, all manifests will be written to an `artillery-manifests` directory. This will be located in the same path
where the `generate` subcommand was run.

Use the `--out/-o` flag to specify a different directory path to write load test manifests and kustomization.yaml.

#### Test scripts are bundled

The `generate` subcommand also copies the artillery test-script file to the output directory. This is
because [Kustomize v2.0 added a security check](https://kubectl.docs.kubernetes.io/faq/kustomize/#security-file-foo-is-not-in-or-below-bar)
that prevents kustomizations from reading files outside their own directory root.

### Example: generate and apply Test

Using the test-script created by the [scaffold](#example-scaffold-test-scripts) command,

```shell
kubectl artillery gen probe -s artillery-scripts/test-script_nginx-probes-mapped.yaml
# artillery-manifests/test-job.yaml generated
# artillery-manifests/kustomization.yaml generated
```

Looking into the `artillery-manifests` directory reveals the generated manifests and bundled a copy of
the `test-script.yaml` file.

```shell
ls -alh ./artillery-manifests
# total 24
# drwx------   5 xxx  xxx   160B 18 Mar 15:40 .
# drwxr-xr-x@ 34 xxx  xxx   1.1K 22 Mar 17:28 ..
# -rw-r--r--   1 xxx  xxx   369B 23 Mar 14:11 kustomization.yaml
# -rw-r--r--   1 xxx  xxx   1.0K 23 Mar 14:11 test-job.yaml
# -rw-r--r--   1 xxx  xxx   302B 23 Mar 14:11 test-script_nginx-probes-mapped.yaml
```

You can edit the files as you please. And finally apply the generated Job to a Kubernetes cluster.

```shell
kubectl apply -k ./artillery-manifests
# configmap/probe-test-script created
# job.batch/probe created
```

The `generate` subcommand has created and configured a Kustomization.yaml file with a `configMapGenerator`. When applied, it has
generated `configmap/probe-test-script` which loads your Artillery test-script as a volume on Kubernetes.

```shell
kubectl describe configmap/probe-test-script
# Name:         probe-test-script
# Namespace:    default
# Labels:       artillery.io/component=artilleryio-test-config
#               artillery.io/part-of=artilleryio-test
# Annotations:  <none> | xargs command

# Data
# ====
# test-script_nginx-probes-mapped.yaml:
# ----
# ...
# config:
#   target: target: http://nginx-probes-mapped/
#   environments:
#      functional:
#        phases:
#          - duration: 1
#            arrivalCount: 1
#        plugins:
#          expect: {}
# scenarios:
#   - flow:
#      - get:
#       ...
#
# BinaryData
# ====
# 
# Events:  <none>
```

Finally, you can check that test Job has run as expected.

```shell
$ kubectl get job probe -o wide
# NAME    COMPLETIONS   DURATION   AGE     CONTAINERS   IMAGES                         SELECTOR
# probe   1/1           25s        6m17s   probe        artilleryio/artillery:latest   controller-uid=427f9384-feab-435c-a389-20a5cf217f27
```

You can see that it ran Artillery test worker Pod to completion.
```shell
$ kubectl get pods
# NAME          READY   STATUS      RESTARTS   AGE
# probe-nf7kt   0/1     Completed   0          11m
```

The test results can be inspected by checking the logs.
```shell
$ kubectl logs probe-nf7kt # to see the test results
#  Telemetry is on. Learn more: https://artillery.io/docs/resources/core/telemetry.html
# Phase started: unnamed (index: 0, duration: 1s) 12:54:18(+0000)

# Phase completed: unnamed (index: 0, duration: 1s) 12:54:19(+0000)

# --------------------------------------
# Metrics for period to: 12:54:20(+0000) (width: 0.12s)
# --------------------------------------

# http.codes.200: ................................................................ 1
# http.request_rate: ............................................................. 1/sec
... 
```

PS: you can also configure your test-script to publish
results to [Prometheus](https://www.artillery.io/docs/guides/plugins/plugin-publish-metrics#prometheus-pushgateway).

## License

The kubectl-artillery plugin is open-source software distributed under the terms of
the [MPLv2](https://www.mozilla.org/en-US/MPL/2.0/) license.
