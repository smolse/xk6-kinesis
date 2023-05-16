package kinesis

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/eventloop"
	"go.k6.io/k6/lib"
)

// mockVU is a mock implementation of the k6 virtual user interface.
type mockVU struct {
	runtime *goja.Runtime
	*eventloop.EventLoop
	mock.Mock
}

func newMockVU() *mockVU {
	return &mockVU{runtime: goja.New()}
}

func (mvu *mockVU) Context() context.Context {
	return context.TODO()
}

func (mvu *mockVU) InitEnv() *common.InitEnvironment {
	return nil
}

func (mvu *mockVU) State() *lib.State {
	return nil
}

func (mvu *mockVU) Runtime() *goja.Runtime {
	return mvu.runtime
}

func (mvu *mockVU) RegisterCallback() (enqueueCallback func(func() error)) {
	if mvu.EventLoop == nil {
		mvu.EventLoop = eventloop.New(mvu)
	}
	return mvu.EventLoop.RegisterCallback()
}

// TestLoadAwsSdkConfig validates behavior of the loadAwsSdkConfig function.
func TestLoadAwsSdkConfig(t *testing.T) {
	assert := assert.New(t)
	var tests = []struct {
		Name                         string
		CredentialsSetUp             func()
		Endpoint                     string
		ExpectedCredentialsSource    string
		ExpectedResolveEndpointError error
		ExpectedResolveEndpointURL   string
	}{
		{
			"use an access key and a default endpoint",
			func() {
				t.Setenv("AWS_ACCESS_KEY_ID", "foo")
				t.Setenv("AWS_SECRET_ACCESS_KEY", "bar")
				t.Setenv("AWS_SESSION_TOKEN", "baz")
				t.Setenv("AWS_REGION", "us-west-2")
			},
			"",
			"EnvConfigCredentials",
			&aws.EndpointNotFoundError{},
			"",
		},
		{
			"use an access key and a custom endpoint",
			func() {
				t.Setenv("AWS_ACCESS_KEY_ID", "foo")
				t.Setenv("AWS_SECRET_ACCESS_KEY", "bar")
				t.Setenv("AWS_SESSION_TOKEN", "baz")
				t.Setenv("AWS_REGION", "us-west-2")
			},
			"http://localhost:4566",
			"EnvConfigCredentials",
			nil,
			"http://localhost:4566",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.CredentialsSetUp()
			rt := goja.New()
			var args []goja.Value
			if tt.Endpoint != "" {
				args = []goja.Value{rt.ToValue(tt.Endpoint)}
			}
			call := &goja.ConstructorCall{Arguments: args}
			cfg := loadAwsSdkConfig(rt, *call)

			creds, err := cfg.Credentials.Retrieve(context.TODO())
			assert.Empty(err)
			assert.Equal(tt.ExpectedCredentialsSource, creds.Source)

			endpoint, err := cfg.EndpointResolverWithOptions.ResolveEndpoint("kinesis", cfg.Region)
			assert.Equal(tt.ExpectedResolveEndpointError, err)
			assert.Equal(tt.ExpectedResolveEndpointURL, endpoint.URL)
		})
	}
}

// TestExports validates the extension exports.
func TestExports(t *testing.T) {
	assert := assert.New(t)
	mi := (&RootModule{}).NewModuleInstance(newMockVU())
	call := goja.ConstructorCall{}

	// Make sure that "Client" exported in the JS module is a synchronous client
	client := mi.Exports().Named["Client"].(func(goja.ConstructorCall) *goja.Object)(call)
	assert.Equal("*kinesis.SyncClient", client.ExportType().String())

	// Make sure that "AsyncClient" exported in the JS module is an asynchronous client
	asyncClient := mi.Exports().Named["AsyncClient"].(func(goja.ConstructorCall) *goja.Object)(call)
	assert.Equal("*kinesis.AsyncClient", asyncClient.ExportType().String())
}
