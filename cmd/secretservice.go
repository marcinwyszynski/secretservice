package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-xray-sdk-go/xray"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/marcinwyszynski/secretservice"
	"github.com/marcinwyszynski/secretservice/backend"
	"github.com/marcinwyszynski/secretservice/handler"
	"github.com/marcinwyszynski/secretservice/resolver"
	"github.com/marcinwyszynski/ssmvars"
)

type config struct {
	BucketName string `envconfig:"S3_BUCKET_NAME" required:"true"`
	KMSKeyID   string `envconfig:"KMS_KEY_ID"`
	LogLevel   string `envconfig:"LOG_LEVEL" default:"INFO"`
	SSMPrefix  string `envconfig:"SSM_PREVIX" required:"true"`
}

func main() {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Error parsing function configuration: %v", err)
	}

	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Invalid log level %s: %v", cfg.LogLevel, err)
	}
	log.SetLevel(level)

	log.Debug("Starting AWS session")
	session := session.Must(session.NewSession())

	log.Debug("Creating SSM API client")
	ssmAPI := ssm.New(session)
	xray.AWS(ssmAPI.Client)

	log.Debug("Creating S3 API client")
	s3API := s3.New(session)
	xray.AWS(s3API.Client)

	log.Debug("Setting up SSM variables handler")
	ssmvars := ssmvars.New(ssmAPI, cfg.SSMPrefix, cfg.KMSKeyID)

	log.Debug("Setting up backend")
	backend := backend.New(ssmvars, s3API, cfg.BucketName)

	log.Debug("Setting up GraphQL schema")
	schema := graphql.MustParseSchema(secretservice.Schema, resolver.New(backend))

	log.Info("Starting Lambda server")
	lambda.Start(handler.New(schema).Handle)
}
