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
