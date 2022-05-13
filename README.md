# image-clone-controller

image-clone-controller is a custom controller for k8s obejcts Deployment and Daemonsets. 
It watches for the new application objects and mirrors the images to backup registry
repository and reconfigures the applications to use these copies. If the applications 
are using the backup registry already, then it will be ignored. Here we are using local
registry as the backup registry. 

**Setup Requirements**

```
golang
Minikube
Operator-SDK
Docker desktop
```

**How to use**

1. Create minikube cluster with insecure registry

```
minikube start --insecure-registry "10.0.0.0/24"
minikube addons enable registry
```
PS: Not down the port given by minikube, this is needed to access local registry from host and also needed in controllers. 
In this example it is 65132

2. Clone the git repo
 ```
git clone https://github.com/sachintiptur/image-clone-controller.git
```
3. Start the controller outside cluster

```
make run
```
Output:
```
go/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/go/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go run ./main.go
1.652435502208323e+09	INFO	controller-runtime.metrics	Metrics server is starting to listen	{"addr": ":8080"}
1.652435502208767e+09	INFO	setup	starting manager
1.6524355022091808e+09	INFO	Starting server	{"path": "/metrics", "kind": "metrics", "addr": "[::]:8080"}
1.65243550220919e+09	INFO	Starting server	{"kind": "health probe", "addr": "[::]:8081"}
1.652435502209484e+09	INFO	controller.deployment	Starting EventSource	{"reconciler group": "apps", "reconciler kind": "Deployment", "source": "kind source: *v1.Deployment"}
1.6524355022095082e+09	INFO	controller.deployment	Starting Controller	{"reconciler group": "apps", "reconciler kind": "Deployment"}
1.652435502209506e+09	INFO	controller.daemonset	Starting EventSource	{"reconciler group": "apps", "reconciler kind": "DaemonSet", "source": "kind source: *v1.DaemonSet"}
1.6524355022095191e+09	INFO	controller.daemonset	Starting Controller	{"reconciler group": "apps", "reconciler kind": "DaemonSet"}
1.652435502309921e+09	INFO	controller.daemonset	Starting workers	{"reconciler group": "apps", "reconciler kind": "DaemonSet", "worker count": 1}
1.652435502309956e+09	INFO	controller.deployment	Starting workers	{"reconciler group": "apps", "reconciler kind": "Deployment", "worker count": 1}
1.652435502310148e+09	INFO	controller.deployment	Deployment	{"reconciler group": "apps", "reconciler kind": "Deployment", "name": "nginx-1", "namespace": "default", "namespace": "default"}
1.652435502310272e+09	DEBUG	controller.deployment	Image	{"reconciler group": "apps", "reconciler kind": "Deployment", "name": "nginx-1", "namespace": "default", "registry": "localhost:5000/nginx"}
1.6524355023102798e+09	DEBUG	controller.deployment	Deployment already using local registry	{"reconciler group": "apps", "reconciler kind": "Deployment", "name": "nginx-1", "namespace": "default"}
1.65243550231024e+09	INFO	controller.daemonset	Daemonset	{"reconciler group": "apps", "reconciler kind": "DaemonSet", "name": "kube-proxy", "namespace": "kube-system", "namespace": "kube-system"}
1.652435502310305e+09	INFO	controller.daemonset	Ignore kube-system daemonsets	{"reconciler group": "apps", "reconciler kind": "DaemonSet", "name": "kube-proxy", "namespace": "kube-system"}
1.652435502310339e+09	INFO	controller.deployment	Deployment	{"reconciler group": "apps", "reconciler kind": "Deployment", "name": "coredns", "namespace": "kube-system", "namespace": "kube-system"}
1.652435502310346e+09	DEBUG	controller.deployment	Ignore deployments belonging to kube-system namespace	{"reconciler group": "apps", "reconciler kind": "Deployment", "name": "coredns", "namespace": "kube-system"}
1.652435502310368e+09	INFO	controller.daemonset	Daemonset	{"reconciler group": "apps", "reconciler kind": "DaemonSet", "name": "registry-proxy", "namespace": "kube-system", "namespace": "kube-system"}
1.652435502310377e+09	INFO	controller.daemonset	Ignore kube-system daemonsets	{"reconciler group": "apps", "reconciler kind": "DaemonSet", "name": "registry-proxy", "namespace": "kube-system"}
1.652435502310583e+09	INFO	controller.daemonset	Daemonset	{"reconciler group": "apps", "reconciler kind": "DaemonSet", "name": "fluentd", "namespace": "default", "namespace": "default"}
1.6524355023106182e+09	INFO	controller.daemonset	Daemonset is already using local registry	{"reconciler group": "apps", "reconciler kind": "DaemonSet", "name": "fluentd", "namespace": "default"}
```

4. Run the below docker container to redirect the minikube port for local registry

```
docker run --rm -it --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):65132"
```

**How to test**

1. Check cluster status
2. Check local registry contents for empty
3. Create a deployment
4. Check local registry and the image is backed up
5. Check deployment yaml file for updated image registry
6. Create a Daemonset
7. Check local registry and the image is backed up
8. Check daemonset yaml file for updated image registry
9. For sample output, see [Output.md](Output.md)

```
minikube status
curl localhost:65132/v2/_catalog

minikube kubectl -- create deployment nginx-1 --image=nginx
curl localhost:65132/v2/_catalog
minikube kubectl -- get deployment nginx-1 -o yaml
minikube kubectl -- get deployment nginx-1 -n default
curl localhost:65132/v2/_ca
talog
minikube kubectl -- create -f daemonset-example.yaml
curl localhost:65132/v2/_catalog

minikube kubectl -- get ds  -o yaml
minikube kubectl -- get ds  -A
curl localhost:65132/v2/_catalog
```

**References**

1. https://minikube.sigs.k8s.io/docs/handbook/registry/
2. https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/
3. https://github.com/google/go-containerregistry
4. https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
