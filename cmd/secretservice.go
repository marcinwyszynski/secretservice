package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-xray-sdk-go/xray"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/marcinwyszynski/secretservice"
	"github.com/marcinwyszynski/secretservice/backend"
	"github.com/marcinwyszynski/secretservice/handler"
	"github.com/marcinwyszynski/secretservice/resolver"
	"github.com/marcinwyszynski/ssmvars"
)

func main() {
	session := session.Must(session.NewSession())

	ssmAPI := ssm.New(session)
	xray.AWS(ssmAPI.Client)

	s3API := s3.New(session)
	xray.AWS(s3API.Client)

	ssmvars := ssmvars.New(ssmAPI, mustGetEnv("SSM_PREFIX"), mustGetEnv("KMS_KEY_ID"))
	backend := backend.New(ssmvars, s3API, mustGetEnv("S3_BUCKET_NAME"))
	schema := graphql.MustParseSchema(secretservice.Schema, resolver.New(backend))

	lambda.Start(handler.New(schema).Handle)
}
