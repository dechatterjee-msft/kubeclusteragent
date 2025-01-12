package error

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GenerateHTTPStatusError(code codes.Code, msg string) error {
	return status.Errorf(code, msg)
}
