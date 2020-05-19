package rnd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"git.sr.ht/~rumpelsepp/opennetzteil"
)

type RND320 struct {
	opennetzteil.NetzteilBase
	ident string
	file  *os.File
}

const (
	ChannelModeCC = 1 << iota
	ChannelModeVC
)

type Status struct {
	ChannelMode string
	Output      bool
}

func NewRND320(path string) *RND320 {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	return &RND320{
		file: file,
	}
}

func (nt *RND320) Probe() error {
	cmd := []byte("*IDN?")
	resp, err := nt.RequestWithTimeout(nt.file, cmd, 100*time.Millisecond)
	if err != nil {
		return err
	}
	nt.ident = string(resp)
	return nil
}

func (nt *RND320) Status() (interface{}, error) {
	cmd := []byte("STATUS?")
	resp, err := nt.RequestWithTimeout(nt.file, cmd, 100*time.Millisecond)
	if err != nil {
		return nil, err
	}
	if len(resp) != 1 {
		fmt.Println(resp)
		return nil, fmt.Errorf("invalid data from device received")
	}

	var (
		mode   string
		output bool
	)
	if resp[0]&0x01 == 1 {
		mode = "CV"
	} else {
		mode = "CC"
	}
	if resp[0]&0x40 == 1 {
		output = true
	} else {
		output = false
	}
	status := Status{
		ChannelMode: mode,
		Output:      output,
	}
	return status, nil
}

func (nt *RND320) GetMaster() (bool, error) {
	status, err := nt.Status()
	if err != nil {
		return false, err
	}
	s := status.(Status)
	return s.Output, nil
}

func (nt *RND320) SetMaster(enabled bool) error {
	var cmd []byte
	if enabled {
		cmd = []byte("OUT1")
	} else {
		cmd = []byte("OUT0")
	}

	if err := nt.SendCommand(nt.file, cmd); err != nil {
		return err
	}
	return nil
}

func (nt *RND320) GetIdent() (string, error) {
	return nt.ident, nil
}

func (nt *RND320) SetBeep(enabled bool) error {
	return opennetzteil.ErrNotImplemented
}

func (nt *RND320) GetChannels() (int, error) {
	return 1, nil
}

func (nt *RND320) GetCurrent(channel int) (float64, error) {
	cmd := []byte(fmt.Sprintf("IOUT%d?", channel))
	resp, err := nt.RequestWithTimeout(nt.file, cmd, 100*time.Millisecond)
	if err != nil {
		return 0, err
	}

	current, err := strconv.ParseFloat(string(resp), 32)
	if err != nil {
		return 0, err
	}
	return current, nil
}

func (nt *RND320) SetCurrent(channel int, current float64) error {
	cmd := []byte(fmt.Sprintf("ISET%d:%.2f", channel, current))
	err := nt.SendCommand(nt.file, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (nt *RND320) GetVoltage(channel int) (float64, error) {
	cmd := []byte(fmt.Sprintf("VOUT%d?", channel))
	resp, err := nt.RequestWithTimeout(nt.file, cmd, 100*time.Millisecond)
	if err != nil {
		return 0, err
	}

	voltage, err := strconv.ParseFloat(string(resp), 32)
	if err != nil {
		return 0, err
	}
	return voltage, nil
}

func (nt *RND320) SetVoltage(channel int, voltage float64) error {
	cmd := []byte(fmt.Sprintf("VSET%d:%.2f", channel, voltage))
	err := nt.SendCommand(nt.file, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (nt *RND320) GetOut(channel int) (bool, error) {
	return nt.GetMaster()
}

func (nt *RND320) SetOut(channel int, enabled bool) error {
	if channel > 1 {
		return fmt.Errorf("channel not avail")
	}
	return nt.SetMaster(enabled)
}

func (nt *RND320) GetOCP(channel int) (bool, error) {
	return false, opennetzteil.ErrNotImplemented
}

func (nt *RND320) SetOCP(channel int, enabled bool) error {
	return opennetzteil.ErrNotImplemented
}

func (nt *RND320) GetOVP(channel int) (bool, error) {
	return false, opennetzteil.ErrNotImplemented
}

func (nt *RND320) SetOVP(channel int, enabled bool) error {
	return opennetzteil.ErrNotImplemented
}
