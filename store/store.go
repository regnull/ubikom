package store

import "github.com/regnull/ubikom/pb"

// Store represents local store for DMSMessages.
type Store interface {
	// Save saves a new message using the receiver key.
	Save(msg *pb.DMSMessage, receiverKey []byte) error

	// Get next returns next message available for this receiver.
	GetNext(receiver []byte) (*pb.DMSMessage, error)

	// Get all returns all messages available for this receiver.
	GetAll(receiver []byte) ([]*pb.DMSMessage, error)

	// Remove removes the message from the local storage.
	Remove(msg *pb.DMSMessage, receiverKey []byte) error
}
