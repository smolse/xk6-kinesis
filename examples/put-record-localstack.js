import kinesis from "k6/x/kinesis";

const client = kinesis.Client("http://localhost:4566");

export default function () {
    const putRecordInput = {
        // Data contains a byte array representation of the "hello world" string
        Data: [0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64],
        PartitionKey: "partition-key",
        StreamName: "xk6-kinesis-test"
    }

    client.putRecord(putRecordInput);
}
