module github.com/cvmfs-contrib/cvmfs-csi

go 1.24.0

toolchain go1.24.3

require (
	github.com/container-storage-interface/spec v1.11.0
	github.com/kubernetes-csi/csi-lib-utils v0.21.0
	github.com/moby/sys/mountinfo v0.7.2
	google.golang.org/grpc v1.72.1
	google.golang.org/protobuf v1.36.6
	k8s.io/apimachinery v0.33.1
	k8s.io/klog/v2 v2.130.1
	k8s.io/mount-utils v0.33.1
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250512202823-5a2f75b736a9 // indirect
	k8s.io/utils v0.0.0-20250502105355-0f33e8f1c979 // indirect
)
