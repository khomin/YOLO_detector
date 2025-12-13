package bootstrap

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Env struct {
	APP_ENV           string `mapstructure:"APP_ENV"`
	LOG_PATH          string `mapstructure:"LOG_PATH"`
	REST_IP           string `mapstructure:"REST_IP"`
	REST_PORT         string `mapstructure:"REST_PORT"`
	EVENT_SERVER_IP   string `mapstructure:"EVENT_SERVER_IP"`
	EVENT_SERVER_PORT string `mapstructure:"EVENT_SERVER_PORT"`
	DB_HOST           string `mapstructure:"DB_HOST"`
	DB_NAME           string `mapstructure:"DB_NAME"`
}

func NewEnv(configPath string) *Env {
	env := Env{}
	viper.SetConfigFile(configPath)
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatalf("can't find the file: %s", err.Error())
	}

	err = viper.Unmarshal(&env)
	if err != nil {
		logrus.Fatalf("environment can't be loaded: %s", err.Error())
	}

	if env.APP_ENV == "development" {
		logrus.Info("the App is running in development env")
	}

	logrus.Infof("environment ready: %s", configPath)
	return &env
}
