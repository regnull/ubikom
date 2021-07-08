package protoio

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/regnull/ubikom/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func Test_ReaderWriter(t *testing.T) {
	assert := assert.New(t)

	msg := pb.DMSMessage{Sender: "foo", Receiver: "bar"}
	var buf bytes.Buffer
	hashWriter := NewSha256Writer(&buf)
	writer := NewWriter(hashWriter)
	err := writer.Write(&msg)
	assert.NoError(err)

	assert.True(len(buf.Bytes()) > 6)
	assert.EqualValues("d34eb71a6a9b9a077536f4b29bc454885e4a17bddce56dcd1910fc8fea327429", fmt.Sprintf("%x", hashWriter.Hash()))

	readBuf := bytes.NewBuffer(buf.Bytes())
	reader := NewReader(readBuf)
	msg1, err := reader.Read(func(b []byte) (proto.Message, error) {
		msg := &pb.DMSMessage{}
		err := proto.Unmarshal(b, msg)
		return msg, err
	})
	assert.NoError(err)
	assert.NotNil(msg1)

	assert.EqualValues(msg.Sender, msg1.(*pb.DMSMessage).Sender)
	assert.EqualValues(msg.Receiver, msg1.(*pb.DMSMessage).Receiver)
}
