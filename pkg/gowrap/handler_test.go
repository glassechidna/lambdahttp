package gowrap

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestHandler_ApplicationLoadBalancer(t *testing.T) {
	pong := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	input, err := ioutil.ReadFile("testdata/alb_input.json")
	assert.NoError(t, err)

	wrap := ApplicationLoadBalancer(pong)
	output, err := wrap.Invoke(context.Background(), input)
	assert.NoError(t, err)

	expected, err := ioutil.ReadFile("testdata/alb_expected_output.json")
	assert.NoError(t, err)

	assert.JSONEq(t, string(expected), string(output))
}

func TestHandler_ApplicationLoadBalancer_CaseSensitivity(t *testing.T) {
	pong := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	input, err := ioutil.ReadFile("testdata/case_issue.json")
	assert.NoError(t, err)

	wrap := ApplicationLoadBalancer(pong)
	output, err := wrap.Invoke(context.Background(), input)
	assert.NoError(t, err)

	expected, err := ioutil.ReadFile("testdata/alb_expected_output.json")
	assert.NoError(t, err)

	assert.JSONEq(t, string(expected), string(output))
}
