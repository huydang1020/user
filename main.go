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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading env:", err)
	}
	config = &Configs{
		GRPCPort:              os.Getenv("GRPC_PORT"),
		DBPath:                os.Getenv("DB_PATH"),
		DBName:                os.Getenv("DB_NAME"),
		HttpPort:              os.Getenv("HTTP_PORT"),
		JwtSecretKey:          os.Getenv("JWT_SECRET_KEY"),
		JwtExpireAccessToken:  os.Getenv("JWT_EXPIRE_ACCESS_TOKEN"),
		JwtExpireRefreshToken: os.Getenv("JWT_EXPIRE_REFRESH_TOKEN"),
		RedisAddr:             os.Getenv("REDIS_ADDR"),
		RedisPassword:         os.Getenv("REDIS_PASSWORD"),
		RedisDb:               os.Getenv("REDIS_DB"),
		MailKey:               os.Getenv("BREVO_API_KEY"),
		MailUrl:               os.Getenv("BREVO_URL"),
		MaxUserAddress:        os.Getenv("MAX_USER_ADDRESS"),
	}
}

func startApp(ctx *cli.Context) error {
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
