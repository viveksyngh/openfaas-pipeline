package function

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	minio "github.com/minio/minio-go"
)

const (
	defaultEndpoint   = "minio:9000"
	secretPath        = "/var/openfaas/secrets/"
	accessKeySecret   = "s3-access-key"
	secretKeySecret   = "s3-secret-key"
	defaultBucketName = "inception"
)

// Handle a serverless request
func Handle(req []byte) string {
	endpoint := os.Getenv("s3_url")
	if len(endpoint) == 0 {
		endpoint = defaultEndpoint
	}

	accessKey := os.Getenv("s3_access_key")
	if accessKey == "" {
		accessKeyByte, err := ioutil.ReadFile(secretPath + accessKeySecret)
		if err != nil {
			return fmt.Sprintf("Cannot read secret %s", accessKeySecret)
		}
		accessKey = string(accessKeyByte)
	}

	secretKey := os.Getenv("s3_secret_key")
	if secretKey == "" {
		secretKeyByte, err := ioutil.ReadFile(secretPath + secretKeySecret)
		if err != nil {
			return fmt.Sprintf("Cannot read secret %s", secretKeySecret)
		}
		secretKey = string(secretKeyByte)
	}

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKey, secretKey, false)
	if err != nil {
		return err.Error()
	}

	bucketName := os.Getenv("bucket")
	location := "us-east-1"

	if len(bucketName) == 0 {
		bucketName = defaultBucketName
	}

	exists, err := minioClient.BucketExists(bucketName)
	if !exists {
		err = minioClient.MakeBucket(bucketName, location)
		if err != nil {
			return err.Error()
		}

	}

	queryParams, err := url.ParseQuery(os.Getenv("Http_Query"))
	if err != nil {
		return err.Error()
	}
	fileName := queryParams.Get("filename")

	var requestJSON interface{}
	err = json.Unmarshal(req, &requestJSON)
	if err != nil {
		return err.Error()
	}

	jsonBytes, err := json.MarshalIndent(requestJSON, "", "    ")
	filePath := "/tmp/" + fileName

	file, err := os.Create(filePath)
	if err != nil {
		return err.Error()
	}

	_, err = file.Write(jsonBytes)
	if err != nil {
		return err.Error()
	}

	n, err := minioClient.FPutObject(bucketName, fileName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Successfully uploaded %s of size %d\n", fileName, n)
}
