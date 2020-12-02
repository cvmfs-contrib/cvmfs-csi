A Helm chart for the CVMFS-CSI driver, allowing the mounting of CVMFS repositories in Kubernetes environments. This chart will deploy the CSI driver as a DaemonSet, thus automatically scaling the driver on each cluster node, and will create a StorageClass for each mounted repository. These storage classes can subsequently be leveraged by PersistentVolumeClaims of other pods to mount corrensponding CVMFS repositories.

## Use

To use, clone this repository and install the chart:
```
git clone https://github.com/cernops/cvmfs-csi
```
Helm v2 Installation:
```
helm install --name cvmfs --namespace cvmfs cvmfs-csi/deployments/helm/cvmfs-csi/
```
Helm v3 Installation:
```
kubectl create namespace cvmfs
helm install cvmfs cvmfs-csi/deployments/helm/cvmfs-csi/ --namespace-cvmfs
```

## Configuration

The following table lists the configurable parameters of the CVMFS-CSI chart.

One can specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values of the parameters can be provided when installing the chart via `-f /path/to/myvalues.yaml`.


| Parameter                                    | Description                                                                                                                                                         |
|----------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `configs`                                    | Dictionary of configmaps to be mounted into the plugin container                                                                                                    |
| `repositories`                               | Dictionary of the repositories to mount. Each key will be used as the StorageClass name corresponding to the repository                                             |
| `cache.localCache.enabled`                   | Boolean representing whether the plugin should use a ReadWriteOnce volume for its cache. Ideal for single-node scenarios. Requires `cache.alienCache.enabled=false` |
| `cache.localCache.storageClass`              | Name of ReadWriteOnce storage class used to provision the local cache volume                                                                                        |
| `cache.localCache.size`                      | Size of local cache volume in Mi                                                                                                                                    |
| `cache.localCache.mountPath`                 | Path on plugin pods at which the local cache should be mounted                                                                                                      |
| `cache.alienCache.enabled`                   | Boolean representing whether the plugin should use a ReadWriteMany volume for its cache. Ideal for multi-node scenarios.  Requires `cache.localCache.enabled=false` |
| `cache.alienCache.storageClass`              | Name of ReadWriteMany storage class used to provision the alien cache volume                                                                                        |
| `cache.alienCache.size`                      | Size of alien cache volume in Mi                                                                                                                                    |
| `cache.alienCache.mountPath`                 | Path on plugin pods at which the alien cache should be mounted                                                                                                      |
| `cache.preload.enabled`                      | Boolean representing whether to deploy non-blocking K8S Job pods alongside the CSI, used to preload a subset of data into the cache                                 |
| `cache.preload.repositories`                 | List of repositories for which to preload data into the cache                                                                                                       |
| `cache.preload.repositories.[].name`         | Name of repository to preload. Must be a repository name. Each repository should only be listed once                                                                |
| `cache.preload.repositories.[].mountPath`    | Path at which to mount this repository on the Job pods                                                                                                              |
| `cache.preload.repositories.[].image`        | Name of image to run for the Job pod container                                                                                                                      |
| `cache.preload.repositories.[].commands`     | List of commands to run inside the Job pod container for this repository                                                                                            |
| `cache.preload.repositories.[].mountConfigs` | List of configs to mount inside the Job pod container for this repository. Must be list of keys from configs dictionary above                                       |
| `cache.preload.repositories.[].mountCache`   | Boolean indicating whether to mount the cache repositories into the Job container for this repository                                                               |
| `http_proxy`                                 | String indicating the http proxies that CVMFS should use, see <https://cvmfs.readthedocs.io/en/stable/cpt-configure.html#proxy-lists>                               |
