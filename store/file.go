package store

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"google.golang.org/protobuf/proto"
	"teralyt.com/ubikom/pb"
)

type File struct {
	baseDir string
}

func NewFile(baseDir string) *File {
	return &File{baseDir: baseDir}
}

func (f *File) Save(msg *pb.DMSMessage) error {
	if len(msg.GetReceiver()) < 10 || len(msg.GetSender()) < 10 {
		return fmt.Errorf("invalid message")
	}

	receiverKey := fmt.Sprintf("%x", msg.GetReceiver())
	fileDir := getReceiverDir(f.baseDir, receiverKey)

	b, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}
	fileName := fmt.Sprintf("%x", sha256.Sum256(b))

	filePath := path.Join(fileDir, fileName)
	err = os.MkdirAll(fileDir, 0770)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	err = ioutil.WriteFile(filePath, b, 0660)
	if err != nil {
		return err
	}
	return nil
}

func (f *File) GetNext(receiver []byte) (*pb.DMSMessage, error) {
	receiverKey := fmt.Sprintf("%x", receiver)

	fileDir := getReceiverDir(f.baseDir, receiverKey)
	files, err := ioutil.ReadDir(fileDir)

	if err != nil || len(files) == 0 {
		// Maybe directory doesn't exist, it's fine.
		return nil, nil
	}

	filePath := path.Join(fileDir, files[0].Name())

	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	msg := &pb.DMSMessage{}
	err = proto.Unmarshal(b, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return msg, nil
}

func (f *File) Remove(msg *pb.DMSMessage) error {
	if len(msg.GetReceiver()) < 10 || len(msg.GetSender()) < 10 {
		return fmt.Errorf("invalid message")
	}

	receiverKey := fmt.Sprintf("%x", msg.GetReceiver())
	fileDir := getReceiverDir(f.baseDir, receiverKey)
	b, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	fileName := fmt.Sprintf("%x", sha256.Sum256(b))
	filePath := path.Join(fileDir, fileName)
	err = os.Remove(filePath)
	if err != nil {
		return err
	}
	return nil
}

func getReceiverDir(baseDir string, receiverKey string) string {
	subDir1 := receiverKey[0:6]
	subDir2 := receiverKey[6:10]
	rest := receiverKey[10:]
	fileDir := path.Join(baseDir, subDir1, subDir2, rest)
	return fileDir
}
