package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Environment       string `mapstructure:"ENVIRONMENT"`
	HttpHost          string `mapstructure:"HTTP_HOST"`
	HttpPort          string `mapstructure:"HTTP_PORT"`
	GrpcAdminHost     string `mapstructure:"GRPC_ADMIN_HOST"`
	GrpcAdminPort     string `mapstructure:"GRPC_ADMIN_PORT"`
	GrpcImplantHost     string `mapstructure:"GRPC_IMPLANT_HOST"`
	GrpcImplantPort     string `mapstructure:"GRPC_IMPLANT_PORT"`
	ApiKey            string `mapstructure:"API_KEY"`
	NistBaseUrl       string `mapstructure:"NIST_BASE_URL"`
	KeywordSearchPath string `mapstructure:"KEYWORD_SEARCH_PATH"`
	//DBUsername    string `mapstructure:"DB_USERNAME"`
	//DBPassword    string `mapstructure:"DB_PASSWORD"`
	//DBHost        string `mapstructure:"DB_HOSTNAME"`
	//DBPort        string `mapstructure:"DB_PORT"`
	//DBName        string `mapstructure:"DB_DBNAME"`
	//DBNameTest    string `mapstructure:"DB_DBNAME_TEST"`
	//MigrationPath string `mapstructure:"MIGRATION_PATH"`
	//DBRecreate    bool   `mapstructure:"DB_RECREATE"`
	//DBUrl         string
}

func LoadConfig(name string, path string) (config Config) {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("config: %v", err)
		return
	}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("config: %v", err)
		return
	}

	return
}
