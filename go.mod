module github.com/cvmfs-contrib/cvmfs-csi

go 1.22.0

toolchain go1.22.2

require (
	github.com/container-storage-interface/spec v1.9.0
	github.com/kubernetes-csi/csi-lib-utils v0.17.0
	github.com/moby/sys/mountinfo v0.7.1
	google.golang.org/grpc v1.63.2
	google.golang.org/protobuf v1.34.0
	k8s.io/apimachinery v0.30.0
	k8s.io/klog/v2 v2.120.1
	k8s.io/mount-utils v0.30.0
)

require (
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240429193739-8cf5692501f6 // indirect
	k8s.io/utils v0.0.0-20240423183400-0849a56e8f22 // indirect
)
