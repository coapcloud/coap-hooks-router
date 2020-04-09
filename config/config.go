package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	DBFilename  = os.Getenv("DB_FILENAME")
	AdminBearer = []byte(os.Getenv("ADMIN_BEARER"))
)
