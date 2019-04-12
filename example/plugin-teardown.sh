#!/bin/bash

deployment_base="${1}"

if [[ -z $deployment_base ]]; then
	deployment_base="../deploy/kubernetes"
fi

cd "$deployment_base" || exit 1

objects=(csi-cvmfsplugin-attacher csi-cvmfsplugin-provisioner csi-cvmfsplugin csi-attacher-rbac csi-provisioner-rbac csi-nodeplugin-rbac namespace)

for obj in ${objects[@]}; do
	kubectl delete -f "./$obj.yaml"
done
