package cloudsearch

import (
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Api struct {
	Env          Env
	Accounts     AccountsStorage
	Auth         AuthBuilder
	OauthService OAuth2Authenticator
}

func (a *Api) Start(port string, done chan error) error {
	gin.SetMode(gin.ReleaseMode)

	s := gin.New()

	s.GET("/oauth/start/:service", a.oauthStart)
	s.GET("/oauth/callback/:service", a.oauthCallback(done))

	logrus.Info("Server starting on ", port)
	return s.Run(port)
}

func (a *Api) oauthStart(ctx *gin.Context) {
	service, err := ParseAccountType(ctx.Param("service"))
	if err != nil {
		renderError(ctx, err)
		return
	}

	url, err := a.OauthService.AuthorizeUrl(service, OauthRedirectUrlFor(a.Env, service))
	if err != nil {
		renderError(ctx, err)
		return
	}

	ctx.Redirect(302, url)
}

func (a *Api) oauthCallback(done chan error) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var err error
		defer func() {
			done <- err
		}()

		service, err := ParseAccountType(ctx.Param("service"))
		code := ctx.Query("code")
		logrus.Debug("Oauth callback for ", service, " - ", code)

		auth, err := a.Auth(service)
		if err != nil {
			renderError(ctx, err)
			return
		}

		aa := AccountType(service)
		acc, err := a.OauthService.AccountFromCode(
			aa,
			code,
			OauthRedirectUrlFor(a.Env, aa),
		)
		if err != nil || acc == nil {
			renderError(ctx, err)
			return
		}

		acc, err = auth.FetchIdentityInfo(*acc)
		if err != nil {
			renderError(ctx, err)
			return
		}

		err = a.Accounts.Save(acc)
		if err != nil {
			renderError(ctx, err)
			return
		}

		logrus.Debug("Account saved: ", acc)
		//	TODO render html?

		page, err := rice.FindBox("static")
		if err != nil {
			logrus.Debug("Data missing: ", err)
			ctx.Status(500)
		}

		ctx.Data(200, "text/html", page.MustBytes("account_linked.html"))
	}
}

func renderError(context *gin.Context, err error) {
	logrus.Debug("Rendering error: ", err)
	context.JSON(
		406,
		map[string]interface{}{
			"error": err.Error(),
		},
	)
}

func OauthRedirectUrlFor(env Env, accountType AccountType) string {
	return fmt.Sprintf("%s%s/oauth/callback/%s",
		env.ServerBase,
		env.HttpPort,
		accountType,
	)
}
