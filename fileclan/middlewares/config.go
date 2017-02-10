package middlewares

import (
	"github.com/bbangert/toml"
	"time"
)

type FileClanConf struct {
	Server struct {
		Addr       string
		AllowedIPs []string
	}
	Mongo struct {
		Addr    string
		Timeout time.Duration
		Db      string
		User    string
		Passwd  string
	}
}

var (
	Conf *FileClanConf
)

func LoadConfig(filePath string) error {
	Conf = &FileClanConf{}

	if _, err := toml.DecodeFile(filePath, &Conf); err != nil {
		return err
	}

	return nil
}
