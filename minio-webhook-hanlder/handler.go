package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jeffail/gabs"
)

const (
	defaultS3URL      = "minio:9000"
	defaultGatewayURL = "gateway:8080"
)

//RequestBody request body for resizer function
type RequestBody struct {
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"objectKey"`
}

// Handle a serverless request
func Handle(req []byte) string {
	fmt.Fprintln(os.Stderr, string(req))
	parsedJSON, _ := gabs.ParseJSON(req)
	record, err := parsedJSON.ArrayElementP(0, "Records")
	if err != nil {
		return fmt.Sprintf("Invalid webhook request data")
	}

	bucket, _ := record.Path("s3.bucket.name").Data().(string)
	objectName, _ := record.Path("s3.object.key").Data().(string)

	err = invokeInception(bucket, objectName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: "+err.Error())
		return err.Error()
	}

	requestBody := RequestBody{
		Bucket:    bucket,
		ObjectKey: objectName,
	}
	requestBytes, _ := json.Marshal(requestBody)
	err = invokeImageResizer(requestBytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: "+err.Error())
		return err.Error()
	}

	return fmt.Sprintf("Hello, Go. You said: %s", string(requestBytes))
}

func invokeInception(bucket, objectKeystring string) error {
	s3URL := os.Getenv("s3_url")
	if len(s3URL) == 0 {
		s3URL = defaultS3URL
	}

	gatewayHostname := os.Getenv("gateway_url")
	if gatewayHostname == "" {
		gatewayHostname = "gateway:8080"
	}

	ext := filepath.Ext(objectKeystring)
	filename := strings.TrimSuffix(objectKeystring, ext) + ".json"

	imageURL := fmt.Sprintf("http://%s/%s/%s", s3URL, bucket, objectKeystring)
	inceptionFnAsyncURL := fmt.Sprintf("http://%s/async-function/inception", gatewayHostname)
	callbackFunctionURL := fmt.Sprintf("http://%s/function/json-uploader?filename=%s", gatewayHostname, filename)

	reader := bytes.NewReader([]byte(imageURL))
	client := http.Client{}
	request, err := http.NewRequest("POST", inceptionFnAsyncURL, reader)
	if err != nil {
		return err
	}
	request.Header.Add("X-Callback-url", callbackFunctionURL)

	res, err := client.Do(request)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Inception Success: "+res.Status)
	return nil
}

func invokeImageResizer(reqBytes []byte) error {
	gatewayHostname := os.Getenv("gateway_url")
	if gatewayHostname == "" {
		gatewayHostname = "gateway:8080"
	}

	imageResizerFnURL := fmt.Sprintf("http://%s/async-function/image-resizer", gatewayHostname)

	reader := bytes.NewReader(reqBytes)
	client := http.Client{}
	request, err := http.NewRequest("POST", imageResizerFnURL, reader)
	if err != nil {
		return err
	}

	res, err := client.Do(request)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Image Resizer Success: "+res.Status)
	return nil
}
