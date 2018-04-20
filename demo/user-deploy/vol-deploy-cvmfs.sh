#!/bin/sh
kubectl create -f ./cvmfs-sc.yaml && kubectl create -f ./cvmfs-pvc.yaml