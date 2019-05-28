package lambdahttp

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewRequest(t *testing.T) {
	req := events.ALBTargetGroupRequest{}
	f, err := os.Open("testdata/base64_issue.json")
	assert.NoError(t, err)
	defer f.Close()

	err = json.NewDecoder(f).Decode(&req)
	assert.NoError(t, err)

	httpReq := NewHttpRequest(req)
	body, err := ioutil.ReadAll(httpReq.Body)
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(body))
}

