package opennetzteil

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~rumpelsepp/helpers"
	"github.com/Fraunhofer-AISEC/penlogger"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type HTTPServer struct {
	ReqLog  io.Writer
	Devices []Netzteil
	Logger  *penlogger.Logger
}

type measurement struct {
	Current float64   `json:"current,omitempty"`
	Voltage float64   `json:"voltage,omitempty"`
	Time    time.Time `json:"time"`
}

func (s *HTTPServer) lookupDevice(w http.ResponseWriter, vars map[string]string) (Netzteil, error) {
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusNotFound)
		return nil, err
	}
	// Opennetzteil ids start with 1.
	// Thus it need to be decremented for the list lookup.
	id--
	if id < 0 || len(s.Devices)-1 < id {
		err := fmt.Errorf("device does not exist")
		helpers.SendJSONError(w, err.Error(), http.StatusNotFound)
		return nil, err
	}
	return s.Devices[id], nil
}

func parseChannel(vars map[string]string) (int, error) {
	channel, err := strconv.Atoi(vars["channel"])
	if err != nil {
		return 0, err
	}
	return channel, nil
}

func (s *HTTPServer) lookupDevAndParseChannel(w http.ResponseWriter, vars map[string]string) (Netzteil, int, error) {
	dev, err := s.lookupDevice(w, vars)
	// TODO: more different error types
	if err != nil {
		return nil, 0, err
	}
	nChannels, err := dev.GetChannels()
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return nil, 0, err
	}
	channel, err := strconv.Atoi(vars["channel"])
	// TODO: more different error types
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusNotFound)
		return nil, 0, err
	}
	if channel > nChannels {
		err := fmt.Errorf("no such channel '%d'; device has '%d' channels", channel, nChannels)
		helpers.SendJSONError(w, err.Error(), http.StatusNotFound)
		return nil, 0, err
	}
	return dev, channel, nil
}

// https://godoc.org/github.com/gorilla/websocket#hdr-Control_Messages
func readLoop(c *websocket.Conn) {
	for {
		if _, _, err := c.NextReader(); err != nil {
			c.Close()
			break
		}
	}
}

// Handlers for full API
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
	s.Logger.LogDebugf("device list: %v", resp)
	helpers.SendJSON(w, resp)
}

