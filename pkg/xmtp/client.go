package xmtp

import (
	"context"
	"crypto/tls"
	"fmt"

	v1 "github.com/xmtp/proto/go/message_api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	clientVersionMetadataKey = "x-client-version"
	appVersionMetadataKey    = "x-app-version"
)

func newConn(ctx context.Context, apiAddress string, useTls bool) (*grpc.ClientConn, error) {
	dialAddr := fmt.Sprintf(apiAddress)
	return grpc.DialContext(
		ctx,
		dialAddr,
		grpc.WithTransportCredentials(getCredentials(useTls)),
	)
}

func getCredentials(useTls bool) credentials.TransportCredentials {
	if useTls {
		return credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: false,
		})
	}
	return insecure.NewCredentials()
}

func NewClient(ctx context.Context, apiAddress string, useTls bool, clientVersion, appVersion string) (v1.MessageApiClient, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, clientVersionMetadataKey, clientVersion)
	ctx = metadata.AppendToOutgoingContext(ctx, appVersionMetadataKey, appVersion)
	conn, err := newConn(ctx, apiAddress, useTls)
	if err != nil {
		return nil, err
	}

	return v1.NewMessageApiClient(conn), nil
}
