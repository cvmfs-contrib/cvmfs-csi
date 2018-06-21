#!/bin/sh
kubectl delete -f cvmfs-demo.yaml
./vol-teardown-cvmfs.sh