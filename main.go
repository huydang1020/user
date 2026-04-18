package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/huyshop/user/db"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

type Configs struct {
	GRPCPort              string
	DBPath                string
	DBName                string
	HttpPort              string
	JwtSecretKey          string
	JwtExpireAccessToken  string
	JwtExpireRefreshToken string
	RedisAddr             string
	RedisPassword         string
	RedisDb               string
	MailKey               string
	MailUrl               string
	MaxUserAddress        string
}

var config *Configs

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: Error loading .env file:", err)
		} else {
			log.Println("Loaded .env file for local development")
		}
	} else {
		log.Println("No .env file found, using system environment variables")
	}

	config = &Configs{
		GRPCPort:              getEnv("GRPC_PORT", "6000"),
		DBPath:                getEnv("DB_PATH", "root:password@tcp(localhost:3306)"),
		DBName:                getEnv("DB_NAME", "user"),
		HttpPort:              getEnv("HTTP_PORT", "6001"),
		JwtSecretKey:          getEnv("JWT_SECRET_KEY", ""),
		JwtExpireAccessToken:  getEnv("JWT_EXPIRE_ACCESS_TOKEN", "30"),
		JwtExpireRefreshToken: getEnv("JWT_EXPIRE_REFRESH_TOKEN", "129600"),
		RedisAddr:             getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:         getEnv("REDIS_PASSWORD", ""),
		RedisDb:               getEnv("REDIS_DB", "0"),
		MailKey:               getEnv("BREVO_API_KEY", ""),
		MailUrl:               getEnv("BREVO_URL", ""),
		MaxUserAddress:        getEnv("MAX_USER_ADDRESS", "5"),
	}
}

// Helper function để lấy env với default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func startApp(ctx *cli.Context) error {
	log.Printf("Starting user service with config:")
	log.Printf("  GRPC Port: %s", config.GRPCPort)
	log.Printf("  DB Path: %s", config.DBPath)
	log.Printf("  DB Name: %s", config.DBName)
	log.Printf("  Redis Addr: %s", config.RedisAddr)
	u, err := NewUser(config)
	if err != nil {
		log.Fatal(err)
		return err
	}

	go func() {
		for {
			select {
			case user := <-u.SendEmail:
				err := u.CheckAndSendEmailVerifyOtp(user)
				if err != nil {
					log.Println("Error sending email:", err)
				}
			}

		}
	}()
	go func() {
		if err := HTTPServe(u); err != nil {
			log.Print(err)
		}
	}()
	if err := startGRPCServe(config.GRPCPort, u); err != nil {
		debug.PrintStack()
		return err
	}
	return nil
}

func createTableDb(ctx *cli.Context) error {
	d := &db.DB{}
	if err := d.ConnectDb(config.DBPath, config.DBName); err != nil {
		debug.PrintStack()
		return err
	}
	if err := d.CreateDb(); err != nil {
		return err
	}
	log.Print("Tables created")
	return nil
}

func appRoot() error {
	app := cli.NewApp()

	app.Action = func(c *cli.Context) error {
		return errors.New("Wow, ^.^ dumb")
	}

	app.Commands = []*cli.Command{
		{Name: "start", Action: startApp},
		{Name: "createDb", Action: createTableDb},
	}

	return app.Run(os.Args)
}
func main() {
	go freeMemory()
	if err := appRoot(); err != nil {
		panic(err)
	}
}

func freeMemory() {
	for {
		fmt.Println("run gc")
		start := time.Now()
		runtime.GC()
		debug.FreeOSMemory()
		elapsed := time.Since(start)
		fmt.Printf("gc took %s\n", elapsed)
		time.Sleep(15 * time.Minute)
	}
}
