package interceptor

import (
	"context"
	"github.com/itksb/go-url-shortener/internal/user"
	"github.com/itksb/go-url-shortener/pkg/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// CreateUnaryAuthInterceptor - create unary interceptor
func CreateUnaryAuthInterceptor(codec session.Codec) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var userIDValue string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get(string(user.FieldID)); len(values) > 0 {
				err := codec.Decode(string(user.FieldID), values[0], &userIDValue)
				if err != nil {
					return nil, status.Error(codes.Internal, err.Error())
				}
			}
		}

		if userIDValue == "" {
			userIDValue = user.GenerateUserID()
		}

		ctx = context.WithValue(ctx, string(user.FieldID), userIDValue)
		return handler(ctx, req)

	}
}
