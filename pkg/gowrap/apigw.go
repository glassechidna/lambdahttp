package gowrap

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

func ApiGateway(handler http.Handler) lambda.Handler {
	return &apigw{handler}
}

type apigw struct {
	http.Handler
}

func (a *apigw) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	input := events.APIGatewayProxyRequest{}
	err := json.Unmarshal(payload, &input)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ctx = context.WithValue(ctx, contextKey, &input)

	w := httptest.NewRecorder()
	r := NewHttpRequest(apiGwToAlb(input))
	r = r.WithContext(ctx)
	a.ServeHTTP(w, r)

	httpResp := w.Result()
	rawBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	b64 := base64.StdEncoding.EncodeToString(rawBody)
	response := events.APIGatewayProxyResponse{
		StatusCode:        httpResp.StatusCode,
		MultiValueHeaders: httpResp.Header,
		Body:              b64,
		IsBase64Encoded:   true,
	}

	payload, err = json.Marshal(response)
	return payload, errors.WithStack(err)
}

type key struct{}
var contextKey = &key{}

func ApiGatewayRequest(ctx context.Context) *events.APIGatewayProxyRequest {
	r, _ := ctx.Value(contextKey).(*events.APIGatewayProxyRequest)
	return r
}
