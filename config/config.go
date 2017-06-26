// define configuration structrues
package config

import (
	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	// local endpoint to listen on, for example :2020
	Listen string `yaml:"listen"`

	// password for encryption
	Password string `yaml:"password"`

	// remote endpoint to connect to, leave empty if you want to be a socks proxy
	Connect string `yaml:"connect,omitempty"`

	// whether to compress connection
	Compress bool `yaml:"compress"`

	// timeout while connecting
	Timeout string `yaml:"timeout,omitempty"`
}

type ClientConfig struct {

	// local endpoint to listen on, for example :2020
	Listen string `yaml:"listen"`

	// remote endpoint to connect to
	Connect string `yaml:"connect"`

	// password for encryption
	Password string `yaml:"password"`

	// whether to compress connection
	Compress bool `yaml:"compress"`

	// timeout while connecting
	Timeout string `yaml:"timeout,omitempty"`
}

// configuration structure
type Config struct {

	// expose as servers
	Servers []*ServerConfig `yaml:"servers,omitempty"`

	// expose as clients
	Clients []*ClientConfig `yaml:"clients,omitempty"`
}

func (c *Config) String() string {
	d, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(d)
}

func ParseConfig(data []byte) (*Config, error) {
	conf := &Config{}

	err := yaml.Unmarshal([]byte(data), conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func SimpleConfig(server bool, listen, connect, password string, compress bool, timeout string) *Config {
	if server {
		return &Config{
			Servers: []*ServerConfig{
				&ServerConfig{
					Listen:   listen,
					Connect:  connect,
					Password: password,
					Compress: compress,
					Timeout:  timeout,
				},
			},
		}
	} else {
		return &Config{
			Clients: []*ClientConfig{
				&ClientConfig{
					Listen:   listen,
					Connect:  connect,
					Password: password,
					Compress: compress,
					Timeout:  timeout,
				},
			},
		}
	}
}
