#!/bin/bash

objects=(csi-cvmfsplugin-attacher-rbac csi-cvmfsplugin-provisioner-rbac csi-cvmfsplugin-rbac csi-cvmfsplugin-attacher csi-cvmfsplugin-provisioner csi-cvmfsplugin)

for obj in "${objects[@]}"; do
	kubectl create -f "./$obj.yaml"
done
