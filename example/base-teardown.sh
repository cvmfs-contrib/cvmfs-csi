#!/bin/bash

objects=(csi-cvmfsplugin-attacher csi-cvmfsplugin-provisioner csi-cvmfsplugin csi-cvmfsplugin-attacher-rbac csi-cvmfsplugin-provisioner-rbac csi-cvmfsplugin-rbac)

for obj in "${objects[@]}"; do
	kubectl delete -f "./$obj.yaml"
done
