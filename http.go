package opennetzteil

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"git.sr.ht/~rumpelsepp/helpers"
	"git.sr.ht/~rumpelsepp/rlog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type HTTPServer struct {
	ReqLog  io.Writer
	Devices []Netzteil
	Logger  *rlog.Logger
}

func (s *HTTPServer) lookupDevice(vars map[string]string) (Netzteil, error) {
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return nil, err
	}
	if len(s.Devices)-1 < id {
		return nil, fmt.Errorf("device does not exist")
	}
	return s.Devices[id], nil
}

func (s *HTTPServer) getDevices(w http.ResponseWriter, r *http.Request) {
	var resp []string
	for _, dev := range s.Devices {
		ident, err := dev.GetIdent()
		if err != nil {
			// TODO: logging, or failing out?
			continue
		}
		resp = append(resp, ident)
	}
	s.Logger.Debugf("device list: %v", resp)
	helpers.SendJSON(w, resp)
}

func (s *HTTPServer) getIndent(w http.ResponseWriter, r *http.Request) {
	dev, err := s.lookupDevice(mux.Vars(r))
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	ident, err := dev.GetIdent()
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.SendJSON(w, ident)
}

func (s *HTTPServer) putBeep(w http.ResponseWriter, r *http.Request) {
}

func (s *HTTPServer) getMaster(w http.ResponseWriter, r *http.Request) {
}
func (s *HTTPServer) putMaster(w http.ResponseWriter, r *http.Request) {
	var req bool
	dev, err := s.lookupDevice(mux.Vars(r))
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = helpers.RecvJSON(r, &req)
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := dev.SetMaster(req); err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *HTTPServer) getChannels(w http.ResponseWriter, r *http.Request) {
}

func (s *HTTPServer) getCurrent(w http.ResponseWriter, r *http.Request) {
}
func (s *HTTPServer) putCurrent(w http.ResponseWriter, r *http.Request) {
}

func (s *HTTPServer) getVoltage(w http.ResponseWriter, r *http.Request) {
	var (
		vars = mux.Vars(r)
	)
	dev, err := s.lookupDevice(vars)
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Make a helper for this
	channel, err := strconv.Atoi(vars["channel"])
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	voltage, err := dev.GetVoltage(channel)
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.SendJSON(w, voltage)
}

func (s *HTTPServer) putVoltage(w http.ResponseWriter, r *http.Request) {
	var (
		req  float64
		vars = mux.Vars(r)
	)
	dev, err := s.lookupDevice(vars)
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = helpers.RecvJSON(r, &req)
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Make a helper for this
	channel, err := strconv.Atoi(vars["channel"])
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := dev.SetVoltage(channel, req); err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *HTTPServer) getOut(w http.ResponseWriter, r *http.Request) {
}
func (s *HTTPServer) putOut(w http.ResponseWriter, r *http.Request) {
}

func (s *HTTPServer) getOcp(w http.ResponseWriter, r *http.Request) {
}
func (s *HTTPServer) putOcp(w http.ResponseWriter, r *http.Request) {
}
func (s *HTTPServer) getOvp(w http.ResponseWriter, r *http.Request) {
}
func (s *HTTPServer) putOvp(w http.ResponseWriter, r *http.Request) {
}

func (s *HTTPServer) CreateHandler() http.Handler {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/devices", s.getDevices).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id}/ident", s.getIndent).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id}/beep", s.putBeep).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id}/out", s.getMaster).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id}/out", s.putMaster).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id}/channels", s.getChannels).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id}/channels/{channel}/current", s.getCurrent).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id}/channels/{channel}/current", s.putCurrent).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id}/channels/{channel}/voltage", s.getVoltage).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id}/channels/{channel}/voltage", s.putVoltage).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id}/channels/{channel}/out", s.getOut).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id}/channels/{channel}/out", s.putOut).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id}/channels/{channel}/ocp", s.getOcp).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id}/channels/{channel}/ocp", s.putOcp).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id}/channels/{channel}/ovp", s.getOvp).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id}/channels/{channel}/ovp", s.putOvp).Methods(http.MethodPut)

	return handlers.LoggingHandler(s.ReqLog, r)
}
