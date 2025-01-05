package config

import (
	"log"
	"os"
)

type Config struct {
	DBHost           string `env:"DBHost"`
	DBName           string `env:"DBName"`
	DBUser           string `env:"DBUser"`
	DBPort           string `env:"DBPort"`
	DBPassword       string `env:"DBPassword"`
	Env              string `env:"Env"`
	APISecret        string `env:"APISecret"`
	APIRefreshSecret string `env:"APIRefreshSecret"`
}

func LoadConfig() Config {

	config := Config{
		DBHost:           os.Getenv("DBHost"),
		DBName:           os.Getenv("DBName"),
		DBUser:           os.Getenv("DBUser"),
		DBPort:           os.Getenv("DBPort"),
		DBPassword:       os.Getenv("DBPassword"),
		Env:              os.Getenv("Env"),
		APISecret:        os.Getenv("APISecret"),
		APIRefreshSecret: os.Getenv("APIRefreshSecret"),
	}

	// Vérification des variables d'environnement obligatoires
	requiredEnvVars := map[string]string{
		//"DBHost": config.DBHost,
		"APISecret":        config.APISecret,
		"APIRefreshSecret": config.APIRefreshSecret,
	}

	for varName, varValue := range requiredEnvVars {
		if varValue == "" {
			log.Fatalf("Missing environnement variable %s", varName)
		}
	}

	return config
}
