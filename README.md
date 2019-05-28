# OpenFaas Pipeline
Event driven pipeline using OpenFaaS, Minio and Tensorflow inception


![OpenFaaS Pipeline](https://github.com/viveksyngh/openfaas-pipeline/blob/master/media/openfaas-pipeline.jpg?raw=true)


## Kubernetes 
For Kubernetes, first install Helm and Tiller

### Install Helm 

#### Install Helm CLI Client

* On Linux and Mac/Darwin
```
curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash
```

* On Mac Via Homebrew 
```
brew install kubernetes-helm 
```

#### Install Tiller

* Create RBAC permissions for tiller
```
kubectl -n kube-system create sa tiller \
  && kubectl create clusterrolebinding tiller \
  --clusterrole cluster-admin \
  --serviceaccount=kube-system:tiller
```

* Install the server-side Tiller component on your cluster
```
helm init --skip-refresh --upgrade --service-account tiller
```

### Install OpenFaaS on Kubernetes

* Create OpenFaaS namespace 
```sh
kubectl apply -f https://raw.githubusercontent.com/openfaas/faas-netes/master/namespaces.yml
```

* Add OpenFaaS helm repository 
```sh
helm repo add openfaas https://openfaas.github.io/faas-netes/
```

* Create basic-auth credentials
```sh
PASSWORD=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
```

```sh
kubectl -n openfaas create secret generic basic-auth \
--from-literal=basic-auth-user=admin \
--from-literal=basic-auth-password="$PASSWORD"
```

* Install OpenFaaS on kubernetes cluster
```sh
helm repo update \
 && helm upgrade openfaas --install openfaas/openfaas \
    --namespace openfaas  \
    --set basic_auth=true \
    --set functionNamespace=openfaas-fn \
    --set operator.create=true

```

### Install and Configure minio

* Create OpenFaaS namespaces
```
kubectl apply -f https://raw.githubusercontent.com/openfaas/faas-netes/master/namespaces.yml
```

* Generate secrets for Minio
```
SECRET_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
ACCESS_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
```

* Create Secrets in kubernetes
```
kubectl create secret generic -n openfaas-fn \
 s3-secret-key --from-literal s3-secret-key="$SECRET_KEY"
```

```
kubectl create secret generic -n openfaas-fn \
 s3-access-key --from-literal s3-access-key="$ACCESS_KEY"
```

* Install minio with helm chart
```
helm install --name cloud-minio \
    --namespace openfaas \
    --set accessKey=$ACCESS_KEY,
            secretKey=$SECRET_KEY,
            replicas=1,
            persistence.enabled=false,
            service.port=9000,
            service.type=NodePort \
    stable/minio
```

* Get Minio NodePort

```
MINIO_PORT=$(kubectl get svc/cloud-minio -n openfaas --output=jsonpath='{.spec.ports[?(@.name=="service")].nodePort}')
``` 

* Configure minio client
```sh
mc config host add minio-kube http://127.0.0.1:$MINIO_PORT $ACCESS_KEY $SECRET_KEY
```

* Get minio config and edit the JSON to add webhook handler
```sh
mc admin config get minio-kube > myconfig.json
```
edit webhook section of `myconfig.json` and save it
```json
"webhook":{
    "1":{
        "enable":true,
        "endpoint":"http://<gateway-ip>:8080/function/minio-webhook-hanlder"
        }
    }
}
```

* Update minio config and restart mino server
Update the mini config and restart minio server
```sh
mc admin config set minio-kube < myconfig.json
```

```sh
mc admin service restart minio-kube
```

* Create buckets
```sh
mc mb minio-kube/images
```
```sh
mc mb minio-kube/images-thumbnail
```
```sh
mc mb minio-kube/inception
```

* Add event for the webhook
```sh
mc event add minio/images arn:minio:sqs::1:webhook --event put --suffix .jpg
```

* Change `images` bucket policy to public so that inception function can download the image without secret
```sh
mc policy public minio-kube/images
```

#### Deploy Functions

```sh
faas-cli deploy -f stack.kubernetes.yml
```

### Docker swarm

### Install OpenFaaS
```sh
git clone https://www.github.com/openfaas/faas && \
        cd faas && ./deploy_stack.sh
```

### Install and Configure minio

* Create minio secret and access key
```sh
SECRET_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
ACCESS_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)

echo -n "$SECRET_KEY" | docker secret create s3-secret-key -
echo -n "$ACCESS_KEY" | docker secret create s3-access-key -
```

* Deploy minio to cluster
```sh
docker service create --constraint="node.role==manager" \
 --name minio \
 --mount type=bind,source=~/minio/data,target=/data \
 --mount type=bind,source=~/minio/config,target=/root/.minio \
 --detach=true --network func_functions \
 --secret s3-access-key \
 --secret s3-secret-key \
 --env MINIO_SECRET_KEY_FILE=s3-secret-key \
 --env MINIO_ACCESS_KEY_FILE=s3-access-key \
minio/minio:latest server /data
```

* Publish port `9000`
```sh
docker service update minio --publish-add 9000:9000
``` 

* Configure minio client
```sh
mc config host add minio http://127.0.0.1:9000 $ACCESS_KEY $SECRET_KEY
```

* Get minio config and edit the JSON to add webhook handler
```sh
mc admin config get minio > myconfig.json
```
edit webhook section of `myconfig.json` and save it
```json
"webhook":{
    "1":{
        "enable":true,
        "endpoint":"http://<gateway-ip>:8080/function/minio-webhook-hanlder"
        }
    }
}
```

* Update minio config and restart mino server
Update the mini config and restart minio server
```sh
mc admin config set minio < myconfig.json
```

```sh
mc admin service restart minio
```

* Create required buckets
```sh
mc mb minio/images
```
```sh
mc mb minio/images-thumbnail
```
```sh
mc mb minio/inception
```

* Add event for the webhook
```sh
mc event add minio/images arn:minio:sqs::1:webhook --event put --suffix .jpg
```

* Change `images` bucket policy to public so that inception function can download the image without secret
```sh
mc policy public minio/images
```

#### Deploy Functions

```sh
faas-cli deploy -f stack.swarm.yml
```