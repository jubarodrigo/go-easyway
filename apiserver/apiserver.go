package apiserver

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type ApiServer struct {
	awsCfg *aws.Config
}

func NewServer(awsCfg *aws.Config) *ApiServer {
	return &ApiServer{
		awsCfg: awsCfg,
	}
}

const (
	SessionHeaderName   = "x-session"
	CSRFTokenHeaderName = "x-csrf-token"
)

func (a *ApiServer) SetupRoutes(envBaseUrl string, r *chi.Mux, port int, settings_cors_origins string) {
	a.setupCORS(settings_cors_origins, r)
	a.setupMiddleware(r)

	envBaseUrl = fmt.Sprintf("/%s", envBaseUrl)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	fs := http.FileServer(http.Dir("web"))
	r.Get("/web/*", http.StripPrefix("/web", fs).ServeHTTP)

	a.registerCommonAPI(envBaseUrl, r)

	serveSwagger(r)
}

func (a *ApiServer) setupCORS(settings_cors_origins string, r *chi.Mux) {
	cors_origins, err := LoadAllowedOrigins(settings_cors_origins)
	if err != nil {
		panic(err)
	}
	if len(cors_origins) > 0 {
		log.Info().Msg(fmt.Sprintf("Cors Host: %s\n", cors_origins))
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   cors_origins, // set allowed origins to all
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", CSRFTokenHeaderName, SessionHeaderName},
			ExposedHeaders:   []string{"Link", CSRFTokenHeaderName, SessionHeaderName},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}))
	}
}

// TODO - Move to API Server
func (a *ApiServer) setupMiddleware(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))
}

func handleHealthCheck(response http.ResponseWriter, request *http.Request) {
	type serverTime struct {
		Message string `json:"message"`
		Time    string `json:"time"`
	}
	now := time.Now()
	data := &serverTime{
		Time:    now.Format(time.RFC3339),
		Message: "Systems Up",
	}
	render.JSON(response, request, data)
}

func (a *ApiServer) registerCommonAPI(envBaseUrl string, subrouter chi.Router) {
	subrouter.Group(func(r chi.Router) {
		r.Get(envBaseUrl+"/health", handleHealthCheck)
	})
}

func LoadAllowedOrigins(settings_origins string) ([]string, error) {
	origins := strings.Split(settings_origins, ",")
	err := ValidateAllowedOrigins(origins)
	if err != nil {
		panic(fmt.Sprintf("Invalid allowed origins: %v", err))
	}
	return origins, nil
}

func ValidateAllowedOrigins(origins []string) error {
	for _, origin := range origins {
		if origin == "*" {
			return errors.New("wildcard '*' is not allowed in CORS allowed origins")
		}
	}
	return nil
}

func getSpecs() map[string]string {
	return map[string]string{
		"App": "docs/swagger.yaml",
	}
}

func serveSwagger(router chi.Router) {
	log.Info().Msg("serving swagger at /swagger/index.html")

	// serve the docs folder at /swagger/docs/
	fs := http.FileServer(http.Dir("docs"))
	router.Handle("/swagger/docs/*", http.StripPrefix("/swagger/docs/", fs))

	// server swagger ui with each swagger spec in getSpecs() configured
	specUrls := ""
	for specName, specUrl := range getSpecs() {
		specUrls += fmt.Sprintf(`{name: "%s", url: "%s"},`, specName, specUrl)
	}
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.UIConfig(map[string]string{
			"urls": fmt.Sprintf("[%s]", specUrls),
		}),
	))
}
