# CernVM-FS CSI driver

csi-cvmfs provides read-only mounting of CVMFS volumes in CSI-enabled container orchestrators.

## Compiling

The CSI CernVM-FS driver can be compiled in a form of a binary file or an image. When compiled as a binary, the resulting file gets stored in `_output` directory. When compiled as an image, it gets stored in local Docker image store.

Building a binary file: `$ make cvmfsplugin`
Building a Docker image: `$ make image`

## Deployment

**Kubernetes 1.10**

Enable CSI in Kubernetes:

- kube-apiserver must be launched with `--feature-gates=CSIPersistentVolume=true,MountPropagation=true` and `--runtime-config=storage.k8s.io/v1alpha1=true`
- kube-controller-manager must be launched with `--feature-gates=CSIPersistentVolume=true`
- kubelet must be launched with `--feature-gates=CSIPersistentVolume=true,MountPropagation=true`

Deploy the external attacher and provisioner sidecar containers as StatefulSets:

```bash
$ kubectl create -f deploy/kubernetes/csi-cvmfsplugin-attacher-rbac.yaml
$ kubectl create -f deploy/kubernetes/csi-cvmfsplugin-attacher.yaml
```
```bash
$ kubectl create -f deploy/kubernetes/csi-cvmfsplugin-provisioner-rbac.yaml
$ kubectl create -f deploy/kubernetes/csi-cvmfsplugin-provisioner.yaml
```

Deploy the driver-registrar sidecar container and the csi-cvmfsplugin as DaemonSet:

```bash
$ kubectl create -f deploy/kubernetes/csi-cvmfsplugin-rbac.yaml
$ kubectl create -f deploy/kubernetes/csi-cvmfsplugin-.yaml
```

## StorageClass parameters

`repository`: mandatory, CVMFS repository address

`tag`: optional, defaults to `trunk`

`hash`: optional

Specifying both `tag` and `hash` is not allowed.

### CVMFS configuration

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
