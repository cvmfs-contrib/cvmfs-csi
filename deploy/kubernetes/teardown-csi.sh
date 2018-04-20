#!/bin/bash

objects=(cvmfsplugin csi-provisioner csi-attacher cvmfs-storage-class)

for obj in ${objects[@]}; do
	kubectl delete -f "./$obj.yaml"
done
