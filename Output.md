Minikube status

```
$image-clone % minikube status

minikube
type: Control Plane
host: Running
kubelet: Running
apiserver: Running
kubeconfig: Configured
```
Empty backup registry 
```
image-clone % curl localhost:65132/v2/_catalog
{"repositories":[]}
```
Create Deployment and backup registry

```
image-clone % minikube kubectl -- create deployment nginx-2 --image=nginx

deployment.apps/nginx-2 created
image-clone % curl localhost:65132/v2/_catalog                           
{"repositories":["nginx"]}
```
Check Deployment yaml

```

image-clone % minikube kubectl -- get deployment nginx-2 -o yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    deployment.kubernetes.io/revision: "2"
  creationTimestamp: "2022-05-13T10:16:08Z"
  generation: 2
  labels:
    app: nginx-2
  name: nginx-2
  namespace: default
  resourceVersion: "41003"
  uid: e2845d4e-4c30-4f20-b398-642702057668
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: nginx-2
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: nginx-2
    spec:
      containers:
      - image: localhost:5000/nginx
        imagePullPolicy: Always
        name: nginx
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
status:
  availableReplicas: 1
  conditions:
  - lastTransitionTime: "2022-05-13T10:16:11Z"
    lastUpdateTime: "2022-05-13T10:16:11Z"
    message: Deployment has minimum availability.
    reason: MinimumReplicasAvailable
    status: "True"
    type: Available
  - lastTransitionTime: "2022-05-13T10:16:08Z"
    lastUpdateTime: "2022-05-13T10:16:12Z"
    message: ReplicaSet "nginx-2-845596cdbd" has successfully progressed.
    reason: NewReplicaSetAvailable
    status: "True"
    type: Progressing
  observedGeneration: 2
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1
  ```
  
  Create Deamonset
  ```
  image-clone % minikube kubectl -- create -f daemonset-example.yaml
daemonset.apps/fluentd created
 image-clone % curl localhost:65132/v2/_catalog                           
{"repositories":["fluentd","nginx"]}
```

Check Daemonset yaml

```
   image-clone % minikube kubectl -- get ds fluentd -o yaml

apiVersion: apps/v1
kind: DaemonSet
metadata:
  annotations:
    deprecated.daemonset.template.generation: "2"
  creationTimestamp: "2022-05-13T10:25:08Z"
  generation: 2
  labels:
    k8s-app: fluentd
  name: fluentd
  namespace: default
  resourceVersion: "41432"
  uid: 1594291d-500b-47f8-85a3-f9b82d40044c
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      name: fluentd
  template:
    metadata:
      creationTimestamp: null
      labels:
        name: fluentd
    spec:
      containers:
      - image: localhost:5000/fluentd:latest
        imagePullPolicy: Always
        name: fluentd
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
  updateStrategy:
    rollingUpdate:
      maxSurge: 0
      maxUnavailable: 1
    type: RollingUpdate
status:
  currentNumberScheduled: 1
  desiredNumberScheduled: 1
  numberAvailable: 1
  numberMisscheduled: 0
  numberReady: 1
  observedGeneration: 2
  updatedNumberScheduled: 1
  
  progressDeadlineSeconds: 600
