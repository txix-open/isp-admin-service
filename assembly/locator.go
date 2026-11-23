package assembly

import (
	"context"
	"time"

	"isp-admin-service/conf"
	"isp-admin-service/controller"
	"isp-admin-service/repository"
	"isp-admin-service/routes"
	"isp-admin-service/service"
	"isp-admin-service/service/delete_old_audit_worker"
	"isp-admin-service/service/inactive_worker"
	"isp-admin-service/service/secure"
	"isp-admin-service/service/session_worker"
	"isp-admin-service/transaction"

	"github.com/txix-open/isp-kit/bgjobx"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/router"
	"github.com/txix-open/isp-kit/log"
)

type DB interface {
	db.DB
	db.Transactional
}

type Locator struct {
	logger   log.Logger
	oauthCli *httpcli.Client
	db       DB
}

func NewLocator(logger log.Logger, oauthCli *httpcli.Client, db DB) Locator {
	return Locator{
		logger:   logger,
		oauthCli: oauthCli,
		db:       db,
	}
}

type Config struct {
	Handler  *router.Router
	BgJobCfg []bgjobx.WorkerConfig
}

//nolint:funlen
func (l Locator) Config(
	ctx context.Context,
	cfg conf.Remote,
	jobPollInterval time.Duration,
) (Config, error) {
	oauthRepo := repository.NewOauth(l.oauthCli)
	roleRepo := repository.NewRole(l.db)
	userRepo := repository.NewUser(l.db)
	tokenRepo := repository.NewToken(l.db)
	auditRepo := repository.NewAudit(l.db)
	auditEventRepo := repository.NewAuditEvent(l.db)
	userRoleRepo := repository.NewUserRole(l.db)

	auditService := service.NewAudit(ctx, l.logger, auditRepo, auditEventRepo, cfg.Audit.EventSettings)
	tokenService := service.NewToken(tokenRepo, cfg.ExpireSec)

	oauthBaseUrl, oauthSettings := l.getOauthSettings(cfg.Oauth)
	oauthService, err := service.NewOauth(oauthSettings, oauthRepo)
	if err != nil {
		return Config{}, errors.WithMessage(err, "new oauth")
	}

	secureService := secure.NewService(tokenRepo, userRoleRepo)

	txManager := transaction.NewManager(l.db)

	userService := service.NewUser(
		userRepo,
		userRoleRepo,
		roleRepo,
		tokenRepo,
		auditService,
		txManager,
		tokenService,
		cfg.IdleTimeoutMs,
		l.logger,
	)
	authService := service.NewAuth(
		userRepo, txManager, tokenService, oauthService, auditService, l.logger,
		cfg.AntiBruteforce.DelayLoginRequestInSec,
		cfg.AntiBruteforce.MaxInFlightLoginRequests,
	)
	roleService := service.NewRole(roleRepo, auditService)

	permissionsService := service.NewPermission(cfg.Permissions)

	uiSettingsService := service.NewUiSettings(oauthBaseUrl, oauthSettings)

	downloadService := service.NewDownload(
		auditService,
		userService,
		tokenService,
		roleService,
		permissionsService,
	)

	userController := controller.NewUser(userService)
	customizationController := controller.NewCustomization(cfg.UiDesign)
	authController := controller.NewAuth(authService, l.logger)
	secureController := controller.NewSecure(secureService)
	sessionController := controller.NewSession(tokenService)
	auditController := controller.NewAudit(auditService)
	roleController := controller.NewRole(roleService)
	permissionController := controller.NewPermissions(permissionsService)
	uiSettingsController := controller.NewUiSettings(uiSettingsService)
	downloadController := controller.NewDownload(downloadService)

	wrapper := endpoint.DefaultWrapper(l.logger, httplog.Log(l.logger, true))
	handler := routes.Handler(
		wrapper,
		routes.Controllers{
			User:          userController,
			Customization: customizationController,
			Auth:          authController,
			Secure:        secureController,
			Session:       sessionController,
			Audit:         auditController,
			Role:          roleController,
			Permissions:   permissionController,
			UiSettings:    uiSettingsController,
			Download:      downloadController,
		},
	)

	inactiveBlocker := inactive_worker.NewInactiveBlocker(
		userRepo,
		auditService,
		userRoleRepo,
		cfg.BlockInactiveWorker,
		l.logger,
	)
	deleteOldAuditWorker := delete_old_audit_worker.NewService(l.logger, auditRepo, cfg.Audit.AuditTTl)
	expireSessionWorker := session_worker.NewExpireSessionWorker(l.logger, txManager)

	return Config{
		Handler: handler,
		BgJobCfg: []bgjobx.WorkerConfig{{
			Queue:        delete_old_audit_worker.QueueName,
			Concurrency:  1,
			PollInterval: jobPollInterval,
			Handle:       deleteOldAuditWorker,
		}, {
			Queue:        inactive_worker.QueueName,
			Concurrency:  1,
			PollInterval: jobPollInterval,
			Handle:       inactiveBlocker,
		}, {
			Queue:        session_worker.QueueName,
			Concurrency:  1,
			PollInterval: jobPollInterval,
			Handle:       expireSessionWorker,
		}},
	}, nil
}

func (l Locator) getOauthSettings(oauthAuth *conf.Oauth) (string, []conf.OauthClientConfig) {
	if oauthAuth == nil {
		return "", []conf.OauthClientConfig{}
	}

	return oauthAuth.Host, oauthAuth.Clients
}
