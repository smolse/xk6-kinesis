# xk6-kinesis

This is a [Kinesis Data Streams](https://aws.amazon.com/kinesis/data-streams/) client library for
[k6](https://github.com/grafana/k6), implemented as an extension using the [xk6](https://github.com/grafana/xk6) system.

| :exclamation: This is a proof of concept, isn't supported by the k6 team, and may break in the future. USE AT YOUR OWN RISK! |
| ---------------------------------------------------------------------------------------------------------------------------- |

## Build

To build a `k6` binary with this extension, first ensure you have the prerequisites:

- [Go toolchain](https://go101.org/article/go-toolchain.html)
- Git

Then:

1. Install `xk6`:
  ```shell
  go install go.k6.io/xk6/cmd/xk6@latest
  ```

2. Build the binary:
  ```shell
  xk6 build --with github.com/smolse/xk6-kinesis
  ```

## Usage

### AWS Credentials

This plugin uses the AWS SDK Go v2 default credential chain. It looks for credentials in the following order:

1. Environment variables.
   1. Static Credentials (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`)
   2. Web Identity Token (`AWS_WEB_IDENTITY_TOKEN_FILE`)
2. Shared configuration files.
   1. SDK defaults to `credentials` file under `.aws` folder that is placed in the home folder on your computer.
   2. SDK defaults to `config` file under `.aws` folder that is placed in the home folder on your computer.
3. If your application uses an ECS task definition or RunTask API operation, IAM role for tasks.
4. If your application is running on an Amazon EC2 instance, IAM role for Amazon EC2.

Source: https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials

### API

Currently, `xk6-kinesis` exposes a small subset of Kinesis API actions that may be extended in the future:
* [PutRecord](https://docs.aws.amazon.com/kinesis/latest/APIReference/API_PutRecord.html)
* [PutRecords](https://docs.aws.amazon.com/kinesis/latest/APIReference/API_PutRecords.html)

This extension implements a synchronous client as well as an asynchronous client that can be used in a promise chain or
via `async/await` since k6 version [v0.43](https://k6.io/blog/k6-product-update-v0-43/). See
[examples/put-record-localstack.js](examples/put-record-localstack.js) and
[examples/put-record-localstack-async.js](examples/put-record-localstack-async.js) for examples.

### Example

The following example writes a batch of records to a Kinesis data stream.

```javascript
// examples/put-records.js
import kinesis from "k6/x/kinesis";

const client = kinesis.Client();

function asciiStringToByteArray(str) {
    var bytes = [];
    for(var i = 0; i < str.length; i++) {
        var char = str.charCodeAt(i);
        bytes.push(char & 0xFF);
    }
    return bytes;
}

export default function () {
    const putRecordsInput = {
        Records: [
            {
                Data: asciiStringToByteArray("foo"),
                PartitionKey: "PK"
            },
            {
                Data: asciiStringToByteArray("bar"),
                PartitionKey: "PK"
            },
            {
                Data: asciiStringToByteArray("baz"),
                PartitionKey: "PK"
            }
        ],
        StreamName: "STREAM_NAME"
    }

    client.putRecords(putRecordsInput);
}
```

```shell
$ export AWS_ACCESS_KEY_ID=<access key id>
$ export AWS_SECRET_ACCESS_KEY=<secret access key>
$ export AWS_REGION=<region>
$ ./k6 run examples/put-records.js
```

Result output:

```shell
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: examples/put-records.js
     output: -

  scenarios: (100.00%) 1 scenario, 1 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 1 iterations for each of 1 VUs (maxDuration: 10m0s, gracefulStop: 30s)


     data_received........: 0 B 0 B/s
     data_sent............: 0 B 0 B/s
     iteration_duration...: avg=118.4ms min=118.4ms med=118.4ms max=118.4ms p(90)=118.4ms p(95)=118.4ms
     iterations...........: 1   8.42726/s


running (00m00.1s), 0/1 VUs, 1 complete and 0 interrupted iterations
default ✓ [======================================] 1 VUs  00m00.1s/10m0s  1/1 iters, 1 per VU
```
