# OpenFaas Pipeline
Event driven pipeline using OpenFaaS, Minio and Tensorflow inception


![OpenFaaS Pipeline](https://github.com/viveksyngh/openfaas-pipeline/blob/master/media/openfaas-pipeline.jpg?raw=true)


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

##### Get minio config and edit the JSON to add webhook handler
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

##### Update minio config and restart mino server
Update the mini config and restart minio server
```sh
mc admin config set minio < myconfig.json
```

```sh
mc admin service restart minio
```

##### Add event for the webhook
```sh
mc event add minio/images arn:minio:sqs::1:webhook --event put --suffix .jpg
```

##### Create required buckets
```sh
mc mb minio/images
```
```sh
mc mb minio/images-thumbnail
```
```sh
mc mb minio/inception
```

##### Change `images` bucket policy to public so that inception function can download the image without secret
```sh
mc policy public minio/images
```
