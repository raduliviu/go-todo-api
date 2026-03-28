package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gin-gonic/gin"
)

// ginHandler adapts a Gin router to a Lambda handler function.
// It returns a closure so that r is captured once and reused across invocations.
func ginHandler(r *gin.Engine) func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		// Build the request body.
		// API Gateway may base64-encode binary payloads — decode if needed.
		var body []byte
		if req.IsBase64Encoded {
			decoded, err := base64.StdEncoding.DecodeString(req.Body)
			if err != nil {
				return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, err
			}
			body = decoded
		} else {
			body = []byte(req.Body)
		}

		// Build the full URL from path + query string parameters.
		params := url.Values{}
		for key, value := range req.QueryStringParameters {
			params.Add(key, value)
		}
		fullURL := req.Path
		if encoded := params.Encode(); encoded != "" {
			fullURL += "?" + encoded
		}

		// Construct an http.Request using the Lambda context so that
		// Lambda's deadline and cancellation signals are respected.
		newReq, err := http.NewRequestWithContext(ctx, req.HTTPMethod, fullURL, bytes.NewReader(body))
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
		}

		// Forward the API Gateway headers to the HTTP request.
		for k, v := range req.Headers {
			newReq.Header.Set(k, v)
		}

		// Run the request through Gin. The recorder captures the response
		// in memory instead of writing to a real network connection.
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, newReq)

		// Convert the recorded response back to the format Lambda expects.
		// MultiValueHeaders preserves headers with multiple values (e.g. Set-Cookie).
		return events.APIGatewayProxyResponse{
			StatusCode:        recorder.Code,
			MultiValueHeaders: recorder.Header(),
			Body:              recorder.Body.String(),
		}, nil
	}
}
