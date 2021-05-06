package store

import "github.com/regnull/ubikom/pb"

type Store interface {
	Save(msg *pb.DMSMessage, receiverKey []byte) error
	GetNext(receiver []byte) (*pb.DMSMessage, error)
	Remove(msg *pb.DMSMessage, receiverKey []byte) error
}
