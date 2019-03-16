package function

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Jeffail/gabs"
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
	requestBody := RequestBody{
		Bucket:    bucket,
		ObjectKey: objectName,
	}
	requestBytes, _ := json.Marshal(requestBody)

	return fmt.Sprintf("Hello, Go. You said: %s", string(requestBytes))
}
