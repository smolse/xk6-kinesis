package kinesis

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/dop251/goja"
	"github.com/mitchellh/mapstructure"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/kinesis", new(RootModule))
}

// RootModule is the global module object type. It is instantiated once per test run and will be used to create
// `k6/x/kinesis` module instances for each VU.
type RootModule struct{}

// ModuleInstance represents an instance of the JS module for every VU.
type ModuleInstance struct {
	vu modules.VU
	*Client
}

// Client represents the Client constructor (i.e., `new kinesis.Client()`) and returns a new Kinesis client object.
type Client struct {
	vu     modules.VU
	client *kinesis.Client
}

// Ensure the interfaces are implemented correctly.
var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &ModuleInstance{}
)

// NewModuleInstance implements the modules.Module interface to return a new instance for each VU.
func (*RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &ModuleInstance{vu: vu, Client: &Client{vu: vu}}
}

// Exports implements the modules.Instance interface and returns the exports of the JS module.
func (mi *ModuleInstance) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"Client": mi.NewClient,
		},
	}
}

// NewClient is the JS constructor for the Kinesis client based on the AWS SDK implementation.
//
// It resolves credentials using the SDK's default credential chain and allows the Kinesis endpoint URL to be redefined
// by passing it as the first argument to the constructor call.
func (mi *ModuleInstance) NewClient(call goja.ConstructorCall) *goja.Object {
	rt := mi.vu.Runtime()

	var endpointUrl string
	err := rt.ExportTo(call.Argument(0), &endpointUrl)
	if err != nil {
		common.Throw(rt, fmt.Errorf("unable to read the endpoint URL argument, %s", err))
	}

	endpointResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if endpointUrl != "" {
				return aws.Endpoint{
					URL: endpointUrl,
				}, nil
			}
			// Returning EndpointNotFoundError will allow the service to fallback to the default resolution
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		},
	)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolverWithOptions(endpointResolver))
	if err != nil {
		common.Throw(rt, fmt.Errorf("unable to read the AWS SDK's default external configurations, %s", err))
	}

	client := &Client{
		vu:     mi.vu,
		client: kinesis.NewFromConfig(cfg),
	}

	return rt.ToValue(client).ToObject(rt)
}

// PutRecord writes a single data record into a Kinesis data stream.
func (c *Client) PutRecord(putRecordInput interface{}) error {
	var kinesisPutRecordInput kinesis.PutRecordInput
	_ = mapstructure.Decode(putRecordInput, &kinesisPutRecordInput)
	_, err := c.client.PutRecord(context.TODO(), &kinesisPutRecordInput)
	return err
}

// PutRecords writes multiple data records into a Kinesis data stream in a single call.
func (c *Client) PutRecords(putRecordsInput interface{}) error {
	var kinesisPutRecordsInput kinesis.PutRecordsInput
	_ = mapstructure.Decode(putRecordsInput, &kinesisPutRecordsInput)
	_, err := c.client.PutRecords(context.TODO(), &kinesisPutRecordsInput)
	return err
}
