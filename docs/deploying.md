# Deploying CVMFS CSI driver in Kubernetes

CVMFS CSI deployment consists of a DaemonSet node plugin that handles node local mount-unmount operations, and ConfigMaps storing CVMFS client configuration.

Cluster administrators may deploy CVMFS CSI manually using the provided Kubernetes manifests, or by installing cvmfs-csi Helm chart.

After successful deployment, you can try examples in [../example/](../example/).

## Manual deployment using manifests

Kubernetes manifests are located in [../deployments/kubernetes](../deployments/kubernetes). They define a node plugin DaemonSet, controller plugin Deployment, ConfigMaps and a CSIDriver object. Deploy them using the following command:

```bash
kubectl create -f deployments/kubernetes
```

You may need to customize `cvmfs-csi-default-local` and `cvmfs-csi-config-d` ConfigMaps defined in [../deployments/kubernetes/configmap-cvmfs-client.yaml](../deployments/kubernetes/cvmfs-client-configmap.yaml) to suite your CVMFS environment.

## Deployment with Helm chart

Helm chart can be installed from CERN registry:

```bash
helm repo add cern https://registry.cern.ch/chartrepo/cern
helm repo update
helm install cvmfs cern/cvmfs-csi
```

Some chart values may need to be customized to suite your CVMFS environment. Please consult the documentation in [../deployments/helm/README.md](../deployments/helm/README.md) to see available values.

## Verifying the deployment

After successful deployment, you should see similar output from `kubectl get all -l app=cvmfs-csi`:

```
$ kubectl get all -l app=cvmfs-csi
NAME                                                READY   STATUS    RESTARTS   AGE
pod/c-cvmfs-csi-controllerplugin-5b44968dc9-jb2ms   2/2     Running   0          90m
pod/c-cvmfs-csi-nodeplugin-t6lvc                    2/2     Running   0          90m
pod/cvmfs-csi-nodeplugin-rgxkh                      2/2     Running   0          90m

NAME                                    DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/c-cvmfs-csi-nodeplugin   2         2         2       2            2           <none>          40h

NAME                                           READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/c-cvmfs-csi-controllerplugin   1/1     1            1           90m

NAME                                                      DESIRED   CURRENT   READY   AGE
replicaset.apps/c-cvmfs-csi-controllerplugin-5b44968dc9   1         1         1       90m
```

## csi-cvmfsplugin command line arguments

CVMFS CSI driver executable accepts following set of command line arguments:

|Name|Default value|Description|
|--|--|--|
|`--endpoint`|`unix:///var/lib/kubelet/plugins/cvmfs.csi.cern.ch/csi.sock`|(string value) CSI endpoint. CVMFS CSI will create a UNIX socket at this location.|
|`--drivername`|`cvmfs.csi.cern.ch`|(string value) Name of the driver that is used to link PersistentVolume objects to CVMFS CSI driver.|
|`--nodeid`|_none, required_|(string value) Unique identifier of the node on which the CVMFS CSI node plugin pod is running. Should be set to the value of `Pod.spec.nodeName`.|
|`--has-alien-cache`|_false_|(boolean value) CVMFS client is using alien cache volume. The volume will be `chmod`'d with correct permissions.|
|`--start-automount-daemon`|_true_|(boolean value) Whether CVMFS CSI nodeplugin Pod should run automount daemon. This is required for automounts to work. If however worker nodes are already running automount daemon (e.g. as a systemd service), you may disable running yet another instance of the daemon using this switch.|
|`automount-startup-timeout`|_5_|number of seconds to wait for automount daemon to start up before exiting|
|`automount-unmount-timeout`|_-1_|number of seconds of idle time after which an autofs-managed CVMFS mount will be unmounted. '0' means never unmount, '-1' leaves automount default option.|
|`--role`|_none, required_|Enable driver service role (comma-separated list or repeated `--role` flags). Allowed values are: `identity`, `node`, `controller`.|
|`--version`|_false_|(boolean value) Print driver version and exit.|
