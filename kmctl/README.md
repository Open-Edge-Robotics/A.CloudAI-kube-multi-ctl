# Build and Deploy Guide

빌드 및 배포를 위한 가이드

## Build

Makefile 을 이용하여 Server, Client 를 빌드하고 Docker Image 를 생성하고 Push

```bash
# Build Server and Client
make build

# Build files
cd ../build && ls
```

## Client Installation

Client 를 빌드하고 설치

```bash
# Client Install
make install

# Test
kmctl --version

# Uninstall
make uninstall
```

## Run Local Server

빌드한 서버를 로컬에서 실행

```bash
# Run Server Locally
make serve
```

## Docker Build

Docker Image 를 빌드하고 Push

### Build Docker Image

```bash
# Build Docker Image
make docker-build
```

### Push Docker Image

```bash
# Build and Push Docker Image
make docker-push
```

## Clean

```bash
# Clean
make clean
```

## Configuration

### config.yaml

기본 설정된 config.yaml 을 수정하여 클라이언트 설정

```bash
# Edit config.yaml
vi config.yaml
```

```yaml
# Server Configuration
server:
  host: "localhost"
  port: 8080
  timeout: 10
```

### Server Deployment Yaml

서버를 배포하기 위한 server-deployment.yaml 을 수정하여 서버 설정

```bash
# Edit server-deployment.yaml
vi server-deployment.yaml
```

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-backend
  namespace: lge-ebme
  labels:
    app: kube-backend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kube-backend
  template:
    metadata:
      labels:
        app: kube-backend
        name: kube-backend
    spec:
      containers:
        - name: kube-backend
          image: lgecloudroboticstask/kube_backend:2024-11-18
          imagePullPolicy: IfNotPresent
          command:
            - /bin/sh
            - -c
          args:
            - ./app serve --host 0.0.0.0
          ports:
            - containerPort: 50051
          resources:
            limits:
              cpu: 1000m
              memory: 500Mi
      imagePullSecrets:
        - name: regcred

---
apiVersion: v1
kind: Service
metadata:
  name: kube-backend
  namespace: lge-ebme
spec:
  selector:
    app: kube-backend
  ports:
    - port: 8080
      targetPort: 50051
      protocol: TCP
      nodePort: 30300
  type: NodePort

---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kube-backend
rules:
  - apiGroups: ["apps", ""]
    resources:
      ["pods", "nodes", "services", "deployments", "configmaps", "pods/log"]
    verbs:
      ["get", "list", "watch", "create", "update", "patch", "delete"]

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kube-backend
subjects:
  - kind: ServiceAccount
    name: default
    namespace: lge-ebme
roleRef:
  kind: ClusterRole
  name: kube-backend
  apiGroup: rbac.authorization.k8s.io
```
