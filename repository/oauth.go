package repository

import (
	"context"
	"fmt"

	"isp-admin-service/entity"

	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
)

//nolint:gosec
const (
	getTokenEndpoint = "/oauth/te"
	getUserEndpoint  = "/oauth/me"
)

type Oauth struct {
	cli *httpcli.Client
}

func NewOauth(cli *httpcli.Client) Oauth {
	return Oauth{
		cli: cli,
	}
}

func (s Oauth) GetToken(ctx context.Context, clientSetting entity.ClientSetting) (*entity.OauthTokenResponse, error) {
	ctx = http_metrics.ClientEndpointToContext(ctx, getTokenEndpoint)

	response := entity.OauthTokenResponse{}
	err := s.cli.Post(getTokenEndpoint).
		BasicAuth(httpcli.BasicAuth{
			Username: clientSetting.ClientId,
			Password: clientSetting.ClientSecret,
		}).
		Header("Content-Type", "application/x-www-form-urlencoded").
		QueryParams(map[string]any{
			"grant_type":   "authorization_code",
			"code":         clientSetting.AuthCode,
			"redirect_uri": clientSetting.RedirectURI,
		}).
		JsonResponseBody(&response).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "call POST: '%s'", getTokenEndpoint)
	}

	return &response, nil
}

func (s Oauth) GetUser(ctx context.Context, accessToken string) (*entity.OauthUserResponse, error) {
	ctx = http_metrics.ClientEndpointToContext(ctx, getUserEndpoint)

	response := entity.OauthUserResponse{}
	err := s.cli.Get(getUserEndpoint).
		Header("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
		JsonResponseBody(&response).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "call GET: '%s'", getUserEndpoint)
	}

	return &response, nil
}
