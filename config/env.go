package config

import (
	"fileblobs/utils"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		utils.LogIfDevelopment("⚠️ Arquivo .env não encontrado, usando valores padrão")
	}
}
