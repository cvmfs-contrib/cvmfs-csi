# CVMFS CSI driver

[![Build Status](https://github.com/cernops/cvmfs-csi/workflows/csi-cvmfsplugin/badge.svg?event=push&branch=master)](https://github.com/cernops/cvmfs-csi/actions?workflow=csi-cvmfsplugin)
[![Go Report Card](https://goreportcard.com/badge/github.com/cernops/cvmfs-csi)](https://goreportcard.com/report/github.com/cernops/cvmfs-csi)
[![GoDoc](https://godoc.org/github.com/cernops/cvmfs-csi?status.svg)](https://godoc.org/github.com/cernops/cvmfs-csi)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

csi-cvmfs provides read-only mounting of CVMFS volumes in CSI-enabled container orchestrators.

## Building

To build the binary in ./bin:
```
make build
```

To build binaries for multiple platforms:
```
make build-cross
```

To make the docker image:
```
make docker
```

## Configuration

**Available command line arguments:**

Option | Default value | Description
------ | ------------- | -----------
`--endpoint` | `unix://tmp/csi.sock` | CSI endpoint, must be a UNIX socket
`--drivername` | `csi-cvmfsplugin` | name of the driver (Kubernetes: `provisioner` field in StorageClass must correspond to this value)
`--nodeid` | _empty_ | This node's ID
`--cvmfsCacheRoot` | `/var/cache/cvmfs` | local CVMFS cache path

**Available volume parameters:**

Parameter | Required | Description
--------- | -------- | -----------
`repository` | yes | Address of the CVMFS repository
`tag` | no | `CVMFS_REPOSITORY_TAG`. Defaults to `trunk`
`hash` | no | `CVMFS_REPOSITORY_HASH`
`proxy` | no | `CVMFS_HTTP_PROXY`. Defaults to the value sourced from `default.local`. See instructions below.

By default, csi-cvmfs is distributed with `default.local` containing CERN defaults. You can override those at runtime by overwriting `/etc/cvmfs/default.local`, which is then sourced into any later CVMFS client configs used for mounting.


**CVMFS configuration in Kubernetes**
You can use Kubernetes [config map](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/) to inject a custom `default.local` file.

```bash
$ cat my-config
CVMFS_HTTP_PROXY="http://my-cvmfs-proxy:3128"
CVMFS_QUOTA_LIMIT=5000
$ kubectl create configmap my-configmap --from-file=cvmfs-override=./my-config
```
Edit the manifest for `csi-cvmfsplugin`:
In the `volumes` section, add an entry for the config map:
```yaml
volumes:
  - name: my-cvmfs-config
    configMap:
      name: my-configmap
```
Add a volume mount for the `csi-cvmfsplugin` container:
```yaml
volumeMounts:
  - name: my-cvmfs-config
    mountPath: /etc/cvmfs/default.local
    subPath: cvmfs-override
```

## Deployment with Kubernetes

Requires Kubernetes 1.11+

Your Kubernetes cluster must allow privileged pods (i.e. `--allow-privileged` flag must be set to true for both the API server and the kubelet). Moreover, as stated in the [mount propagation docs](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation), the Docker daemon of the cluster nodes must allow shared mounts.

YAML manifests are located in `deploy/kubernetes`.

**Deploy RBACs for sidecar containers and node plugins:**

```bash
$ kubectl create -f namespace.yaml
$ kubectl create -f csi-attacher-rbac.yaml
$ kubectl create -f csi-provisioner-rbac.yaml
$ kubectl create -f csi-nodeplugin-rbac.yaml
```

Those manifests deploy service accounts, cluster roles and cluster role bindings.

**Deploy CSI sidecar containers:**

```bash
$ kubectl create -f csi-cvmfsplugin-attacher.yaml
$ kubectl create -f csi-cvmfsplugin-provisioner.yaml
```

Deploys stateful sets for external-attacher and external-provisioner sidecar containers for CSI CernVM-FS.

**Deploy the CSI CernVM-FS driver:**

```bash
$ kubectl create -f csi-cvmfsplugin.yaml
```

Deploys a daemon set with two containers: CSI driver-registrar and the CSI CernVM-FS driver.

## Verifying the deployment in Kubernetes

After successfuly completing the steps above, you should see output similar to this:
```bash
$ kubectl get all --namespace=cvmfs
NAME                                READY     STATUS    RESTARTS   AGE
pod/csi-cvmfsplugin-attacher-0      1/1       Running   0          1m
pod/csi-cvmfsplugin-bhqck           2/2       Running   0          1m
pod/csi-cvmfsplugin-provisioner-0   1/1       Running   0          1m

NAME                                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
service/csi-cvmfsplugin-attacher      ClusterIP   10.104.25.190   <none>        12345/TCP   1m
service/csi-cvmfsplugin-provisioner   ClusterIP   10.101.197.42   <none>        12345/TCP   1m

NAME                             DESIRED   CURRENT   READY     UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/csi-cvmfsplugin   1         1         1         1            1           <none>          1m

NAME                                           DESIRED   CURRENT   AGE
statefulset.apps/csi-cvmfsplugin-attacher      1         1         1m
statefulset.apps/csi-cvmfsplugin-provisioner   1         1         1m

...
```

You can try deploying a demo pod from `examples/` to test the deployment further.
