package kinesis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/dop251/goja"
	"github.com/mitchellh/mapstructure"
	"go.k6.io/k6/js/modules"
)

// client is the interface for the supported subset of the AWS SDK Kinesis client.
type client interface {
	PutRecord(context.Context, *kinesis.PutRecordInput, ...func(*kinesis.Options)) (*kinesis.PutRecordOutput, error)
	PutRecords(context.Context, *kinesis.PutRecordsInput, ...func(*kinesis.Options)) (*kinesis.PutRecordsOutput, error)
}

// SyncClient represents the synchronous Client constructor (i.e., `new kinesis.Client()`) and returns a new
// synchronous Kinesis client object.
type SyncClient struct {
	vu modules.VU
	client
}

// PutRecord synchronously writes a single data record into a Kinesis data stream.
func (c *SyncClient) PutRecord(putRecordInput interface{}) (interface{}, error) {
	var kinesisPutRecordInput kinesis.PutRecordInput
	err := mapstructure.Decode(putRecordInput, &kinesisPutRecordInput)
	if err != nil {
		return nil, err
	}
	output, err := c.client.PutRecord(context.TODO(), &kinesisPutRecordInput)
	return output, err
}

// PutRecords synchronously writes multiple data records into a Kinesis data stream in a single call.
func (c *SyncClient) PutRecords(putRecordsInput interface{}) (interface{}, error) {
	var kinesisPutRecordsInput kinesis.PutRecordsInput
	err := mapstructure.Decode(putRecordsInput, &kinesisPutRecordsInput)
	if err != nil {
		return nil, err
	}
	output, err := c.client.PutRecords(context.TODO(), &kinesisPutRecordsInput)
	return output, err
}

// AsyncClient represents the AsyncClient constructor (i.e., `new kinesis.AsyncClient()`) and returns a new
// asynchronous Kinesis client object.
type AsyncClient struct {
	vu modules.VU
	client
}

// makeHandledPromise will create a promise and return its resolve and reject methods, wrapped in such a way that it
// will block the eventloop from exiting before they are called even if the promise isn't resolved by the time the
// current script ends executing.
func (c *AsyncClient) makeHandledPromise() (*goja.Promise, func(interface{}), func(interface{})) {
	runtime := c.vu.Runtime()
	callback := c.vu.RegisterCallback()
	promise, resolve, reject := runtime.NewPromise()

	return promise,
		func(i interface{}) {
			callback(
				func() error {
					resolve(i)
					return nil
				},
			)
		},
		func(i interface{}) {
			callback(
				func() error {
					reject(i)
					return nil
				},
			)
		}
}

// PutRecord asynchronously writes a single data record into a Kinesis data stream.
func (c *AsyncClient) PutRecord(putRecordInput interface{}) *goja.Promise {
	promise, resolve, reject := c.makeHandledPromise()

	go func() {
		var kinesisPutRecordInput kinesis.PutRecordInput
		err := mapstructure.Decode(putRecordInput, &kinesisPutRecordInput)
		if err != nil {
			reject(err)
			return
		}
		output, err := c.client.PutRecord(context.TODO(), &kinesisPutRecordInput)
		if err != nil {
			reject(err)
			return
		}
		resolve(output)
	}()

	return promise
}

// PutRecords asynchronously writes multiple data records into a Kinesis data stream in a single call.
func (c *AsyncClient) PutRecords(putRecordsInput interface{}) *goja.Promise {
	promise, resolve, reject := c.makeHandledPromise()

	go func() {
		var kinesisPutRecordsInput kinesis.PutRecordsInput
		err := mapstructure.Decode(putRecordsInput, &kinesisPutRecordsInput)
		if err != nil {
			reject(err)
			return
		}
		output, err := c.client.PutRecords(context.TODO(), &kinesisPutRecordsInput)
		if err != nil {
			reject(err)
			return
		}
		resolve(output)
	}()

	return promise
}
