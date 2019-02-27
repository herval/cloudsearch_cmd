package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/herval/cloudsearch"
	"github.com/sirupsen/logrus"
)

type api struct {
	env          cloudsearch.Env
	accounts     cloudsearch.AccountsStorage
	auth         cloudsearch.AuthBuilder
	oauthService cloudsearch.OAuth2Authenticator
}

func (a *api) Start(port string, done chan bool) error {
	gin.SetMode(gin.ReleaseMode)
	s := gin.Default()

	s.GET("/oauth/start/:service", a.oauthStart)
	s.GET("/oauth/callback/:service", a.oauthCallback(done))

	logrus.Debug("Server starting on ", port)
	return s.Run(port)
}

func (a *api) oauthStart(ctx *gin.Context) {
	service, err := cloudsearch.ParseAccountType(ctx.Param("service"))
	if err != nil {
		renderError(ctx, err)
		return
	}

	url, err := a.oauthService.AuthorizeUrl(service, OauthRedirectUrlFor(a.env, service))
	if err != nil {
		renderError(ctx, err)
		return
	}

	ctx.Redirect(302, url)
}

func (a *api) oauthCallback(done chan bool) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		service, err := cloudsearch.ParseAccountType(ctx.Param("service"))
		code := ctx.Query("code")
		logrus.Debug("Oauth callback for ", service, " - ", code)

		auth, err := a.auth(service)
		if err != nil {
			renderError(ctx, err)
			return
		}

		aa := cloudsearch.AccountType(service)
		acc, err := a.oauthService.AccountFromCode(
			aa,
			code,
			OauthRedirectUrlFor(a.env, aa),
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

		err = a.accounts.Save(acc)
		if err != nil {
			renderError(ctx, err)
			return
		}

		logrus.Debug("Account saved: ", acc)
		//	TODO render html?
		ctx.Status(200)
		done <- true
	}
}

func renderError(context *gin.Context, err error) {
	logrus.Error("Rendering error: ", err)
	context.JSON(
		406,
		map[string]interface{}{
			"error": err.Error(),
		},
	)
}

func OauthRedirectUrlFor(env cloudsearch.Env, accountType cloudsearch.AccountType) string {
	return fmt.Sprintf("%s%s/oauth/callback/%s",
		env.ServerBase,
		env.HttpPort,
		accountType,
	)
}