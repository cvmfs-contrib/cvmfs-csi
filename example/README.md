## How to test CernVM-FS CSI plugin with Kubernetes 1.11

You can use `plugin-deploy.sh` and `plugin-teardown.sh` helper scripts to help you deploy/tear down RBACs, sidecar containers and the plugin in one go. By default, they look for the YAML manifests in `../deploy/kubernetes`. You can override this path by running e.g. `$ ./plugin-deploy.sh /path/to/my/manifests`.

Once the plugin is successfuly deployed, you'll need to customize `storageclass.yaml` manifest to reflect your CVMFS setup. Please consult the documentation for info about available parameters.

After configuring the secrets, monitors, etc. you can deploy a testing Pod mounting a CVMFS volume:
```bash
$ kubectl create -f storageclass.yaml
$ kubectl create -f pvc.yaml
$ kubectl create -f pod.yaml
```

Other helper scripts:
* `logs.sh` output of the plugin
* `exec-bash.sh` logs into the plugin's container and runs bash
