#!/bin/sh
kubectl delete -f ./cephfs-pvc.yaml
kubectl delete -f ./cephfs-sc.yaml