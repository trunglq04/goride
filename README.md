## Project overview

In this project‑driven course, a backend microservices system for a Uber‑style ride‑sharing app from the ground up—using Go, Docker, and Kubernetes.

## Trip Scheduling Flow

[![](https://mermaid.ink/img/pako:eNqNVt9v2jAQ_lcsP21qGvGjZSEPlSpaTX1YxWDVpAmpMvZBIkicOQ6UVf3fd4mdEgcKzQOK47v7vjt_d-aVcimAhnSW5vC3gJTDXcyWiiWzlOCTMaVjHmcs1eQpB3X49Xb88J1p2LLd4d4vFWdTUJuYw-HmnYo3oM5sH34fs10Cqf7Qb6oRFL-bnZLz5c3NnmRIRgrwteJGJmXOuTa2eyP0aFAPyXIyHtWO5YaxN7-PEoNJpNrM1nOSCw3Y_QuPWLq0nBvWl4h30fIos_Bhg5n6vMIVDTgVLyNN5II4LO9L65A8wrbyJimAyAkjwlbSBHBwENisQ2vl80T4pfezapbGRW1RHckkYaloILMNi9dsvgYXtHUQv2E-lXwF-hCccQ5ZE3sNiwZ0A_O2sjSw6oPTbB_nWbT3TB3dGEiykGrLlABBtCQTNp_H-sdP8sUwK3EmkGcS2-lnAQV8rUtwVCchGSvJIc8tJ2KoMGxDjxSZKIVapZZrpovc9_2j4nF7whGPifvM8jxepp8XkW2SzAQmOVKMZXoyF6_Nwq5bunetkLxp2HfIUQR8JQts5Bqz9DJGx3K1ZtZdHAVpCc9mZStkc3s-0WZtTFukOsGnB5QeE7u6PM4gKScQsozktmF_TmuN1qidUHZJa6q9V04m2RowlrV1SnbQc5GUq33YacFL_R2faHtPzxXt0ZN1O-6iXbRW1Zu4HxeiqjTJivk6ziPTcp-U1aXD-NPgZ46a21KL061Qj6kzc38_facarzCyv1tOdGgVk7OUpCipOSxjZEI9moBKWCzwKn8tQ8yojiCBGQ3xVTC1muEV_4Z2rNByuks5DbUqwKNKFsuIhgu2znFlZo79C1Cb4L36R8rmkoav9IWGvW_-1XVn0O_1-kE3GAyHgUd3-Lnb8fu9frc_xKfbvQ6CN4_-qyJ0_KDX7Q86QTDoDAfD66ve23_1IPGQ?type=png)](https://mermaid.live/edit#pako:eNqNVt9v2jAQ_lcsP21qGvGjZSEPlSpaTX1YxWDVpAmpMvZBIkicOQ6UVf3fd4mdEgcKzQOK47v7vjt_d-aVcimAhnSW5vC3gJTDXcyWiiWzlOCTMaVjHmcs1eQpB3X49Xb88J1p2LLd4d4vFWdTUJuYw-HmnYo3oM5sH34fs10Cqf7Qb6oRFL-bnZLz5c3NnmRIRgrwteJGJmXOuTa2eyP0aFAPyXIyHtWO5YaxN7-PEoNJpNrM1nOSCw3Y_QuPWLq0nBvWl4h30fIos_Bhg5n6vMIVDTgVLyNN5II4LO9L65A8wrbyJimAyAkjwlbSBHBwENisQ2vl80T4pfezapbGRW1RHckkYaloILMNi9dsvgYXtHUQv2E-lXwF-hCccQ5ZE3sNiwZ0A_O2sjSw6oPTbB_nWbT3TB3dGEiykGrLlABBtCQTNp_H-sdP8sUwK3EmkGcS2-lnAQV8rUtwVCchGSvJIc8tJ2KoMGxDjxSZKIVapZZrpovc9_2j4nF7whGPifvM8jxepp8XkW2SzAQmOVKMZXoyF6_Nwq5bunetkLxp2HfIUQR8JQts5Bqz9DJGx3K1ZtZdHAVpCc9mZStkc3s-0WZtTFukOsGnB5QeE7u6PM4gKScQsozktmF_TmuN1qidUHZJa6q9V04m2RowlrV1SnbQc5GUq33YacFL_R2faHtPzxXt0ZN1O-6iXbRW1Zu4HxeiqjTJivk6ziPTcp-U1aXD-NPgZ46a21KL061Qj6kzc38_facarzCyv1tOdGgVk7OUpCipOSxjZEI9moBKWCzwKn8tQ8yojiCBGQ3xVTC1muEV_4Z2rNByuks5DbUqwKNKFsuIhgu2znFlZo79C1Cb4L36R8rmkoav9IWGvW_-1XVn0O_1-kE3GAyHgUd3-Lnb8fu9frc_xKfbvQ6CN4_-qyJ0_KDX7Q86QTDoDAfD66ve23_1IPGQ)

## Installation

The project requires a couple tools to run, most of which are part of many developer's toolchains.

- Docker
- Go
- Tilt
- A local Kubernetes cluster

### Windows (WSL)

This is a step by step guide to install Go on Windows using WSL.
You can either install via WSL (recommended) or using powershell (not covered, but similar to WSL).

1. Install WSL for Windows from [Microsoft's official website](https://learn.microsoft.com/en-us/windows/wsl/install)

2. Install Docker for Windows from [Docker's official website](https://www.docker.com/products/docker-desktop/)

3. Install Minikube from [Minikube's official website](https://minikube.sigs.k8s.io/docs/)

4. Install Tilt from [Tilt's official website](https://tilt.dev/)

5. Install Go on Windows using WSL:

```bash
# 1. Get the Go binary
wget https://dl.google.com/go/go1.23.0.linux-amd64.tar.gz

# 2. Extract the tarball
sudo tar -xvf go1.23.0.linux-amd64.tar.gz

# 3. Move the extracted folder to /usr/local
sudo mv go /usr/local

# 4. Add Go to PATH (following the steps from the video)
cd ~
explorer.exe .

# Open .bashrc file and add following lines at the bottom and save the file.
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# 5. Verify the installation
go version
```

6. Make sure [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/) is installed.

## Run

```bash
tilt up
```

## Monitor

```bash
kubectl get pods
```

or

```bash
minikube dashboard
```

## Deployment (Azure example)

It's advisable to first run the steps manually and then build a proper CI/CD flow according to your infrastructure.

### 0. Environments

Export these in your `~/.bashrc` (or `.zshrc`) so every command below works as-is:

```bash
export REGION=malaysiawest              # pick the Azure region closest to you
export RESOURCE_GROUP=goride-rg         # logical container for all your Azure resources
export ACR_NAME=goridecr                # container registry name (globally unique, alphanumeric only)
export AKS_CLUSTER=goride-aks
export ACR_SERVER=$ACR_NAME.azurecr.io  # full registry hostname used in image tags
```

### 1. Add secrets.yaml file to the production folder

The production folder (`infra/production/azure/k8s/`) needs a `secrets.yaml` containing real values for:

- `STRIPE_SECRET_KEY` (test mode: `sk_test_...`)
- `STRIPE_WEBHOOK_KEY` (you get this in step 7 after registering the webhook)
- `MONGODB_URI`, `JWT_SECRET`, RabbitMQ credentials

You can copy the development secrets as a starting point, then fill in real values. Never commit them — `secrets.yaml` is already gitignored.

### 2. Create the Container Registry

Azure Container Registry (ACR) replaces Google Artifact Registry:

```bash
az group create --name $RESOURCE_GROUP --location $REGION

az acr create --resource-group $RESOURCE_GROUP --name $ACR_NAME --sku Basic

az acr login --name $ACR_NAME
```

`az acr login` configures your local Docker daemon to push/pull to `$ACR_SERVER`.

### 3. Build & push Docker images

Five images total — four Go services plus the Next.js frontend. All builds target `linux/amd64` because AKS nodes are x86-64 (without this flag you get `exec format error` pods on ARM laptops):

```bash
# API Gateway
docker build -t $ACR_SERVER/goride/api-gateway:latest --platform linux/amd64 \
  -f infra/production/azure/docker/api-gateway.Dockerfile .

# Trip Service
docker build -t $ACR_SERVER/goride/trip-service:latest --platform linux/amd64 \
  -f infra/production/azure/docker/trip-service.Dockerfile .

# Driver Service
docker build -t $ACR_SERVER/goride/driver-service:latest --platform linux/amd64 \
  -f infra/production/azure/docker/driver-service.Dockerfile .

# Payment Service
docker build -t $ACR_SERVER/goride/payment-service:latest --platform linux/amd64 \
  -f infra/production/azure/docker/payment-service.Dockerfile .

docker push $ACR_SERVER/goride/api-gateway:latest
docker push $ACR_SERVER/goride/trip-service:latest
docker push $ACR_SERVER/goride/driver-service:latest
docker push $ACR_SERVER/goride/payment-service:latest
```

#### Web frontend (special case)

The frontend's `NEXT_PUBLIC_*` variables are **inlined into the JavaScript bundle at build time**, not read at runtime. They are passed as Docker build args from a gitignored file so real IPs never land in git:

```bash
# One time: create the build-env file (values depend on your deployed gateway IP, see step 7)
cat > infra/production/azure/docker/web.env.build <<EOF
NEXT_PUBLIC_API_URL=http://<API_GATEWAY_EXTERNAL_IP>:8081
NEXT_PUBLIC_WEBSOCKET_URL=ws://<API_GATEWAY_EXTERNAL_IP>:8081/ws
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_...
EOF

docker build -t $ACR_SERVER/goride/web:latest --platform linux/amd64 \
  --build-arg NEXT_PUBLIC_API_URL=$(grep '^NEXT_PUBLIC_API_URL=' infra/production/azure/docker/web.env.build | cut -d= -f2-) \
  --build-arg NEXT_PUBLIC_WEBSOCKET_URL=$(grep '^NEXT_PUBLIC_WEBSOCKET_URL=' infra/production/azure/docker/web.env.build | cut -d= -f2-) \
  --build-arg NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=$(grep '^NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=' infra/production/azure/docker/web.env.build | cut -d= -f2-) \
  -f infra/production/azure/docker/web.Dockerfile .

docker push $ACR_SERVER/goride/web:latest
```

> Tip: `./infra/production/azure/build-push-images.sh` builds and pushes everything in one shot (it reads `web.env.build` automatically for the web image).

### 4. Create an AKS cluster

```bash
az provider register --namespace Microsoft.ContainerService

az aks create \
  --resource-group $RESOURCE_GROUP \
  --name $AKS_CLUSTER \
  --node-count 2 \
  --enable-managed-identity \
  --attach-acr $ACR_NAME \
  --generate-ssh-keys
```

Key flag: `--attach-acr` grants the cluster's managed identity `AcrPull` on your registry — pods pull images with zero extra configuration (no `imagePullSecrets` needed).

To tear the cluster down later:

```bash
az aks delete --resource-group $RESOURCE_GROUP --name $AKS_CLUSTER --yes
```

### 5. Connect kubectl to AKS

Replaces `gcloud container clusters get-credentials`:

```bash
az aks get-credentials --resource-group $RESOURCE_GROUP --name $AKS_CLUSTER --overwrite-existing
kubectl get nodes   # verify connectivity
```

To go back to local development afterwards:

```bash
kubectl config get-contexts
kubectl config use-context minikube   # or docker-desktop
```

### 6. Apply Kubernetes manifests

Apply in dependency order — shared config first, then infrastructure, then apps:

```bash
# Config & secrets first
kubectl apply -f infra/production/azure/k8s/app-config.yaml
kubectl apply -f infra/production/azure/k8s/secrets.yaml

# Infrastructure services
kubectl apply -f infra/production/azure/k8s/jaeger-deployment.yaml
kubectl apply -f infra/production/azure/k8s/rabbitmq-deployment.yaml

# Wait until Jaeger and RabbitMQ are Running (backend services fatal-exit without RabbitMQ)
kubectl get pods -w

# Application services
kubectl apply -f infra/production/azure/k8s/api-gateway-deployment.yaml
kubectl apply -f infra/production/azure/k8s/driver-service-deployment.yaml
kubectl apply -f infra/production/azure/k8s/trip-service-deployment.yaml
kubectl apply -f infra/production/azure/k8s/payment-service-deployment.yaml
kubectl apply -f infra/production/azure/k8s/web-deployment.yaml
```

Redeploying:

```bash
kubectl rollout restart deployment/<name>   # restart one service
kubectl rollout restart deployment          # restart everything
kubectl delete pod <pod-name>               # nudge a single stuck pod
```

### 7. Get your URLs & wire up Stripe

```bash
kubectl get svc
```

Read the `EXTERNAL-IP` column of the `LoadBalancer` services (ignore `ClusterIP` rows — those are internal-only):

- `web` EXTERNAL-IP → your frontend URL (`http://<WEB_IP>`)
- `api-gateway` EXTERNAL-IP → your backend URL (`http://<GW_IP>:8081`)

Register the Stripe webhook: Dashboard → Developers → Webhooks → Add endpoint:

```
http://<GW_IP>:8081/webhook/stripe
```

Select the `checkout.session.completed` event, copy the signing secret (`whsec_...`) into `secrets.yaml` as `STRIPE_WEBHOOK_KEY`, then:

```bash
kubectl apply -f infra/production/azure/k8s/secrets.yaml
kubectl rollout restart deployment/api-gateway
```

Finally point checkout redirects at the frontend in `app-config.yaml`:

```yaml
STRIPE_SUCCESS_URL: "http://<WEB_IP>?payment=success"
STRIPE_CANCEL_URL: "http://<WEB_IP>?payment=cancel"
```

```bash
kubectl apply -f infra/production/azure/k8s/app-config.yaml
kubectl rollout restart deployment/payment-service
```

> Remember: if either EXTERNAL-IP ever changes, you must rebuild the **web** image (build-time baked URLs) and re-register the Stripe webhook URL.

## Adding HTTPS to your API

> Already applied on `goride-aks` (2026-08-26) — see
> [docs/architecture/ingress-setup-guide.md](docs/architecture/ingress-setup-guide.md)
> for what was actually done and why, including how the 2-public-IP quota issue was resolved.

Everything above serves plain HTTP on bare IPs. That's fine for testing, but it breaks real things:

- Browsers disable powerful APIs on insecure origins — e.g. `crypto.randomUUID()` only exists in HTTPS/localhost contexts, which crashed this very app when served over `http://<IP>`
- Stripe **requires** HTTPS endpoints for live-mode webhooks
- No valid TLS = no trust: users see warnings, and cookies marked `Secure` don't work

### How TLS on Kubernetes works (the mental model)

```
Browser ──HTTPS──▶ Cloudflare edge ──HTTPS (Full strict)──▶ Ingress ──HTTP──▶ api-gateway Service ──▶ Pod
 (TLS hop 1)         (CDN/WAF, hides origin)                (TLS hop 2 terminates here,
                                                             reads cert from k8s Secret,
                                                             auto-renewed by cert-manager)
```

Four moving parts:

1. **A domain name** — Certificate Authorities (like Let's Encrypt) only issue certs for domains, not bare IPs. Buy one (~$10/year) and create an `A` record pointing it at your cluster.
2. **Ingress controller** (nginx) — the single public entrypoint that owns port 443 and routes by hostname/path.
3. **cert-manager** — automation daemon that proves domain ownership to Let's Encrypt (HTTP-01 challenge) and keeps the certificate renewed forever.
4. **Cloudflare (proxy)** — both hostnames are proxied through Cloudflare: public DNS returns CF anycast IPs, the real origin IP stays hidden, and TLS hop #2 must be set to **Full (strict)** so Cloudflare validates our Let's Encrypt cert.

### 0. Reserve a static public IP

Load balancer IPs change whenever the Service is recreated — which silently breaks your DNS. Reserve one first:

```bash
NODE_RG=$(az aks show -g $RESOURCE_GROUP -n $AKS_CLUSTER --query nodeResourceGroup -o tsv)

az network public-ip create \
  --resource-group $NODE_RG \
  --name goride-api-ip \
  --sku Standard \
  --allocation-method static
```

Then pin it to the gateway service in `api-gateway-deployment.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: api-gateway
spec:
  type: LoadBalancer
  # deprecated in upstream k8s, but still the standard mechanism on AKS
  loadBalancerIP: <YOUR_STATIC_IP>
  ports:
    - port: 8081
      targetPort: 8081
  selector:
    app: api-gateway
```

Point your domain's `A` record at this IP now (DNS propagation takes minutes to hours).

### 1. Install the nginx ingress controller

Requires [Helm](https://helm.sh/docs/intro/install/):

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install nginx-ingress ingress-nginx/ingress-nginx \
  --namespace ingress-nginx --create-namespace \
  --set controller.service.loadBalancerIP=<YOUR_STATIC_IP> \
  --set controller.service.annotations."service\.beta\.kubernetes\.io/azure-load-balancer-health-probe-request-path"=/healthz
```

This creates *another* LoadBalancer — the one above gets the static IP instead. Verify:

```bash
kubectl get svc -n ingress-nginx
```

If you give the static IP to nginx-ingress (recommended — it becomes your permanent entrypoint), remove `loadBalancerIP` from api-gateway and let it become ClusterIP in step 4.

### 2. Install cert-manager

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml

kubectl get pods -n cert-manager   # wait until all three pods are Running
```

### 3. Create a Let's Encrypt issuer

Save as `infra/production/azure/k8s/letsencrypt-issuer.yaml`:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: you@example.com
    privateKeySecretRef:
      name: letsencrypt-key
    solvers:
      - http01:
          ingress:
            class: nginx
```

```bash
kubectl apply -f infra/production/azure/k8s/letsencrypt-issuer.yaml
```

How the challenge works: Let's Encrypt asks "prove you own `api.yourdomain.com`" by requesting `http://api.yourdomain.com/.well-known/acme-challenge/<token>`. cert-manager temporarily publishes that token through the nginx ingress. If it resolves, the cert issues and auto-renews every ~60 days.

### 4. Create the Ingress & switch the service to ClusterIP

Save as `infra/production/azure/k8s/api-gateway-ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-gateway-ingress
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.yourdomain.com
      secretName: api-gateway-tls     # cert-manager fills this Secret automatically
  rules:
    - host: api.yourdomain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: api-gateway
                port:
                  number: 8081
```

Then flip the api-gateway Service from `type: LoadBalancer` to `type: ClusterIP` (external traffic now enters through nginx, not directly):

```bash
kubectl apply -f infra/production/azure/k8s/api-gateway-ingress.yaml
kubectl apply -f infra/production/azure/k8s/api-gateway-deployment.yaml
```

### 5. Verify

```bash
kubectl get certificate -n default
# READY=True means issued; False = check events:
kubectl describe certificate api-gateway-tls
```

Your API is now `https://api.yourdomain.com`. Final wiring updates:

1. Stripe Dashboard: update the webhook URL to `https://api.yourdomain.com/webhook/stripe` (same signing secret)
2. Rebuild the web image with `NEXT_PUBLIC_API_URL=https://api.yourdomain.com` and redeploy
3. Update `STRIPE_SUCCESS_URL`/`CANCEL_URL` to `https://<your-web-host>`
