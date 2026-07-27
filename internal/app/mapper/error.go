package mapper

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Lagwick/catalog-service/internal/app/entity"
)

func ErrorToGRPC(err error) error {
	var appErr *entity.AppError

	if errors.As(err, &appErr) {
		return status.Error(appErr.GRPCCode(), appErr.Message)
	}
	return status.Error(codes.Internal, "internal error")
}
