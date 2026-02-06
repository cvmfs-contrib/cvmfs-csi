module github.com/cvmfs-contrib/cvmfs-csi

go 1.25.0

require (
	github.com/container-storage-interface/spec v1.12.0
	github.com/kubernetes-csi/csi-lib-utils v0.23.2
	github.com/moby/sys/mountinfo v0.7.2
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
	k8s.io/apimachinery v0.35.0
	k8s.io/klog/v2 v2.130.1
	k8s.io/mount-utils v0.35.0
)

require (
	github.com/go-logr/logr v1.4.3 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	k8s.io/utils v0.0.0-20260108192941-914a6e750570 // indirect
)
