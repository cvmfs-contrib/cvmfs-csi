#!/bin/sh
kubectl create -f ./cephfs-sc.yaml && kubectl create -f ./cephfs-pvc.yaml