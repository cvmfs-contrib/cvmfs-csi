#!/bin/sh
kubectl delete -f ./cvmfs-pvc.yaml
kubectl delete -f ./cvmfs-sc.yaml