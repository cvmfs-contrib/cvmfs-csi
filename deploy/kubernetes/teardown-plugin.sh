#!/bin/sh

kubectl delete -f ./cvmfsplugin.yaml
kubectl delete -f ./cvmfs-storage-class.yaml
