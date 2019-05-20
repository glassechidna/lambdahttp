package lambdahttp

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestHandler_ApplicationLoadBalancer(t *testing.T) {
	req := events.ALBTargetGroupRequest{}
	f, err := os.Open("testdata/alb_input.json")
	assert.NoError(t, err)
	defer f.Close()

	err = json.NewDecoder(f).Decode(&req)
	assert.NoError(t, err)

	pong := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	wrap := &Handler{Handler: pong}

	resp, err := wrap.ApplicationLoadBalancer(context.Background(), req)
	assert.NoError(t, err)

	f, err = os.Open("testdata/alb_expected_output.json")
	assert.NoError(t, err)
	defer f.Close()

	expectedResp := events.ALBTargetGroupResponse{}
	err = json.NewDecoder(f).Decode(&expectedResp)
	assert.NoError(t, err)

	assert.Equal(t, expectedResp, resp)
}
