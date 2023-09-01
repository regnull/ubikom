package store

import (
	"crypto/sha256"
	"fmt"

	"github.com/regnull/ubikom/pb"
	"google.golang.org/protobuf/proto"
)

type MemStore struct {
	data map[string]map[string]*pb.DMSMessage
}

func NewMem() Store {
	return &MemStore{
		data: make(map[string]map[string]*pb.DMSMessage),
	}
}

func (s *MemStore) Save(msg *pb.DMSMessage, receiverKey []byte) error {
	receiverKeyStr := fmt.Sprintf("%x", receiverKey)
	if s.data[receiverKeyStr] == nil {
		s.data[receiverKeyStr] = make(map[string]*pb.DMSMessage)
	}
	s.data[receiverKeyStr][messageHash(msg)] = msg
	return nil
}

func (s *MemStore) GetNext(receiverKey []byte) (*pb.DMSMessage, error) {
	receiverKeyStr := fmt.Sprintf("%x", receiverKey)
	if s.data[receiverKeyStr] == nil {
		return nil, nil
	}
	if len(s.data[receiverKeyStr]) == 0 {
		return nil, nil
	}
	for _, msg := range s.data[receiverKeyStr] {
		return msg, nil
	}
	return nil, nil
}

func (s *MemStore) GetAll(receiverKey []byte) ([]*pb.DMSMessage, error) {
	receiverKeyStr := fmt.Sprintf("%x", receiverKey)
	if s.data[receiverKeyStr] == nil {
		return nil, nil
	}
	if len(s.data[receiverKeyStr]) == 0 {
		return nil, nil
	}
	var ret []*pb.DMSMessage
	for _, msg := range s.data[receiverKeyStr] {
		ret = append(ret, msg)
	}
	return ret, nil
}

func (s *MemStore) Remove(msg *pb.DMSMessage, receiverKey []byte) error {
	receiverKeyStr := fmt.Sprintf("%x", receiverKey)
	if s.data[receiverKeyStr] == nil {
		return nil
	}
	if len(s.data[receiverKeyStr]) == 0 {
		return nil
	}
	delete(s.data[receiverKeyStr], messageHash(msg))
	return nil
}

func messageHash(msg *pb.DMSMessage) string {
	b, err := proto.Marshal(msg)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256(b))
}
