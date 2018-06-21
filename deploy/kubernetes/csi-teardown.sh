#!/bin/bash

objects=(cvmfsplugin csi-provisioner csi-attacher)

for obj in ${objects[@]}; do
	kubectl delete -f "./$obj.yaml"
done
