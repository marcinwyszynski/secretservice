package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
)

type gqlRequest struct {
	OperationName string                 `json:"operationName"`
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
}

type Handler struct {
	schema *graphql.Schema
}

func New(schema *graphql.Schema) *Handler {
	return &Handler{schema: schema}
}

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
	var gql gqlRequest
	if err := json.NewDecoder(strings.NewReader(input)).Decode(&gql); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal request")
	}
	ret, err := json.Marshal(h.schema.Exec(ctx, gql.Query, gql.OperationName, gql.Variables))
	if err != nil {
		return nil, errors.Wrap(err, "could not write back the reponse")
	}
	return ret, nil
}
