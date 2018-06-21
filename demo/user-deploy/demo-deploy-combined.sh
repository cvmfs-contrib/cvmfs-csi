#!/bin/sh
./vol-deploy-cephfs.sh && ./vol-deploy-cvmfs.sh && kubectl create -f combined-demo.yaml