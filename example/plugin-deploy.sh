#!/bin/bash

deployment_base="${1}"

if [[ -z $deployment_base ]]; then
	deployment_base="../deploy/kubernetes"
fi

cd "$deployment_base" || exit 1

objects=(namespace csi-attacher-rbac csi-provisioner-rbac csi-nodeplugin-rbac csi-cvmfsplugin-attacher csi-cvmfsplugin-provisioner csi-cvmfsplugin)

for obj in ${objects[@]}; do
	kubectl create -f "./$obj.yaml"
done
