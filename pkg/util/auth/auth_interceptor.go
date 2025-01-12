package auth

import (
	"context"
	"kubeclusteragent/pkg/util/log/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	jwtManager  *JWTManager
	accessRules map[string][]string
}

func CreateAuthInterceptor(jwtManager *JWTManager) *AuthInterceptor {
	accessRules := getAccessRules()
	return &AuthInterceptor{jwtManager, accessRules}
}

func getAccessRules() map[string][]string {
	rules := make(map[string][]string)
	base := "/agent.v1alpha1.AgentAPI/"
	rules[base+"GetCluster"] = []string{"admin", "view"}
	rules[base+"CreateCluster"] = []string{"admin"}
	rules[base+"UpgradeCluster"] = []string{"admin"}
	rules[base+"PatchCluster"] = []string{"admin"}
	rules[base+"DeleteCluster"] = []string{"admin"}
	rules[base+"GetKubeconfig"] = []string{"admin", "view"}
	rules[base+"ResetKubeconfig"] = []string{"admin"}
	return rules
}

func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		err = interceptor.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (interceptor *AuthInterceptor) authorize(ctx context.Context, method string) error {
	logger := log.From(ctx).WithName("AuthInterceptor")
	allowedRoles, present := interceptor.accessRules[method]
	if !present {
		return nil
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Info("metadata not found in the incoming request")
		return status.Errorf(codes.Unauthenticated, "metadata not found")
	}
	values := md["authorization"]
	if len(values) == 0 {
		logger.Info("access token not found in the incoming request")
		return status.Errorf(codes.Unauthenticated, "access token not found")
	}
	accessToken := values[0]
	claims, err := interceptor.jwtManager.VerifyToken(accessToken)
	if err != nil {
		logger.Error(err, "access token invalid")
		return status.Errorf(codes.Unauthenticated, "access token invalid: %v", err)
	}
	for _, role := range allowedRoles {
		if role == claims.Role {
			return nil
		}
	}
	logger.Info("user does not have privilege to access the API")
	return status.Errorf(codes.PermissionDenied, "user does not have privilege to access the API")
}
