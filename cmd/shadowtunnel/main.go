package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/golang/glog"
	"github.com/ziyan/shadowtunnel/pkg/shadowtunnel/client"
	"github.com/ziyan/shadowtunnel/pkg/shadowtunnel/config"
	"github.com/ziyan/shadowtunnel/pkg/shadowtunnel/server"
)

var (
	CONFIG   = flag.String("config", "", "path to config file")
	SERVER   = flag.Bool("server", false, "server mode")
	PASSWORD = flag.String("password", "", "password used to encrypt connection")
	LISTEN   = flag.String("listen", "", "listen on local endpoint")
	CONNECT  = flag.String("connect", "", "connect to remote endpoint")
	TIMEOUT  = flag.String("timeout", "2s", "connect timeout")
)

func main() {

	// use all CPU cores for maximum performance
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	// load the config
	var c *config.Config
	if *CONFIG != "" {
		data, err := ioutil.ReadFile(*CONFIG)
		if err != nil {
			panic(err)
		}

		c, err = config.ParseConfig(data)
		if err != nil {
			panic(err)
		}
	} else {
		c = config.SimpleConfig(*SERVER, *LISTEN, *CONNECT, *PASSWORD, *TIMEOUT)
	}

	glog.Infof("configuration loaded:\n%v", c)

	servers := make([]*server.Server, 0, len(c.Servers))
	for _, c := range c.Servers {
		timeout, err := time.ParseDuration(c.Timeout)
		if err != nil {
			timeout = 2 * time.Second
		}
		server, err := server.NewServer([]byte(c.Password), c.Listen, c.Connect, timeout)
		if err != nil {
			panic(err)
		}
		servers = append(servers, server)
	}
	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	clients := make([]*client.Client, 0, len(c.Clients))
	for _, c := range c.Clients {
		timeout, err := time.ParseDuration(c.Timeout)
		if err != nil {
			timeout = 2 * time.Second
		}
		client, err := client.NewClient([]byte(c.Password), c.Listen, c.Connect, timeout)
		if err != nil {
			panic(err)
		}
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

	glog.Infof("exiting ...")
}

