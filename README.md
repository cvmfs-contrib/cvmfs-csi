# CernVM-FS CSI driver

Currently supports only Kubernetes 1.10+

## StorageClass parameters

`repository`: mandatory, CVMFS repository address

`tag`: optional, defaults to `trunk`

`hash`: optional

Specifying both `tag` and `hash` is not allowed.

## Deployment

Deploy `external-attacher`, `external-provisioner`, `driver-registrar` sidecar containers and the `csi-cvmfsplugin`:

```bash
$ ./deploy/kubernetes/csi-deploy.sh
```

Create the csi-cvmfs storage class, PVC and a Pod:

```bash
$ ./deploy/kubernetes/user-deploy.sh
```

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
