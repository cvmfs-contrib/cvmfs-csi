#!/bin/bash

objects=(csi-attacher csi-provisioner cvmfsplugin)

for obj in ${objects[@]}; do
	kubectl create -f "./$obj.yaml"
done
