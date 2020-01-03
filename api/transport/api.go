package transport

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/optimatiq/threatbite/api/controllers"
	"github.com/optimatiq/threatbite/api/transport/middlewares"
	"github.com/optimatiq/threatbite/config"
	"github.com/optimatiq/threatbite/email"
	emailDatasource "github.com/optimatiq/threatbite/email/datasource"
	"github.com/optimatiq/threatbite/ip"
	ipDatasource "github.com/optimatiq/threatbite/ip/datasource"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/acme/autocert"
)

// API state container
type API struct {
	config            *config.Config
	echo              *echo.Echo
	controllerEmail   *controllers.Email
	controllerIP      *controllers.IP
	controllerRequest *controllers.Request
}

// NewAPI returns new HTTP server, which is listening on given port
func NewAPI(config *config.Config) (*API, error) {
	emailData := email.NewEmail(
		config.PwnedKey,
		config.SMTPHello,
		config.SMTPFrom,
		emailDatasource.NewURLDataSource(config.EmailDisposalList),
		emailDatasource.NewURLDataSource(config.EmailFreeList),
	)
	emailData.RunUpdates()

	emailController, err := controllers.NewEmail(emailData)
	if err != nil {
		return nil, err
	}

	ipdata := ip.NewIP(
		config.MaxmindKey,
		ipDatasource.NewURLDataSource(config.ProxyList),
		ipDatasource.NewURLDataSource(config.SpamList),
		ipDatasource.NewURLDataSource(config.VPNList),
		ipDatasource.NewURLDataSource(config.DCList),
	)
	ipdata.RunUpdates()

	ipController, err := controllers.NewIP(ipdata)
	if err != nil {
		return nil, err
	}

	requestController, err := controllers.NewRequest(ipdata)
	if err != nil {
		return nil, err
	}

	return &API{
		config:            config,
		echo:              echo.New(),
		controllerEmail:   emailController,
		controllerIP:      ipController,
		controllerRequest: requestController,
	}, nil
}

// Run starts HTTP server and exposes endpoints
func (a *API) Run() {
	a.echo.HideBanner = true
	a.echo.Server = &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	a.echo.Use(middleware.Logger())
	a.echo.Use(middleware.Recover())
	a.echo.Use(middleware.BodyLimit("2M"))
	a.echo.Use(middleware.RequestID())
	a.echo.Use(middlewares.NewRatelimiter())
	a.echo.Use(middlewares.NewPrometheus())

	// Internal endpoints should be protected from public access
	internal := a.echo.Group("/internal")
	internal.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	internal.GET("/health", a.handleHealth)
	internal.GET("/debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	internal.GET("/debug/pprof/heap", echo.WrapHandler(pprof.Handler("heap")))
	internal.GET("/routes", a.handleRoutes)

	// Public pages
	a.echo.File("/", "resources/static/index.html")
	a.echo.File("/swagger.yml", "resources/static/swagger.yml")

	// Public API endpoints authorization token required
	endpoints := a.echo.Group("/v1/score")
	endpoints.GET("/ip/:ip", a.handleIP)
	endpoints.POST("/request", a.handleRequest)
	endpoints.GET("/email/:email", a.handleEmail)

	if a.config.AutoTLS {
		a.echo.AutoTLSManager.Cache = autocert.DirCache("./resources/tls_cache")
		a.echo.Logger.Fatal(a.echo.StartAutoTLS(fmt.Sprintf(":%d", a.config.Port)))
	}
	a.echo.Logger.Fatal(a.echo.Start(fmt.Sprintf(":%d", a.config.Port)))
}

func (a *API) handleHealth(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (a *API) handleRoutes(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, a.echo.Routes(), " ")
}

func (a *API) handleIP(c echo.Context) error {
	// echo params are not urledecoded automatically
	ip, err := url.QueryUnescape(c.Param("ip"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid IP address")
	}
	if err := a.controllerIP.Validate(ip); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := a.controllerIP.Check(ip)
	if err != nil {
		log.Errorf("ip: %s, error: %s", ip, err)
		return echo.ErrInternalServerError
	}

	return c.JSONPretty(http.StatusOK, result, "  ")
}

func (a *API) handleEmail(c echo.Context) error {
	// echo params are not urledecoded automatically, so query like this lame%40o2.pl will not be valid email.
	email, err := url.QueryUnescape(c.Param("email"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email")
	}
	if err := a.controllerEmail.Validate(email); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := a.controllerEmail.Check(email)
	if err != nil {
		log.Errorf("err: %s, email: %s", err, email)
		return echo.ErrInternalServerError
	}

	return c.JSONPretty(http.StatusOK, result, "  ")
}

func (a *API) handleRequest(c echo.Context) error {
	request := controllers.RequestQuery{}
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := a.controllerRequest.Validate(request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := a.controllerRequest.Check(request)
	if err != nil {
		log.Errorf("err: %s, email: %s", err, request)
		return echo.ErrInternalServerError
	}

	return c.JSONPretty(http.StatusOK, result, "  ")
}
