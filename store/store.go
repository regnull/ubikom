package store

import "teralyt.com/ubikom/pb"

type Store interface {
	Save(msg *pb.DMSMessage) error
	GetNext(receiver []byte) (*pb.DMSMessage, error)
	Remove(msg *pb.DMSMessage) error
}
