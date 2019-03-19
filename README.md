# OpenFaas Pipeline
Event driven pipeline using OpenFaaS, Minio and Tensorflow inception

## Deploy minio and configure webhook

### Docker swarm

##### Create minio secret and access key
```sh
SECRET_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)
ACCESS_KEY=$(head -c 12 /dev/urandom | shasum| cut -d' ' -f1)

echo -n "$SECRET_KEY" | docker secret create s3-secret-key -
echo -n "$ACCESS_KEY" | docker secret create s3-access-key -
```

##### Deploy minio to cluster
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

##### Publish port `9000`
```sh
docker service update minio --publish-add 9000:9000
``` 

##### Configure minio client
```sh
mc config host add minio http://127.0.0.1:9000 $ACCESS_KEY $SECRET_KEY
```
