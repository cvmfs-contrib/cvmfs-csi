#!/bin/sh

kubectl create -f ./cvmfs-storage-class.yaml
kubectl create -f ./cvmfsplugin.yaml
