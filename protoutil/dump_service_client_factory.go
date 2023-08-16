package protoutil

import (
	"context"
	"time"

	"github.com/regnull/ubikom/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultTimeout = time.Second * 5
)

// DumpServiceClientFactory creates DMSDumpService client, which allows sending and receiving of messages.
type DumpServiceClientFactory interface {
	// CreateDumpServiceClient returns a new DMSDumpServiceClient to interact with the server
	// at the given endpoint. The operation times out after timeout duration. If timeout is zero,
	// the default timeout will be used.
	// The function returns client, cleanup function, and error.
	// If the cleanup function is not nil, it must be closed after the caller is done
	// with the client to free up the connection.
	CreateDumpServiceClient(ctx context.Context, url string,
		timeout time.Duration) (pb.DMSDumpServiceClient, func(), error)
}

type dumpServiceClientFactoryImpl struct{}

// NewDumpServiceClientFactory creates and returns a new DumpServiceClientFactory.
func NewDumpServiceClientFactory() DumpServiceClientFactory {
	return &dumpServiceClientFactoryImpl{}
}

func (f *dumpServiceClientFactoryImpl) CreateDumpServiceClient(ctx context.Context,
	url string, timeout time.Duration) (
	pb.DMSDumpServiceClient, func(), error) {
	if timeout == 0 {
		timeout = defaultTimeout
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctxWithTimeout, url, opts...)
	if err != nil {
		return nil, nil, err
	}
	client := pb.NewDMSDumpServiceClient(conn)
	return client, func() { conn.Close() }, nil
}
