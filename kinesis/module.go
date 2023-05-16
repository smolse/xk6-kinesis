// Package kinesis implements a Kinesis client for k6.
package kinesis

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/dop251/goja"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
)

type (
	// RootModule is the global module object type. It is instantiated once per test run and will be used to create
	// `k6/x/kinesis` module instances for each VU.
	RootModule struct{}

	// ModuleInstance represents an instance of the JS module for every VU.
	ModuleInstance struct {
		vu modules.VU
	}
)

// Ensure the interfaces are implemented correctly.
var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &ModuleInstance{}
)

// NewModuleInstance implements the modules.Module interface to return a new instance for each VU.
func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ModuleInstance{vu: vu}
}

// Exports implements the modules.Instance interface and returns the exports of the JS module.
func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"Client":      mi.NewClient,
			"AsyncClient": mi.NewAsyncClient,
		},
	}
}

// loadAwsSdkConfig resolves AWS credentials using the SDK's default credential chain and allows the Kinesis endpoint
// URL to be redefined by passing it as the first argument to the constructor call.
func loadAwsSdkConfig(rt *goja.Runtime, call goja.ConstructorCall) aws.Config {
	var endpointURL string
	err := rt.ExportTo(call.Argument(0), &endpointURL)
	if err != nil {
		common.Throw(rt, fmt.Errorf("unable to read the endpoint URL argument, %w", err))
	}

	endpointResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if endpointURL != "" {
				return aws.Endpoint{
					URL: endpointURL,
				}, nil
			}
			// Returning EndpointNotFoundError will allow the service to fallback to the default resolution
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		},
	)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(endpointResolver))
	if err != nil {
		common.Throw(rt, fmt.Errorf("unable to read the AWS SDK's default external configurations, %w", err))
	}
	return cfg
}

// NewClient is the JS constructor for the default synchronous Kinesis client based on the AWS SDK implementation.
func (mi *ModuleInstance) NewClient(call goja.ConstructorCall) *goja.Object {
	rt := mi.vu.Runtime()
	cfg := loadAwsSdkConfig(rt, call)
	client := &SyncClient{vu: mi.vu, client: kinesis.NewFromConfig(cfg)}
	return rt.ToValue(client).ToObject(rt)
}

// NewAsyncClient is the JS constructor for the asynchronous Kinesis client based on the AWS SDK implementation.
func (mi *ModuleInstance) NewAsyncClient(call goja.ConstructorCall) *goja.Object {
	rt := mi.vu.Runtime()
	cfg := loadAwsSdkConfig(rt, call)
	client := &AsyncClient{vu: mi.vu, client: kinesis.NewFromConfig(cfg)}
	return rt.ToValue(client).ToObject(rt)
}
