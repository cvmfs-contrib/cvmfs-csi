#!/bin/sh
./vol-deploy-cvmfs.sh && kubectl create -f cvmfs-demo.yaml