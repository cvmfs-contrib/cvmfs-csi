# CVMFS-CSI Helm Chart

This is an initial version of a CVMFS-CSI helm chart.

## Basic Usage

1. Change the configuration options in the `values.yaml` file for your CVMFS setup.
In principle, the chart should be fully configurable through the `values.yaml` file.
For more information on the different possible ways to specify values, see: https://helm.sh/docs/helm/#helm-install

2. To install the chart with the release name `my-release` in the `cvmfs` namespace:
```
helm install --name my-release --namespace cvmfs -f mynewvalues.yaml ./cvmfs-csi
```
If no name is specified, helm will create a random release name.

3. To delete a release:
```
helm delete --purge my-release
```

For more advanced options and information, visit the helm docs:
https://helm.sh/docs/helm/
