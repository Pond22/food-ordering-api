package config

import "os"

var (
	DBUser     = os.Getenv("DB_USER")     // ค่าจาก environment variable DB_USER
	DBPassword = os.Getenv("DB_PASSWORD") // ค่าจาก environment variable DB_PASSWORD
	DBName     = os.Getenv("DB_NAME")     // ค่าจาก environment variable DB_NAME
	DBHost     = os.Getenv("DB_HOST")     // ค่าจาก environment variable DB_HOST
	DBPort     = os.Getenv("DB_PORT")     // ค่าจาก environment variable DB_PORT
)