func (s *HTTPServer) getIndent(w http.ResponseWriter, r *http.Request) {
	dev, err := s.lookupDevice(w, mux.Vars(r))
	if err != nil {
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
	helpers.SendJSONError(w, "not implemented", http.StatusInternalServerError)
}

func (s *HTTPServer) getMaster(w http.ResponseWriter, r *http.Request) {
	dev, err := s.lookupDevice(w, mux.Vars(r))
	if err != nil {
		return
	}
	state, err := dev.GetMaster()
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.SendJSON(w, state)
}

func (s *HTTPServer) putMaster(w http.ResponseWriter, r *http.Request) {
	var req bool
	dev, err := s.lookupDevice(w, mux.Vars(r))
	if err != nil {
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

func (s *HTTPServer) getStatus(w http.ResponseWriter, r *http.Request) {
	dev, err := s.lookupDevice(w, mux.Vars(r))
	if err != nil {
		return
	}
	status, err := dev.Status()
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.SendJSON(w, status)
}

func (s *HTTPServer) getChannels(w http.ResponseWriter, r *http.Request) {
	var (
		vars = mux.Vars(r)
	)
	dev, err := s.lookupDevice(w, vars)
	if err != nil {
		return
	}

	channels, err := dev.GetChannels()
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.SendJSON(w, channels)
}

func (s *HTTPServer) getCurrent(w http.ResponseWriter, r *http.Request) {
	var (
		vars = mux.Vars(r)
	)
	dev, channel, err := s.lookupDevAndParseChannel(w, vars)
	if err != nil {
		return
	}

	current, err := dev.GetCurrent(channel)
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.SendJSON(w, current)
}

func (s *HTTPServer) putCurrent(w http.ResponseWriter, r *http.Request) {
	var (
		req  float64
		vars = mux.Vars(r)
	)
	dev, channel, err := s.lookupDevAndParseChannel(w, vars)
	if err != nil {
		return
	}
	err = helpers.RecvJSON(r, &req)
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := dev.SetCurrent(channel, req); err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *HTTPServer) getVoltage(w http.ResponseWriter, r *http.Request) {
	var (
		vars = mux.Vars(r)
	)
	dev, channel, err := s.lookupDevAndParseChannel(w, vars)
	if err != nil {
		return
	}
	voltage, err := dev.GetVoltage(channel)
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.SendJSON(w, voltage)
}

const (
	measurementVoltage = iota
	measurementCurrent
	measurementBoth
)

func (s *HTTPServer) continousMeasurement(w http.ResponseWriter, r *http.Request, dev Netzteil, channel, mtype, interval int) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Logger.LogError(err)
		return
	}
	defer conn.Close()
	go readLoop(conn)

	for {
		var (
			val float64
			err error
			m   measurement
		)
		switch mtype {
		case measurementVoltage:
			val, err = dev.GetVoltage(channel)
			m.Time = time.Now()
			m.Voltage = val
		case measurementCurrent:
			val, err = dev.GetCurrent(channel)
			m.Time = time.Now()
			m.Current = val
		case measurementBoth:
			val, err = dev.GetVoltage(channel)
			if err != nil {
				m := map[string]string{"error": err.Error()}
				if err := conn.WriteJSON(m); err != nil {
					return
				}
				continue
			}
			m.Time = time.Now()
			m.Voltage = val
			val, err = dev.GetCurrent(channel)
			m.Current = val
		default:
			panic("BUG: this invalid measurement type")
		}
		if err != nil {
			m := map[string]string{"error": err.Error()}
			if err := conn.WriteJSON(m); err != nil {
				return
			}
			continue
		}
		if err := conn.WriteJSON(m); err != nil {
			return
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

func (s *HTTPServer) getVoltageWS(w http.ResponseWriter, r *http.Request) {
	var (
		vars = mux.Vars(r)
	)
	dev, channel, err := s.lookupDevAndParseChannel(w, vars)
	if err != nil {
		return
	}

	query := r.URL.Query()
	interval, err := strconv.ParseUint(query.Get("interval"), 10, 32)
	if err != nil {
		s.Logger.LogError(err)
		return
	}
	s.continousMeasurement(w, r, dev, channel, measurementVoltage, int(interval))
}

func (s *HTTPServer) getCurrentWS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dev, channel, err := s.lookupDevAndParseChannel(w, vars)
	if err != nil {
		return
	}

	query := r.URL.Query()
	interval, err := strconv.ParseUint(query.Get("interval"), 10, 32)
	if err != nil {
		s.Logger.LogError(err)
		return
	}
	s.continousMeasurement(w, r, dev, channel, measurementCurrent, int(interval))
}

func (s *HTTPServer) getMeasurementsWS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dev, channel, err := s.lookupDevAndParseChannel(w, vars)
	if err != nil {
		return
	}

	query := r.URL.Query()
	interval, err := strconv.ParseUint(query.Get("interval"), 10, 32)
	if err != nil {
		s.Logger.LogError(err)
		return
	}
	s.continousMeasurement(w, r, dev, channel, measurementBoth, int(interval))
}

func (s *HTTPServer) putVoltage(w http.ResponseWriter, r *http.Request) {
	var (
		req  float64
		vars = mux.Vars(r)
	)
	dev, channel, err := s.lookupDevAndParseChannel(w, vars)
	if err != nil {
		return
	}
	err = helpers.RecvJSON(r, &req)
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
	var (
		on   bool
		err  error
		vars = mux.Vars(r)
	)
	dev, channel, err := s.lookupDevAndParseChannel(w, vars)
	if err != nil {
		return
	}
	if channel == 0 {
		on, err = dev.GetMaster()
	} else {
		on, err = dev.GetOut(channel)
	}
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.SendJSON(w, on)
}

func (s *HTTPServer) putOut(w http.ResponseWriter, r *http.Request) {
	var (
		req  bool
		vars = mux.Vars(r)
	)
	dev, channel, err := s.lookupDevAndParseChannel(w, vars)
	if err != nil {
		return
	}

	err = helpers.RecvJSON(r, &req)
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = dev.SetOut(channel, req)
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *HTTPServer) getOcp(w http.ResponseWriter, r *http.Request) {
	helpers.SendJSONError(w, "not implemented", http.StatusInternalServerError)
}

func (s *HTTPServer) putOcp(w http.ResponseWriter, r *http.Request) {
	helpers.SendJSONError(w, "not implemented", http.StatusInternalServerError)
}

func (s *HTTPServer) getOvp(w http.ResponseWriter, r *http.Request) {
	helpers.SendJSONError(w, "not implemented", http.StatusInternalServerError)
}

func (s *HTTPServer) putOvp(w http.ResponseWriter, r *http.Request) {
	helpers.SendJSONError(w, "not implemented", http.StatusInternalServerError)
}

// Magic handler for reduced API
func (s *HTTPServer) redAPI(w http.ResponseWriter, r *http.Request) {
	var (
		u         = r.URL
		vars      = mux.Vars(r)
		id        = vars["id"]
		devPrefix = "/_netzteil/api/device/"
		chPrefix  = fmt.Sprintf("/_netzteil/api/devices/%s/channel/", id)
		path      = ""
	)
	if strings.HasPrefix(u.Path, devPrefix) {
		pathSuffix := strings.TrimPrefix(u.Path, devPrefix)
		path = fmt.Sprintf("/_netzteil/api/devices/0/%s", pathSuffix)
	} else if strings.HasPrefix(u.Path, chPrefix) {
		pathSuffix := strings.TrimPrefix(u.Path, chPrefix)
		path = fmt.Sprintf("/_netzteil/api/devices/%s/channels/0/%s", id, pathSuffix)
	} else {
		helpers.SendJSONError(w, "wrong prefix", http.StatusNotFound)
	}
	http.Redirect(w, r, path, http.StatusPermanentRedirect)
}

func (s *HTTPServer) CreateHandler() http.Handler {
	r := mux.NewRouter()
	api := r.PathPrefix("/_netzteil/api").Subrouter()
	api.HandleFunc("/devices", s.getDevices).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id:[0-9]+}/ident", s.getIndent).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id:[0-9]+}/beep", s.putBeep).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id:[0-9]+}/out", s.getMaster).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id:[0-9]+}/out", s.putMaster).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id:[0-9]+}/status", s.getStatus).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id:[0-9]+}/channels", s.getChannels).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/current", s.getCurrent).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/current", s.putCurrent).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/current/ws", s.getCurrentWS).Methods(http.MethodGet).Queries("interval", "{interval:[0-9]+}")
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/voltage", s.getVoltage).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/voltage", s.putVoltage).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/voltage/ws", s.getVoltageWS).Methods(http.MethodGet).Queries("interval", "{interval:[0-9]+}")
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/measurements/ws", s.getMeasurementsWS).Methods(http.MethodGet).Queries("interval", "{interval:[0-9]+}")
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/out", s.getOut).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/out", s.putOut).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/ocp", s.getOcp).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/ocp", s.putOcp).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/ovp", s.getOvp).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/ovp", s.putOvp).Methods(http.MethodPut)
	chPrefix := api.PathPrefix("/devices/{id:[0-9]+}/channel/")
	chPrefix.HandlerFunc(s.redAPI).Methods(http.MethodGet, http.MethodPut)

	// Enable reduced API if only one powersupply device is registered.
	if len(s.Devices) == 1 {
		deviceChPrefix := api.PathPrefix("/device/")
		deviceChPrefix.HandlerFunc(s.redAPI).Methods(http.MethodGet, http.MethodPut)
	}

	return handlers.LoggingHandler(s.ReqLog, r)
}
