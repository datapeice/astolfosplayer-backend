# Kubernetes Deployment Guide for Astolfo's Player Backend

This guide explains how to deploy the microservices (Auth, File, Sync) and infrastructure (MinIO) to a Kubernetes cluster.

## Prerequisites

- A running Kubernetes cluster (e.g., Minikube, K3s, GKE, EKS).
- `kubectl` configured to communicate with your cluster.
- `docker` installed for building images (if not using a registry).

## 1. Create Namespace

Create a dedicated namespace for the application:

```bash
kubectl create namespace astolfos-player
```

## 2. Deploy MinIO (S3 Storage)

Create `k8s/minio-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
  namespace: astolfos-player
spec:
  replicas: 1
  selector:
    matchLabels:
      app: minio
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: minio/minio:latest
        args:
        - server
        - /data
        - --console-address
        - :9001
        env:
        - name: MINIO_ROOT_USER
          value: "minioadmin"
        - name: MINIO_ROOT_PASSWORD
          value: "minioadmin"
        ports:
        - containerPort: 9000
        - containerPort: 9001
        volumeMounts:
        - name: minio-storage
          mountPath: /data
      volumes:
      - name: minio-storage
        emptyDir: {} # Use PersistentVolumeClaim in production
---
apiVersion: v1
kind: Service
metadata:
  name: minio
  namespace: astolfos-player
spec:
  selector:
    app: minio
  ports:
    - protocol: TCP
      port: 9000
      targetPort: 9000
      name: api
    - protocol: TCP
      port: 9001
      targetPort: 9001
      name: console
```

Apply it:
```bash
kubectl apply -f k8s/minio-deployment.yaml
```

## 3. Deploy Auth Service

Create `k8s/auth-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: astolfos-player
spec:
  replicas: 1
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: astolfos-auth:latest # Ensure this image is available in your cluster
        imagePullPolicy: IfNotPresent
        env:
        - name: DATABASE_URL
          value: "/data/auth.db"
        - name: SECRET_KEY
          value: "prod-secret-key"
        - name: SECURITY_KEY
          value: "prod-security-key"
        ports:
        - containerPort: 50051
        volumeMounts:
        - name: auth-data
          mountPath: /data
      volumes:
      - name: auth-data
        emptyDir: {} # Use PersistentVolumeClaim in production
---
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: astolfos-player
spec:
  selector:
    app: auth-service
  ports:
    - protocol: TCP
      port: 50051
      targetPort: 50051
```

Apply it:
```bash
kubectl apply -f k8s/auth-deployment.yaml
```

## 4. Deploy File Service

Create `k8s/file-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: file-service
  namespace: astolfos-player
spec:
  replicas: 1
  selector:
    matchLabels:
      app: file-service
  template:
    metadata:
      labels:
        app: file-service
    spec:
      containers:
      - name: file-service
        image: astolfos-file:latest
        imagePullPolicy: IfNotPresent
        env:
        - name: S3_ENDPOINT
          value: "minio:9000"
        - name: S3_ACCESS_KEY
          value: "minioadmin"
        - name: S3_SECRET_KEY
          value: "minioadmin"
        - name: S3_BUCKET
          value: "music"
        - name: S3_USE_SSL
          value: "false"
        - name: DATABASE_URL
          value: "/data/metadata.db"
        ports:
        - containerPort: 50052
        volumeMounts:
        - name: metadata-storage
          mountPath: /data
      volumes:
      - name: metadata-storage
        emptyDir: {} # Use PersistentVolumeClaim in production
---
apiVersion: v1
kind: Service
metadata:
  name: file-service
  namespace: astolfos-player
spec:
  selector:
    app: file-service
  ports:
    - protocol: TCP
      port: 50052
      targetPort: 50052
```

Apply it:
```bash
kubectl apply -f k8s/file-deployment.yaml
```

## 5. Deploy Sync Service

Create `k8s/sync-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sync-service
  namespace: astolfos-player
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sync-service
  template:
    metadata:
      labels:
        app: sync-service
    spec:
      containers:
      - name: sync-service
        image: astolfos-sync:latest
        imagePullPolicy: IfNotPresent
        env:
        - name: DATABASE_URL
          value: "/data/metadata.db"
        ports:
        - containerPort: 50053
        volumeMounts:
        - name: metadata-storage
          mountPath: /data
      volumes:
      - name: metadata-storage
        emptyDir: {} # IMPORTANT: In production, File and Sync services MUST share the same PVC for SQLite
```

Apply it:
```bash
kubectl apply -f k8s/sync-deployment.yaml
```

## Important Notes for Production

1.  **Persistent Storage**: The examples above use `emptyDir` which loses data on pod restart. You MUST replace these with `PersistentVolumeClaim` (PVC).
2.  **Shared Volume**: The `file-service` and `sync-service` both need access to the same SQLite database (`metadata.db`). In Kubernetes, this requires a `ReadWriteMany` PVC (like NFS) or running them in the same Pod as sidecars. Alternatively, switch to PostgreSQL for easier shared access.
3.  **Secrets**: Do not hardcode passwords/keys in YAML. Use Kubernetes Secrets.
4.  **Ingress**: To expose services to the outside world, configure an Ingress controller (e.g., Nginx).
