package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"git.sr.ht/~rumpelsepp/opennetzteil"
	"git.sr.ht/~rumpelsepp/opennetzteil/rnd"
	"git.sr.ht/~rumpelsepp/opennetzteil/rs"
	"git.sr.ht/~rumpelsepp/rlog"
	"git.sr.ht/~sircmpwn/getopt"
	"github.com/pelletier/go-toml"
)

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
		case "rnd320":
			if handle.Scheme != "file" {
				return nil, fmt.Errorf("invalid handle for: %s", nc.Model)
			}
			nt = rnd.NewRND320(handle.Path)
		case "hmc804":
			if handle.Scheme != "tcp" {
				return nil, fmt.Errorf("invalid handle for: %s", nc.Model)
			}
			fmt.Println(handle.Host)
			nt = rs.NewHMC804(handle.Host)
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
		rlog.Crit(err)
	}

	if opts.help {
		getopt.Usage()
		os.Exit(0)
	}

	if opts.verbose {
		rlog.SetLogLevel(rlog.DEBUG)
	}

	config, err := loadConfig(opts.config)
	if err != nil {
		rlog.Crit(err)
	}

	netzteile, err := initNetzteile(config)
	if err != nil {
		rlog.Crit(err)
	}

	apiSRV := opennetzteil.HTTPServer{
		ReqLog:  os.Stderr,
		Logger:  rlog.NewLogger(os.Stderr),
		Devices: netzteile,
	}
	apiSRV.Logger.SetLogLevel(rlog.DEBUG)
	srv := &http.Server{
		Addr:         config.HTTP.Bind,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      apiSRV.CreateHandler(),
	}

	if err := srv.ListenAndServe(); err != nil {
		rlog.Critln(err)
	}
}
