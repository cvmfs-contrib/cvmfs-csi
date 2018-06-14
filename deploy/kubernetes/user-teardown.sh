#!/bin/bash

objects=(pod pvc)

for obj in ${objects[@]}; do
	kubectl delete -f "./$obj.yaml"
done

sleep 1; kubectl delete -f ./cvmfs-storage-class.yaml
