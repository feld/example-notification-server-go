package delivery

import (
	"context"
	"errors"
	"io/ioutil"
	"time"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"github.com/xmtp/example-notification-server-go/pkg/options"
	"go.uber.org/zap"
)

type ApnsDelivery struct {
	logger            *zap.Logger
	notificationTopic string
	apnsClient        *apns2.Client
	opts              options.ApnsOptions
}

func NewApnsDelivery(logger *zap.Logger, opts options.ApnsOptions) (*ApnsDelivery, error) {
	var bytes []byte
	var err error

	if opts.P8Certificate == "" {
		bytes, err = ioutil.ReadFile(opts.P8CertificateFilePath)

		if err != nil {
			return nil, err
		}
	} else {
		bytes = []byte(opts.P8Certificate)
	}

	client, err := getApnsClient(bytes, opts.KeyId, opts.TeamId)

	if opts.Mode == "production" {
		client.Production()
	} else if opts.Mode == "development" {
		client.Development()
	} else {
		return nil, errors.New("invalid APNS mode")
	}

	if err != nil {
		return nil, err
	}

	return &ApnsDelivery{
		logger:     logger.Named("delivery-service"),
		apnsClient: client,
		opts:       opts,
	}, nil
}

func (a ApnsDelivery) Send(ctx context.Context, deviceToken, topic, message string) error {
	// TODO: Figure out the message format
	notificationPayload := payload.NewPayload().
		MutableContent().
		Alert("New message from XMTP").
		Custom("topic", topic).
		Custom("encryptedMessage", message)

	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       a.opts.Topic,
		Payload:     notificationPayload,
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	res, err := a.apnsClient.PushWithContext(ctx, notification)
	if res != nil {
		a.logger.Info(
			"Sent notification",
			zap.String("apns_id", res.ApnsID),
			zap.Int("status_code", res.StatusCode),
			zap.String("reason", res.Reason),
		)
	}

	return err
}

func getApnsClient(authKey []byte, keyId, teamId string) (*apns2.Client, error) {
	key, err := token.AuthKeyFromBytes(authKey)
	if err != nil {
		return nil, err
	}

	authToken := &token.Token{
		AuthKey: key,
		KeyID:   keyId,
		TeamID:  teamId,
	}

	return apns2.NewTokenClient(authToken), nil
}
