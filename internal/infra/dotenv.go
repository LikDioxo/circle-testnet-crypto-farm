package infra

import (
	"fmt"
	"os"
	"path"

	"github.com/joho/godotenv"
)

func LoadEnv() (error) {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("can't get current working directory")
	}

	err = godotenv.Load(path.Join(wd, "/.env"))
	if err != nil {
		return fmt.Errorf("error loading .env file")
	}

	return nil
}