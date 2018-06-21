#!/bin/sh
kubectl delete -f cephfs-demo.yaml
./vol-teardown-cephfs.sh