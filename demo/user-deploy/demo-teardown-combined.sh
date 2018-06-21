#!/bin/sh
kubectl delete -f combined-demo.yaml
./vol-teardown-cephfs.sh
./vol-teardown-cvmfs.sh