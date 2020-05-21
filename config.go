package main

import (
	"flag"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Configuration is a struct containing the variables that are read from the environment
type Configuration struct {
	LocalDatastore bool          `envconfig:"local"      required:"false" default:"false"    desc:"when enabled, uses internal datastore rather than postgres"`
	Port           uint16        `envconfig:"port"       required:"false" default:"8080"     desc:"http port used for listening"`
	PingTimeout    time.Duration `envconfig:"dbtimeout"  required:"false" default:"20s"      desc:"ping timeout for db connection"`
	PingInterval   time.Duration `envconfig:"dbinterval" required:"false" default:"200ms"    desc:"ping interval for db connection"`
	DbHost         string        `envconfig:"dbhost"     required:"false" default:"postgres" desc:"hostname for postgres db"`
	DbPort         uint16        `envconfig:"dbport"     required:"false" default:"5432"     desc:"port for postgres db"`
	DbUser         string        `envconfig:"dbuser"     required:"false" default:"postgres" desc:"username for postgres db"`
	DbPass         string        `envconfig:"dbpass"     required:"false" default:"postgres" desc:"password for postgres db"`
	DbName         string        `envconfig:"dbname"     required:"false" default:"test"     desc:"database name for postgres db"`
}

// NewConfiguration returns the application config based on the default values and the environment.
func NewConfiguration() (*Configuration, error) {
	var ev Configuration
	err := envconfig.Process("cc", &ev)
	if err != nil {
		return nil, err
	}
	flag.Usage = func() {
		envconfig.Usagef("cc", ev, os.Stderr, envconfig.DefaultListFormat)
	}
	flag.Parse()
	return &ev, nil
}
