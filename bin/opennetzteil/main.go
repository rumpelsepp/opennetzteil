package main

import (
	"net/http"
	"os"
	"time"

	"git.sr.ht/~rumpelsepp/opennetzteil"
	"git.sr.ht/~rumpelsepp/opennetzteil/rnd"
	"git.sr.ht/~rumpelsepp/rlog"
)

type runtimeOptions struct {
	bind string
}

func main() {
	opts := runtimeOptions{}
	opts.bind = ":8000"
	netzteil := rnd.NewRND320("/dev/ttyACM0")
	rlog.SetLogLevel(rlog.DEBUG)

	if err := netzteil.Probe(); err != nil {
		rlog.Critf("netzteil probe failed: %s", err)
	}
	rlog.Debug("probing complete")

	apiSRV := opennetzteil.HTTPServer{
		ReqLog:  os.Stderr,
		Logger:  rlog.NewLogger(os.Stderr),
		Devices: []opennetzteil.Netzteil{netzteil},
	}
	apiSRV.Logger.SetLogLevel(rlog.DEBUG)
	srv := &http.Server{
		Addr:         opts.bind,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      apiSRV.CreateHandler(),
	}

	if err := srv.ListenAndServe(); err != nil {
		rlog.Critln(err)
	}
}
