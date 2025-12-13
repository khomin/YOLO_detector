package bootstrap

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Env struct {
	APP_ENV           string `mapstructure:"APP_ENV"`
	SERVER_PUBLIC_URL string `mapstructure:"SERVER_PUBLIC_URL"`
	// db
	DB_HOST                       string `mapstructure:"DB_HOST"`
	DB_NAME                       string `mapstructure:"DB_NAME"`
	CLIENT_COLLECTION             string `mapstructure:"CLIENT_COLLECTION"`
	VIEWS_COLLECTION              string `mapstructure:"VIEWS_COLLECTION"`
	SESSION_COLLECTION            string `mapstructure:"SESSION_COLLECTION"`
	SESSION_SUBSCRIBER_COLLECTION string `mapstructure:"SESSION_SUBSCRIBER_COLLECTION"`
	SESSION_DEVICE_COLLECTION     string `mapstructure:"SESSION_DEVICE_COLLECTION"`
	// auth
	AUTH0_DOMAIN                     string   `mapstructure:"AUTH0_DOMAIN"`
	AUTH0_DOMAIN_SHEME               string   `mapstructure:"AUTH0_DOMAIN_SHEME"`
	AUTH0_CLIENT_ID                  string   `mapstructure:"AUTH0_CLIENT_ID"`
	AUTH0_WEBHOOK_SECRET             string   `mapstructure:"AUTH0_WEBHOOK_SECRET"`
	AUTH0_WEBHOOK_PROXY              []string `mapstructure:"AUTH0_WEBHOOK_PROXY"`
	AUTH0_CLIENT_SECRET              string   `mapstructure:"AUTH0_CLIENT_SECRET"`
	AUTH0_CALLBACK_URL               string   `mapstructure:"AUTH0_CALLBACK_URL"`
	AUTH0_CACHE_EXPIRATION_MIN       int      `mapstructure:"AUTH0_CACHE_EXPIRATION_MIN"`
	AUTH0_CACHE_CLEANUP_INTERVAL_MIN int      `mapstructure:"AUTH0_CACHE_CLEANUP_INTERVAL_MIN"`
	AUTH0_USE_EMAIL_VERIFICATION     bool     `mapstructure:"AUTH0_USE_EMAIL_VERIFICATION"`
	// uuid
	UUID_TOKEN_SECRET     string `mapstructure:"UUID_TOKEN_SECRET"`
	UUID_TOKEN_EXPIRY_SEC int64  `mapstructure:"UUID_TOKEN_EXPIRY_SEC"`
	// billing
	STRIPE_DOMAIN                       string   `mapstructure:"STRIPE_DOMAIN"`
	STRIPE_PUBLISHABLE_KEY              string   `mapstructure:"STRIPE_PUBLISHABLE_KEY"`
	STRIPE_SECRET_KEY                   string   `mapstructure:"STRIPE_SECRET_KEY"`
	STRIPE_WEBHOOK_PROXY                []string `mapstructure:"STRIPE_WEBHOOK_PROXY"`
	CUSTOMER_CACHE_EXPIRATION_MIN       int      `mapstructure:"CUSTOMER_CACHE_EXPIRATION_MIN"`
	CUSTOMER_CACHE_CLEANUP_INTERVAL_MIN int      `mapstructure:"CUSTOMER_CACHE_CLEANUP_INTERVAL_MIN"`
	STRIPE_WEBHOOK_ENDPOINT_SECRET      string   `mapstructure:"STRIPE_WEBHOOK_ENDPOINT_SECRET"`
	BILL_CHECKOUT_SESSION_TIMEOUT_SEC   int64    `mapstructure:"BILL_CHECKOUT_SESSION_TIMEOUT_SEC"`
	// timeout
	HANDLE_ACTIVE_CLIENT_MS         int `mapstructure:"HANDLE_ACTIVE_CLIENT_MS"`
	CLIENT_TIMEOUT_MS               int `mapstructure:"CLIENT_TIMEOUT_MS"`
	CLIENT_RETRY_NOTIFY_MS          int `mapstructure:"CLIENT_RETRY_NOTIFY_MS"`
	CONTEXT_TIMEOUT                 int `mapstructure:"CONTEXT_TIMEOUT"`
	STATS_HANDLE_INTERVAL_SEC       int `mapstructure:"STATS_HANDLE_INTERVAL_SEC"`
	STATS_RESET_ERRORS_INTERVAL_SEC int `mapstructure:"STATS_RESET_ERRORS_INTERVAL_SEC"`
	// log
	LOG_PATH string `mapstructure:"LOG_PATH"`
	// internal-system
	ORCH_PORT                 string `mapstructure:"ORCH_PORT"`
	ORCH_IP                   string `mapstructure:"ORCH_IP"`
	ORCH_BLOB_GET_TIMEOUT_SEC int    `mapstructure:"ORCH_BLOB_GET_TIMEOUT_SEC"`
	// dashboard
	DASHBOARD_ADDR             string `mapstructure:"DASHBOARD_ADDR"`
	DASHBOARD_VPN_ADDR         string `mapstructure:"DASHBOARD_VPN_ADDR"`
	DASHBOARD_API_PORT         string `mapstructure:"DASHBOARD_API_PORT"`
	DASHBOARD_HTML_PORT        string `mapstructure:"DASHBOARD_HTML_PORT"`
	DASHBOARD_HASH             string `mapstructure:"DASHBOARD_HASH"`
	DASHBOARD_TOKEN_SECRET     string `mapstructure:"DASHBOARD_TOKEN_SECRET"`
	DASHBOARD_TOKEN_EXPIRY_SEC int64  `mapstructure:"DASHBOARD_TOKEN_EXPIRY_SEC"`
	DASHBOARD_GET_CLIENT_MAX   int    `mapstructure:"DASHBOARD_GET_CLIENT_MAX"`
	DASHBOARD_PATH             string `mapstructure:"DASHBOARD_PATH"`
	INSTALL_SCRIPT_ARM         string `mapstructure:"INSTALL_SCRIPT_ARM"`
	INSTALL_SCRIPT_INTELL      string `mapstructure:"INSTALL_SCRIPT_INTELL"`
	INSTALL_IMAGE_ARM          string `mapstructure:"INSTALL_IMAGE_ARM"`
	INSTALL_IMAGE_INTELL       string `mapstructure:"INSTALL_IMAGE_INTELL"`
	// web server
	WEB_SERVER_RESOURCE_PATH string `mapstructure:"WEB_SERVER_RESOURCE_PATH"`
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
