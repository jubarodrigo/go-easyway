package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/rs/zerolog/log"
)

type Settings struct {
	CorsOrigins     string        `json:"cors_origins"`
	Host            string        `json:"db_host"`
	HostRead        string        `json:"db_host_read"`
	Port            string        `json:"db_port" default:"3306"`
	Username        string        `json:"db_user" required:"true"`
	Password        string        `json:"db_password"`
	Database        string        `json:"db_database" required:"true"`
	ConnectTimeout  time.Duration `json:"db_connect_timeout" default:"2s"`
	MaxConnLifetime time.Duration `json:"db_max_conn_life_time" default:"1h"`
	MaxConns        int           `json:"db_max_conns" default:"5"`
	MinConns        int           `json:"db_min_conns" default:"5"`
	RefreshPassword bool          `json:"db_refresh_password" default:"false"`
	TZ              *string       `json:"db_time_zone"`
	DBRestore       bool          `json:"db_restore" default:"false"`
	DeviceKeySecret string        `json:"device_key"`
}

func GetParamOr(param, orElse string) string {
	p, f := os.LookupEnv(param)
	if !f {
		return orElse
	}
	return p
}

func GetSettings(mode string, awsCfg aws.Config) *Settings {
	parseJsonSettings := func(secret string) *Settings {
		var config Settings
		err := json.Unmarshal([]byte(secret), &config)
		if err != nil {
			panic(err)
		}

		// verify that all required secrets are present
		requiredSettings := []string{"Database", "CorsOrigins"}

		v := reflect.ValueOf(config)
		for _, field := range requiredSettings {
			f := v.FieldByName(field)
			if !f.IsValid() {
				panic(fmt.Sprintf("Required setting %s is missing", field))
			}
		}

		return &config
	}

	var jsonSettings string
	if mode == "local" {
		// Get secrets from local file
		content, err := parsingEnvToSettingsJsonObj()
		if err != nil {
			panic(err)
		}
		jsonSettings = content
	} else {
		// Create a Secrets Manager client
		secretsClient := secretsmanager.NewFromConfig(awsCfg)

		// Build the request input
		secretName := fmt.Sprintf("%s/template", mode)
		secretInput := &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretName),
		}

		// Get secret value
		log.Info().Msg(fmt.Sprintf("Trying to get secret: %s\n", secretName))
		sv, err := secretsClient.GetSecretValue(context.Background(), secretInput)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("error getting secret: %v", err))
			panic(err)
		}

		if sv.SecretString != nil {
			log.Info().Msg(fmt.Sprintf("Got secret: %s\n", secretName))
			jsonSettings = ReplaceSecretValues(*sv.SecretString)
		} else {
			log.Info().Msg(fmt.Sprintf("could not retrieve secret: %v\n", secretName))
		}
	}

	return parseJsonSettings(jsonSettings)
}

func parsingEnvToSettingsJsonObj() (string, error) {
	envData := make(map[string]string)
	for _, k := range os.Environ() {
		parts := strings.SplitN(k, "=", 2)
		if len(parts) == 2 {
			envData[parts[0]] = parts[1]
		}
	}

	jsonData, err := json.Marshal(envData)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func ReplaceSecretValues(secretSettings string) string {
	content, err := parsingEnvToSettingsJsonObj()
	if err != nil {
		panic(err)
	}
	var envData, secretData map[string]interface{}
	errEnv := json.Unmarshal([]byte(content), &envData)
	errSec := json.Unmarshal([]byte(secretSettings), &secretData)
	if errEnv != nil || errSec != nil {
		panic(fmt.Sprintf("Error while parsing environment variables file: %s, %s", errEnv, errSec))
	} else {
		envData["db_user"] = secretData["db_user"]
		envData["db_password"] = secretData["db_password"]
		envData["device_key"] = secretData["device_key"]
		jsonData, err := json.Marshal(envData)
		if err != nil {
			panic(err)
		}
		return string(jsonData)
	}
}
