#!/bin/sh
kubectl delete -f ./daemonsets.yaml && kubectl delete -f ./statefulsets.yaml && kubectl delete -f ./services.yaml && kubectl delete -f ./perms.yaml