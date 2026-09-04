package service

import (
	"context"
	"isp-admin-service/conf"
	"isp-admin-service/domain"
	"net/url"

	"github.com/txix-open/isp-kit/errors"
)

type UiSettings struct {
	oauthBaseUrl string
	settings     map[string]conf.OauthClientConfig
}

func NewUiSettings(oauthBaseUrl string, settings []conf.OauthClientConfig) UiSettings {
	settingsMap := make(map[string]conf.OauthClientConfig, len(settings))
	for _, setting := range settings {
		if setting.UiSetting == nil {
			continue
		}
		settingsMap[setting.ClientName] = setting
	}
	return UiSettings{
		oauthBaseUrl: oauthBaseUrl,
		settings:     settingsMap,
	}
}

func (s UiSettings) OauthSettings(ctx context.Context, req domain.ClientNameRequest) (*domain.UiOauthSetting, error) {
	setting, ok := s.settings[req.ClientName]
	if !ok {
		return nil, domain.ErrUnknownClient
	}

	loginUrl, err := s.buildUrl(
		setting.UiSetting.LoginEndpoint,
		map[string]string{
			"client_id":    setting.ClientId,
			"redirect_uri": setting.RedirectURI,
		})
	if err != nil {
		return nil, errors.Errorf("build url for login: %w", err)
	}

	logoutUrl, err := s.buildUrl(setting.UiSetting.LogoutEndpoint, nil)
	if err != nil {
		return nil, errors.Errorf("build url for logout: %w", err)
	}
	return &domain.UiOauthSetting{
		LoginUrl:  loginUrl,
		LogoutUrl: logoutUrl,
	}, nil
}

func (s UiSettings) buildUrl(methodPath string, additionalQuery map[string]string) (string, error) {
	method, err := url.Parse(methodPath)
	if err != nil {
		return "", errors.Errorf("parse method path")
	}

	joined, err := url.JoinPath(s.oauthBaseUrl, method.Path)
	if err != nil {
		return "", errors.Errorf("join method path and baseUrl")
	}

	u, err := url.Parse(joined)
	if err != nil {
		return "", errors.Errorf("parse joined")
	}

	q := method.Query()
	for k, v := range additionalQuery {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}
