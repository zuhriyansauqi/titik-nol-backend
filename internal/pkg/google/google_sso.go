package google

import (
	"context"

	"google.golang.org/api/idtoken"
)

type GoogleSSOService interface {
	VerifyIDToken(ctx context.Context, idToken string) (*idtoken.Payload, error)
}

type googleSSOService struct {
	clientID string
}

func NewGoogleSSOService(clientID string) GoogleSSOService {
	return &googleSSOService{
		clientID: clientID,
	}
}

func (s *googleSSOService) VerifyIDToken(ctx context.Context, idToken string) (*idtoken.Payload, error) {
	payload, err := idtoken.Validate(ctx, idToken, s.clientID)
	if err != nil {
		return nil, err
	}
	return payload, nil
}
