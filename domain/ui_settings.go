package domain

type ClientNameRequest struct {
	ClientName string `validate:"required"`
}

type UiOauthSetting struct {
	LoginUrl  string // URL для получения authCode для авторизации через OAuth
	LogoutUrl string // URL для завершения сессии
}
