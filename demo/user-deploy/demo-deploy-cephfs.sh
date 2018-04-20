#!/bin/sh
./vol-deploy-cephfs.sh && kubectl create -f cephfs-demo.yaml