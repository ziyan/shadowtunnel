package cli

import (
	"io/ioutil"
	"os"
	"os/signal"
	"time"
	"errors"

	"github.com/op/go-logging"
	"github.com/urfave/cli"

	"github.com/ziyan/shadowtunnel/client"
	"github.com/ziyan/shadowtunnel/config"
	"github.com/ziyan/shadowtunnel/server"
)

var log = logging.MustGetLogger("cli")

var (
	ErrInvalidArgument = errors.New("invalid argument")
)

func configureLogging(level, format string) {
	logging.SetBackend(logging.NewBackendFormatter(
		logging.NewLogBackend(os.Stderr, "", 0),
		logging.MustStringFormatter(format),
	))
	if level, err := logging.LogLevel(level); err == nil {
		logging.SetLevel(level, "")
	}
	log.Debugf("log level set to %s", logging.GetLevel(""))
}

func Run(args []string) {

	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "gatewaysshd"
	app.Version = "0.1.0"
	app.Usage = "A daemon that provides a meeting place for all your SSH tunnels."

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level",
			Value: "INFO",
			Usage: "log level",
		},
		cli.StringFlag{
			Name:  "log-format",
			Value: "%{color}%{time:2006-01-02T15:04:05.000Z07:00} [%{level:.4s}] [%{shortfile} %{shortfunc}] %{message}%{color:reset}",
			Usage: "log format",
		},
		cli.StringFlag{
			Name:  "config",
			Value: "",
			Usage: "path to config file",
		},
		cli.BoolFlag{
			Name:  "server",
			Usage: "server mode",
		},
		cli.StringFlag{
			Name:  "password",
			Value: "",
			Usage: "pre-shared password used to establish encryption",
		},
		cli.StringFlag{
			Name:  "listen",
			Value: "",
			Usage: "listen on local endpoint",
		},
		cli.StringFlag{
			Name:  "connect",
			Value: "",
			Usage: "connect to remote endpoint",
		},
		cli.StringFlag{
			Name:  "timeout",
			Value: "2s",
			Usage: "connect timeout",
		},
	}

	app.Action = func(c *cli.Context) error {
		configureLogging(c.String("log-level"), c.String("log-format"))

		// load the config
		var configs *config.Config
		if c.String("config") != "" {
			data, err := ioutil.ReadFile(c.String("config"))
			if err != nil {
				log.Errorf("failed to load configuration from file \"%s\": %s", c.String("config"), err)
				return err
			}
			configs, err = config.ParseConfig(data)
			if err != nil {
				log.Errorf("failed to parse configuration from file \"%s\": %s", c.String("config"), err)
				return err
			}
		} else {
			configs = config.SimpleConfig(c.Bool("server"), c.String("listen"), c.String("connect"), c.String("password"), c.String("timeout"))
		}

		log.Infof("configuration loaded:\n%v", configs)

		servers := make([]*server.Server, 0, len(configs.Servers))
		for _, config := range configs.Servers {
			if config.Listen == "" {
				log.Errorf("listen endpoint not specified")
				return ErrInvalidArgument
			}
			if config.Connect == "" {
				log.Errorf("connect endpoint not specified")
				return ErrInvalidArgument
			}
			if config.Password == "" {
				log.Errorf("pre-shared password not specified")
				return ErrInvalidArgument
			}
			timeout, err := time.ParseDuration(config.Timeout)
			if err != nil {
				log.Errorf("failed to parse timeout \"%s\": %s", config.Timeout, err)
				return err
			}
			server, err := server.NewServer([]byte(config.Password), config.Listen, config.Connect, timeout)
			if err != nil {
				log.Errorf("failed to create server on endpoint \"%s\": %s", config.Listen, err)
				return err
			}

			log.Noticef("listening on %s in server mode", config.Listen)
			servers = append(servers, server)
		}
		defer func() {
			for _, server := range servers {
				server.Close()
			}
		}()

		clients := make([]*client.Client, 0, len(configs.Clients))
		for _, config := range configs.Clients {
			if config.Listen == "" {
				log.Errorf("listen endpoint not specified")
				return ErrInvalidArgument
			}
			if config.Connect == "" {
				log.Errorf("connect endpoint not specified")
				return ErrInvalidArgument
			}
			if config.Password == "" {
				log.Errorf("pre-shared password not specified")
				return ErrInvalidArgument
			}
			timeout, err := time.ParseDuration(config.Timeout)
			if err != nil {
				log.Errorf("failed to parse timeout \"%s\": %s", config.Timeout, err)
				return err
			}
			client, err := client.NewClient([]byte(config.Password), config.Listen, config.Connect, timeout)
			if err != nil {
				log.Errorf("failed to create client on endpoint \"%s\": %s", config.Listen, err)
				return err
			}

			log.Noticef("listening on %s in client mode", config.Listen)
			clients = append(clients, client)
		}
		defer func() {
			for _, client := range clients {
				client.Close()
			}
		}()

		// wait till exit
		signaling := make(chan os.Signal, 1)
		signal.Notify(signaling, os.Interrupt)
		quit := false
		for !quit {
			select {
			case <-signaling:
				quit = true
			case <-time.After(30 * time.Second):
				// gateway.ScavengeSessions(time.Duration(*IDLE_TIMEOUT) * time.Second)
			}
		}

		log.Noticef("exiting ...")
		return nil
	}

	app.Run(args)
}
