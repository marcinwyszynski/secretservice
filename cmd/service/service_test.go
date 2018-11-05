package main

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
)

func TestBuildHandler(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY_ID", "accesskey")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
	session := session.Must(session.NewSession())

	handler, err := buildHandler(session, new(config))

	assert.NotNil(t, handler)
	assert.NoError(t, err)
}
