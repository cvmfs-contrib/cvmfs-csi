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
