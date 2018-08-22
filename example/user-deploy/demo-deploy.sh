#!/bin/sh
./vol-deploy.sh && kubectl create -f cvmfs-demo.yaml
