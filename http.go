package opennetzteil

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~rumpelsepp/helpers"
	"github.com/Fraunhofer-AISEC/penlog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type HTTPServer struct {
	ReqLog  io.Writer
	Devices []Netzteil
	Logger  *penlog.Logger
}

type measurement struct {
	Current float64   `json:"current,omitempty"`
	Voltage float64   `json:"voltage,omitempty"`
	Time    time.Time `json:"time"`
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
	helpers.SendJSONError(w, "not implemented", http.StatusInternalServerError)
}

func (s *HTTPServer) getMaster(w http.ResponseWriter, r *http.Request) {
	dev, err := s.lookupDevice(mux.Vars(r))
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
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

func (s *HTTPServer) getStatus(w http.ResponseWriter, r *http.Request) {
	dev, err := s.lookupDevice(mux.Vars(r))
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
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
	dev, err := s.lookupDevice(vars)
	if err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusBadRequest)
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
	if err := dev.SetCurrent(channel, req); err != nil {
		helpers.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

func (s *HTTPServer) getVoltageWS(w http.ResponseWriter, r *http.Request) {
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

	query := r.URL.Query()
	interval, _ := strconv.ParseUint(query.Get("interval"), 10, 32)
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Logger.LogError(err)
		return
	}
	defer conn.Close()
	go readLoop(conn)

	for {
		voltage, err := dev.GetVoltage(channel)
		if err != nil {
			m := map[string]string{"error": err.Error()}
			if err := conn.WriteJSON(m); err != nil {
				return
			}
			continue
		}
		m := measurement{Time: time.Now(), Voltage: voltage}
		if err := conn.WriteJSON(m); err != nil {
			return
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
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
	helpers.SendJSONError(w, "not implemented", http.StatusInternalServerError)
}

func (s *HTTPServer) putOut(w http.ResponseWriter, r *http.Request) {
	var (
		req  bool
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
	// api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/current/ws", s.getCurrentWS).Methods(http.MethodGet).Queries("interval", "{interval:[0-9]+}")
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/voltage", s.getVoltage).Methods(http.MethodGet)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/voltage", s.putVoltage).Methods(http.MethodPut)
	api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/voltage/ws", s.getVoltageWS).Methods(http.MethodGet).Queries("interval", "{interval:[0-9]+}")
	// api.HandleFunc("/devices/{id:[0-9]+}/channels/{channel:[0-9]+}/measurements/ws", s.getMeasurementsWS).Methods(http.MethodGet).Queries("interval", "{interval:[0-9]+}")
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
