package gowrap

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httptest"
)

func ApplicationLoadBalancer(handler http.Handler) lambda.Handler {
	return &alb{handler}
}

type alb struct {
	http.Handler
}

func (a *alb) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	input := events.ALBTargetGroupRequest{}
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	w := httptest.NewRecorder()
	r := NewHttpRequest(input)
	r = r.WithContext(ctx)
	a.ServeHTTP(w, r)

	httpResp := w.Result()
	response, err := NewLambdaResponse(httpResp)
	if err != nil {
		return nil, err
	}

	payload, err = json.Marshal(response)
	return payload, errors.WithStack(err)
}
