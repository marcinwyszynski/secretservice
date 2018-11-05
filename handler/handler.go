package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
)

// Handler wraps a GraphQL schema to interface with AWS API Gateway.
type Handler struct {
	schema *graphql.Schema
}

// New returns an instance of a Handler.
func New(schema *graphql.Schema) *Handler {
	return &Handler{schema: schema}
}

// Handle serves as a main Lambda handler, translating API Gateway requests
// to GraphQL, and GraphQL responses back to API Gateway ones.
func (h *Handler) Handle(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var ret events.APIGatewayProxyResponse
	data, err := h.handle(ctx, event.Body)
	if err != nil {
		ret.StatusCode = http.StatusInternalServerError
		ret.Body = err.Error()
	} else {
		ret.StatusCode = http.StatusOK
		ret.Body = string(data)
	}
	return ret, err
}

func (h *Handler) handle(ctx context.Context, input string) ([]byte, error) {
	var req struct {
		OperationName string                 `json:"operationName"`
		Query         string                 `json:"query"`
		Variables     map[string]interface{} `json:"variables"`
	}

	if err := json.NewDecoder(strings.NewReader(input)).Decode(&req); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal request")
	}

	xray.AddAnnotation(ctx, "Operation", req.OperationName)

	ret, err := json.Marshal(h.schema.Exec(ctx, req.Query, req.OperationName, req.Variables))
	if err != nil {
		return nil, errors.Wrap(err, "could not write back the response")
	}
	return ret, nil
}
