package imap

import (
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
)

type Backend struct {
	dumpClient   pb.DMSDumpServiceClient
	lookupClient pb.LookupServiceClient
	privateKey   *easyecc.PrivateKey
}

func (b *Backend) Login(_ *imap.ConnInfo, user, pass string) (backend.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func NewBackend(dumpClient pb.DMSDumpServiceClient, lookupClient pb.LookupServiceClient,
	privateKey *easyecc.PrivateKey, user, password string) *Backend {
	return &Backend{}
}
