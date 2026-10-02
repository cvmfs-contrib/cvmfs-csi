module github.com/cvmfs-contrib/cvmfs-csi

go 1.26.0

require (
	github.com/container-storage-interface/spec v1.13.0
	github.com/kubernetes-csi/csi-lib-utils v0.25.0
	github.com/moby/sys/mountinfo v0.7.2
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
	k8s.io/apimachinery v0.37.1
	k8s.io/klog/v2 v2.140.0
	k8s.io/mount-utils v0.37.1
)

require (
	github.com/go-logr/logr v1.4.4 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260928230214-8a89bd6388cc // indirect
	k8s.io/utils v0.0.0-20260707023825-cf1189d6abe3 // indirect
)
