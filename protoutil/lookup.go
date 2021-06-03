package protoutil

import (
	"context"

	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"google.golang.org/grpc/codes"
)

func CheckNameAvailability(ctx context.Context, client pb.LookupServiceClient, name string) (bool, error) {
	_, err := client.LookupName(ctx, &pb.LookupNameRequest{Name: name})
	if err != nil && util.StatusCodeFromError(err) != codes.NotFound {
		return false, err
	}

	if err != nil && util.StatusCodeFromError(err) == codes.NotFound {
		return true, nil
	}
	return false, nil
}
