#!/bin/sh

kubectl exec -it $(kubectl get pods -l app=csi-cvmfsplugin -o=name | head -n 1 | cut -f2 -d"/") -c csi-cvmfsplugin bash
