package store

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/regnull/ubikom/pb"
	"google.golang.org/protobuf/proto"
)

type File struct {
	baseDir string
	maxAge  time.Duration
}

func NewFile(baseDir string, maxAge time.Duration) *File {
	return &File{baseDir: baseDir, maxAge: maxAge}
}

func (f *File) Save(msg *pb.DMSMessage, receiverKey []byte) error {
	receiverKeyStr := fmt.Sprintf("%x", receiverKey)
	fileDir := getReceiverDir(f.baseDir, receiverKeyStr)

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
	err = os.WriteFile(filePath, b, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (f *File) GetNext(receiverKey []byte) (*pb.DMSMessage, error) {
	receiverKeyStr := fmt.Sprintf("%x", receiverKey)

	fileDir := getReceiverDir(f.baseDir, receiverKeyStr)
	files, err := os.ReadDir(fileDir)

	if err != nil || len(files) == 0 {
		// Maybe directory doesn't exist, it's fine.
		return nil, nil
	}

	now := time.Now()
	for _, file := range files {
		filePath := path.Join(fileDir, file.Name())
		info, err := file.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to read file info: %w", err)
		}
		age := now.Sub(info.ModTime())
		if age > f.maxAge {
			// Delete file if it's too old.
			os.Remove(filePath)
			continue
		}
		b, err := os.ReadFile(filePath)
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
	return nil, nil
}

func (f *File) GetAll(receiverKey []byte) ([]*pb.DMSMessage, error) {
	receiverKeyStr := fmt.Sprintf("%x", receiverKey)

	fileDir := getReceiverDir(f.baseDir, receiverKeyStr)
	files, err := os.ReadDir(fileDir)

	if err != nil || len(files) == 0 {
		// Maybe directory doesn't exist, it's fine.
		return nil, nil
	}

	var ret []*pb.DMSMessage

	now := time.Now()
	for _, file := range files {
		filePath := path.Join(fileDir, file.Name())

		info, err := file.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to read file info: %w", err)
		}
		age := now.Sub(info.ModTime())
		if age > f.maxAge {
			// Delete file if it's too old.
			os.Remove(filePath)
			continue
		}

		b, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		msg := &pb.DMSMessage{}
		err = proto.Unmarshal(b, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %w", err)
		}
		ret = append(ret, msg)
	}

	return ret, nil
}

func (f *File) Remove(msg *pb.DMSMessage, receiverKey []byte) error {
	receiverKeyStr := fmt.Sprintf("%x", receiverKey)

	fileDir := getReceiverDir(f.baseDir, receiverKeyStr)
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
