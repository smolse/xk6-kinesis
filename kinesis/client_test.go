package kinesis

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockKinesisClient is a mock implementation of the AWS SDK Kinesis client.
type mockKinesisClient struct {
	mock.Mock
}

func (mc *mockKinesisClient) PutRecord(
	context.Context,
	*kinesis.PutRecordInput,
	...func(*kinesis.Options),
) (*kinesis.PutRecordOutput, error) {
	args := mc.Called()
	return args.Get(0).(*kinesis.PutRecordOutput), args.Error(1)
}

func (mc *mockKinesisClient) PutRecords(
	context.Context,
	*kinesis.PutRecordsInput,
	...func(*kinesis.Options),
) (*kinesis.PutRecordsOutput, error) {
	args := mc.Called()
	return args.Get(0).(*kinesis.PutRecordsOutput), args.Error(1)
}

// TestSyncClient validates the synchronous client implementation.
func TestSyncClient(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		Name          string
		Action        string
		Input         interface{}
		Output        interface{}
		OutputError   error
		ExpectedError string
	}{
		{
			"sync client PutRecord method invocation receives invalid input",
			"PutRecord",
			map[string]interface{}{
				"Data": 123, // must be a byte array
			},
			&kinesis.PutRecordOutput{},
			nil,
			"1 error(s) decoding:\n\n* 'Data': source data must be an array or slice, got int",
		},
		{
			"sync client PutRecord method invocation is throttled",
			"PutRecord",
			map[string]interface{}{
				"Data": []byte("hello world"),
			},
			&kinesis.PutRecordOutput{},
			&types.ProvisionedThroughputExceededException{
				Message: aws.String("<error msg>"),
			},
			"ProvisionedThroughputExceededException: <error msg>",
		},
		{
			"sync client PutRecord method invocation is successful",
			"PutRecord",
			map[string]interface{}{
				"Data":         []byte("hello world"),
				"PartitionKey": "pk",
				"StreamName":   "stream",
			},
			&kinesis.PutRecordOutput{
				SequenceNumber: aws.String("01"),
			},
			nil,
			"",
		},
		{
			"sync client PutRecords method invocation receives invalid input",
			"PutRecords",
			map[string]interface{}{
				"Records": []map[string]interface{}{
					{"PartitionKey": 123}, // must be a string
				},
			},
			&kinesis.PutRecordsOutput{},
			nil,
			"1 error(s) decoding:\n\n* 'Records[0].PartitionKey' expected type 'string', got unconvertible type 'int', value: '123'",
		},
		{
			"sync client PutRecords method invocation is throttled",
			"PutRecords",
			map[string]interface{}{
				"Records": []map[string]interface{}{
					{
						"Data":         []byte("hello world"),
						"PartitionKey": "pk",
					},
				},
				"StreamName": "stream",
			},
			&kinesis.PutRecordsOutput{
				FailedRecordCount: aws.Int32(0),
			},
			&types.ProvisionedThroughputExceededException{
				Message: aws.String("<error msg>"),
			},
			"ProvisionedThroughputExceededException: <error msg>",
		},
		{
			"sync client PutRecords method invocation is successful",
			"PutRecords",
			map[string]interface{}{
				"Records": []map[string]interface{}{
					{
						"Data":         []byte("hello world"),
						"PartitionKey": "pk",
					},
				},
				"StreamName": "stream",
			},
			&kinesis.PutRecordsOutput{
				FailedRecordCount: aws.Int32(0),
			},
			nil,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			mockSdkClient := &mockKinesisClient{}
			client := &SyncClient{vu: newMockVU(), client: mockSdkClient}
			mockSdkClient.On(tt.Action).Return(tt.Output, tt.OutputError)

			var output interface{}
			var err error
			switch tt.Action {
			case "PutRecord":
				output, err = client.PutRecord(tt.Input)
			case "PutRecords":
				output, err = client.PutRecords(tt.Input)
			}

			if tt.ExpectedError != "" {
				assert.Equal(tt.ExpectedError, err.Error())
			} else {
				assert.Equal(tt.Output, output)
			}
		})
	}
}

// TestAsyncClient validates the asynchronous client implementation.
func TestAsyncClient(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		Name                  string
		Action                string
		Input                 interface{}
		Output                interface{}
		OutputError           error
		ExpectedPromiseState  goja.PromiseState
		ExpectedPromiseResult string
	}{
		{
			"async client PutRecord method invocation receives invalid input",
			"PutRecord",
			map[string]interface{}{
				"Data": 123, // must be a byte array
			},
			&kinesis.PutRecordOutput{},
			nil,
			goja.PromiseStateRejected,
			"1 error(s) decoding:\n\n* 'Data': source data must be an array or slice, got int",
		},
		{
			"async client PutRecord method invocation is throttled",
			"PutRecord",
			map[string]interface{}{
				"Data": []byte("hello world"),
			},
			&kinesis.PutRecordOutput{},
			&types.ProvisionedThroughputExceededException{
				Message: aws.String("<error msg>"),
			},
			goja.PromiseStateRejected,
			"ProvisionedThroughputExceededException: <error msg>",
		},
		{
			"async client PutRecord method invocation is successful",
			"PutRecord",
			map[string]interface{}{
				"Data":         []byte("hello world"),
				"PartitionKey": "pk",
				"StreamName":   "stream",
			},
			&kinesis.PutRecordOutput{
				SequenceNumber: aws.String("01"),
			},
			nil,
			goja.PromiseStateFulfilled,
			"[object Object]",
		},
		{
			"async client PutRecords method invocation receives invalid input",
			"PutRecords",
			map[string]interface{}{
				"Records": []map[string]interface{}{
					{"PartitionKey": 123}, // must be a string
				},
			},
			&kinesis.PutRecordsOutput{},
			nil,
			goja.PromiseStateRejected,
			"1 error(s) decoding:\n\n* 'Records[0].PartitionKey' expected type 'string', got unconvertible type 'int', value: '123'",
		},
		{
			"async client PutRecords method invocation is throttled",
			"PutRecords",
			map[string]interface{}{
				"Records": []map[string]interface{}{
					{
						"Data":         []byte("hello world"),
						"PartitionKey": "pk",
					},
				},
				"StreamName": "stream",
			},
			&kinesis.PutRecordsOutput{},
			&types.ProvisionedThroughputExceededException{
				Message: aws.String("<error msg>"),
			},
			goja.PromiseStateRejected,
			"ProvisionedThroughputExceededException: <error msg>",
		},
		{
			"async client PutRecords method invocation is successful",
			"PutRecords",
			map[string]interface{}{
				"Records": []map[string]interface{}{
					{
						"Data":         []byte("hello world"),
						"PartitionKey": "pk",
					},
				},
				"StreamName": "stream",
			},
			&kinesis.PutRecordsOutput{
				FailedRecordCount: aws.Int32(0),
			},
			nil,
			goja.PromiseStateFulfilled,
			"[object Object]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			mockSdkClient := &mockKinesisClient{}
			client := &AsyncClient{vu: newMockVU(), client: mockSdkClient}
			mockSdkClient.On(tt.Action).Return(tt.Output, tt.OutputError)
			var promise *goja.Promise
			switch tt.Action {
			case "PutRecord":
				promise = client.PutRecord(tt.Input)
			case "PutRecords":
				promise = client.PutRecords(tt.Input)
			}

			_ = client.vu.(*mockVU).EventLoop.Start(func() error { return nil })
			assert.Equal(tt.ExpectedPromiseState, promise.State())
			assert.Equal(tt.ExpectedPromiseResult, promise.Result().String())
		})
	}
}
