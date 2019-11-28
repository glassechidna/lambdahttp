package proxy

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cmd := exec.Command("testdata/e2e.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	require.NoError(t, err)

	zip, err := ioutil.ReadFile("../../lambda.zip")
	require.NoError(t, err)

	sess, err := session.NewSession(aws.NewConfig().WithRegion("ap-southeast-2"))
	require.NoError(t, err)

	api := lambda.New(sess)
	updateResp, err := api.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String("lambdahttptest"),
		Publish:      aws.Bool(true),
		ZipFile:      zip,
	})
	require.NoError(t, err)

	payload, err := ioutil.ReadFile("testdata/alb_input.json")
	require.NoError(t, err)

	invokeResp, err := api.Invoke(&lambda.InvokeInput{
		FunctionName: updateResp.FunctionArn,
		Payload:      payload,
	})
	require.NoError(t, err)

	expected, err := ioutil.ReadFile("testdata/alb_expected_output.json")
	assert.NoError(t, err)
	assert.JSONEq(t, string(expected), normalizeDateInResponse(invokeResp.Payload))
}

func normalizeDateInResponse(payload []byte) string {
	response := string(payload)
	regex := regexp.MustCompile(`"Date":"([^"]+)"`)
	matches := regex.FindStringSubmatch(response)
	returnedDate := matches[1]
	return strings.ReplaceAll(response, returnedDate, "Sun, 06 Oct 2019 06:53:36 GMT")
}
