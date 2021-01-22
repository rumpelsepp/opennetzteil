package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.sr.ht/~rumpelsepp/opennetzteil"
	"git.sr.ht/~rumpelsepp/opennetzteil/devices/dummy"
	"git.sr.ht/~rumpelsepp/opennetzteil/devices/ea"
	"git.sr.ht/~rumpelsepp/opennetzteil/devices/rnd"
	"git.sr.ht/~rumpelsepp/opennetzteil/devices/rs"
	"git.sr.ht/~sircmpwn/getopt"
	"github.com/Fraunhofer-AISEC/penlog"
	"github.com/pelletier/go-toml"
)

type requestLogger struct {
	penlog.Logger
}

func (l *requestLogger) Write(p []byte) (int, error) {
	l.Logger.LogDebug(strings.TrimSpace(string(p)))
	return 0, nil
}

type runtimeOptions struct {
	config  string
	verbose bool
	help    bool
}

type HTTPConfig struct {
	Bind string
}

type NetzteilConfig struct {
	Handle string
	Model  string
	Name   string
}

type config struct {
	HTTP      HTTPConfig
	Netzteile []NetzteilConfig
}

func loadConfig(path string) (*config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var conf config
	if err := toml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func configPath() string {
	path, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(path, "netzteil/config.toml")
}

func initNetzteile(conf *config) ([]opennetzteil.Netzteil, error) {
	var netzteile []opennetzteil.Netzteil
	for _, nc := range conf.Netzteile {
		var nt opennetzteil.Netzteil
		handle, err := url.Parse(nc.Handle)
		if err != nil {
			return nil, err
		}

		switch nc.Model {
		case "dummy":
			nt = &dummy.DummyDevice{
				NetzteilBase: opennetzteil.NetzteilBase{
					Ident: "dummy-device",
					Name:  nc.Name,
				},
			}
		case "rnd320":
			if handle.Scheme != "file" {
				return nil, fmt.Errorf("invalid handle for: %s", nc.Model)
			}
			nt, err = rnd.NewRND320(handle.Path, nc.Name)
			if err != nil {
				return nil, err
			}
		case "hmc804":
			if handle.Scheme != "tcp" {
				return nil, fmt.Errorf("invalid handle for: %s", nc.Model)
			}
			nt = rs.NewHMC804(handle.Host, nc.Name)
		case "ea8000":
			if handle.Scheme != "tcp" {
				return nil, fmt.Errorf("invalid handle for: %s", nc.Model)
			}
			nt = ea.NewEA8000(handle.Host, nc.Name)
		default:
			return nil, fmt.Errorf("unsupported power supply")
		}

		if err := nt.Probe(); err != nil {
			return nil, fmt.Errorf("probe failed: %s", err)
		}
		netzteile = append(netzteile, nt)
	}
	return netzteile, nil
}

func main() {
	opts := runtimeOptions{}
	getopt.StringVar(&opts.config, "c", configPath(), "path to the config file")
	getopt.BoolVar(&opts.verbose, "v", false, "enable debugging output")
	getopt.BoolVar(&opts.help, "h", false, "show this page and exit")

	err := getopt.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if opts.help {
		getopt.Usage()
		os.Exit(0)
	}

	var (
		httpLogger = penlog.NewLogger("http", os.Stderr)
		reqLogger  = penlog.NewLogger("http-req", os.Stderr)
	)

	if opts.verbose {
		reqLogger.SetLogLevel(penlog.PrioDebug)
		reqLogger.SetLogLevel(penlog.PrioDebug)
	}

	config, err := loadConfig(opts.config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	netzteile, err := initNetzteile(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	apiSRV := opennetzteil.HTTPServer{
		ReqLog:  reqLogger,
		Logger:  httpLogger,
		Devices: netzteile,
	}
	apiSRV.Logger.SetLogLevel(penlog.PrioDebug)
	srv := &http.Server{
		Addr:         config.HTTP.Bind,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      apiSRV.CreateHandler(),
	}

	if err := srv.ListenAndServe(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
