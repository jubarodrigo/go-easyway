package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/docgen"
	"github.com/jessevdk/go-flags"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "github.com/swaggo/http-swagger/example/go-chi/docs"

	"template/apiserver"
	"template/config"
	"template/datastore/db/mysql"
	awsUtils "template/pkg"
)

const (
	localEnv         = "local"
	awsRegionEnv     = "aws_region"
	awsRegionDefault = "us-west-2"
)

type Args struct {
	Address  string `short:"a" long:"address" description:"The address to listen on for HTTP requests" default:"0.0.0.0"`
	Port     int    `short:"p" long:"port" description:"The port to listen on for HTTP requests" default:"3333"`
	Routes   bool   `short:"r" long:"routes" description:"Generate router documentation"`
	Database bool   `short:"d" long:"database" description:"Use a database"`
}

type Starship struct {
	args        Args
	s3Utils     awsUtils.S3Utils
	awsCfg      aws.Config
	settingsMap *config.Settings
	mode        string
	Database    mysql.DB
}

func NewStarship() *Starship {
	return &Starship{}
}

type StarshipBuilder interface {
	setConfig()
	setDatabase()
	setRepositories()
	setServices()
	setWebServer()
}

type BuildDirector struct {
	builder StarshipBuilder
}

func NewStarshipBuilder(sb StarshipBuilder) *BuildDirector {
	return &BuildDirector{
		builder: sb,
	}
}

func (sbd *BuildDirector) BuildStarship() {
	sbd.builder.setConfig()
	sbd.builder.setDatabase()
	sbd.builder.setRepositories()
	sbd.builder.setServices()
	sbd.builder.setWebServer()
}

func (star *Starship) setConfig() {
	var awsCfg aws.Config
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	parser := flags.NewParser(&star.args, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Err(err)
		return
	}

	errDotEnv := godotenv.Load(".env")
	if errDotEnv != nil {
		panic(fmt.Sprintf("Error loading .env file, err: %s", err))
	}

	star.mode = config.GetParamOr("mode", localEnv)
	awsRegion := config.GetParamOr(awsRegionEnv, awsRegionDefault)

	log.Info().Msg(fmt.Sprintf("Starting Service in mode ** %s ** in region ** %s ** port ** %d **\n", star.mode, awsRegion, star.args.Port))

	if star.mode != localEnv {
		awsCfg, err = awsConfig.LoadDefaultConfig(context.Background(), awsConfig.WithRegion(awsRegion))
		if err != nil {
			log.Info().Msg(fmt.Sprintf("error loading AWS config: %v\n", err))
			os.Exit(1)
		}
	}

	star.settingsMap = config.GetSettings(star.mode, awsCfg)
	star.awsCfg = awsCfg
}

func (star *Starship) setWebServer() {
	log.Info().Msg("Starting Web Server...")

	port := fmt.Sprintf(":%d", star.args.Port)
	webServer := apiserver.NewServer(&star.awsCfg)

	r := chi.NewRouter()
	webServer.SetupRoutes(star.mode, r, star.args.Port, star.settingsMap.CorsOrigins)

	if star.args.Routes {
		log.Info().Msg(docgen.JSONRoutesDoc(r))

		return
	}

	if err := http.ListenAndServe(port, r); err != nil {
		panic(err)
	}
}

func (star *Starship) setDatabase() {
	star.Database = mysql.MustSetupDB(context.Background(), star.awsCfg, mysql.DBConfig{Settings: *star.settingsMap})
}

func (star *Starship) setRepositories() {
}

func (star *Starship) setServices() {
}
