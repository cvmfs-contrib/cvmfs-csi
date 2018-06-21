#!/bin/sh
kubectl create -f ./perms.yaml && kubectl create -f ./services.yaml && kubectl create -f ./statefulsets.yaml && kubectl create -f ./daemonsets.yaml