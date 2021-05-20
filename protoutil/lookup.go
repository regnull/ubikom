package protoutil

import (
	"context"
	"fmt"

	"github.com/regnull/ubikom/pb"
)

func CheckNameAvailability(ctx context.Context, client pb.LookupServiceClient, name string) (bool, error) {
	resp, err := client.LookupName(ctx, &pb.LookupNameRequest{Name: name})
	if err != nil {
		return false, err
	}
	if resp.GetResult().GetResult() == pb.ResultCode_RC_RECORD_NOT_FOUND {
		return true, nil
	}
	if resp.GetResult().GetResult() != pb.ResultCode_RC_OK {
		return false, fmt.Errorf("server returned %s", resp.GetResult().GetResult().String())
	}
	return false, nil
}
