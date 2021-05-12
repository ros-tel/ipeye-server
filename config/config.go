package config

import (
	"github.com/gin-gonic/gin"
)

type (
	// configuration contains the application settings
	Conf struct {
		Listen   Listen    `yaml:"listen"`
		Recorder Recorder  `yaml:"recorder"`
		Balancer *Balancer `yaml:"balancer"`
	}

	Balancer struct {
		Servers map[string]Server `yaml:"servers"`
		Routes  map[string]string `yaml:"routes"`
	}

	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}

	Listen struct {
		Api    string `yaml:"api"`
		Server string `yaml:"server"`
	}

	Recorder struct {
		BaseDir string `yaml:"base_dir"`
		Cmd     string `yaml:"cmd"`
		Params  string `yaml:"params"`
	}
)

func Inject(cnf *Conf) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("cnf", cnf.Balancer)
	}
}
