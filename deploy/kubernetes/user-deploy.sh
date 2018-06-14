#!/bin/bash

objects=(cvmfs-storage-class pvc pod)

for obj in ${objects[@]}; do
	kubectl create -f "./$obj.yaml"
done
