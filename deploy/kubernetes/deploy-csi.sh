#!/bin/bash

objects=(cvmfs-storage-class cvmfsplugin csi-attacher csi-provisioner)

for obj in ${objects[@]}; do
	kubectl create -f "./$obj.yaml"
done
