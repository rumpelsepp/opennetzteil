package rs

import (
	"fmt"
	"net"

	"git.sr.ht/~rumpelsepp/opennetzteil"
)

type HMC804 struct {
	opennetzteil.NetzteilBase
	ident  string
	target string
}

type Status struct {
	ChannelMode string
	Output      bool
}

func NewHMC804(target string) *HMC804 {
	return &HMC804{
		target: target,
	}
}

func (nt *HMC804) send(cmd []byte) error {
	conn, err := net.Dial("tcp", nt.target)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nt.SendCommandLine(conn, cmd)
}

func (nt *HMC804) request(cmd []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", nt.target)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return nt.RequestLine(conn, cmd)
}

func (nt *HMC804) Probe() error {
	cmd := []byte("*IDN?")
	resp, err := nt.request(cmd)
	if err != nil {
		return err
	}
	nt.ident = string(resp)
	return nil
}

func (nt *HMC804) Status() (interface{}, error) {
	return false, opennetzteil.ErrNotImplemented
}

func (nt *HMC804) GetMaster() (bool, error) {
	return false, opennetzteil.ErrNotImplemented
}

func (nt *HMC804) SetMaster(enabled bool) error {
	var cmd []byte
	if enabled {
		cmd = []byte("OUTP:MAST ON")
	} else {
		cmd = []byte("OUTP:MAST OFF")
	}
	if err := nt.send(cmd); err != nil {
		return err
	}
	return nil
}

func (nt *HMC804) GetIdent() (string, error) {
	return nt.ident, nil
}

func (nt *HMC804) SetBeep(enabled bool) error {
	return opennetzteil.ErrNotImplemented
}

func (nt *HMC804) GetChannels() (int, error) {
	return 3, nil
}

func (nt *HMC804) GetCurrent(channel int) (float64, error) {
	return 0, opennetzteil.ErrNotImplemented
}

func (nt *HMC804) SetCurrent(channel int, current float64) error {
	return opennetzteil.ErrNotImplemented
}

func (nt *HMC804) GetVoltage(channel int) (float64, error) {
	return 0, opennetzteil.ErrNotImplemented
}

func (nt *HMC804) SetVoltage(channel int, voltage float64) error {
	return opennetzteil.ErrNotImplemented
}

func (nt *HMC804) GetOut(channel int) (bool, error) {
	return nt.GetMaster()
}

func (nt *HMC804) SetOut(channel int, enabled bool) error {
	cmd := []byte(fmt.Sprintf("INST OUT%d", channel))
	if err := nt.send(cmd); err != nil {
		return err
	}
	if enabled {
		cmd = []byte("OUTP:CHAN ON")
	} else {
		cmd = []byte("OUTP:CHAN OFF")
	}
	if err := nt.send(cmd); err != nil {
		return err
	}
	return nil
}

func (nt *HMC804) GetOCP(channel int) (bool, error) {
	return false, opennetzteil.ErrNotImplemented
}

func (nt *HMC804) SetOCP(channel int, enabled bool) error {
	return opennetzteil.ErrNotImplemented
}

func (nt *HMC804) GetOVP(channel int) (bool, error) {
	return false, opennetzteil.ErrNotImplemented
}

func (nt *HMC804) SetOVP(channel int, enabled bool) error {
	return opennetzteil.ErrNotImplemented
}