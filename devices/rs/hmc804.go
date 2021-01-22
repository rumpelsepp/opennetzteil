package rs

import (
	"fmt"
	"strconv"

	"git.sr.ht/~rumpelsepp/opennetzteil"
)

type HMC804 struct {
	opennetzteil.NetzteilBase
	target string
}

type Status struct {
	ChannelMode string
	Output      bool
}

func NewHMC804(target, name string) *HMC804 {
	return &HMC804{
		target: target,
	}
}

func (nt *HMC804) getIdent() (string, error) {
	cmd := "*IDN?"
	resp, err := nt.TCPRequest(nt.target, cmd)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (nt *HMC804) Probe() error {
	ident, err := nt.getIdent()
	if err != nil {
		return err
	}
	nt.NetzteilBase.Ident = ident
	return nil
}

func (nt *HMC804) Status() (interface{}, error) {
	return false, opennetzteil.ErrNotImplemented
}

func (nt *HMC804) GetMaster() (bool, error) {
	resp, err := nt.TCPRequest(nt.target, "OUTP:MAST:STAT?")
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(string(resp))
}

func (nt *HMC804) SetMaster(enabled bool) error {
	var cmd string
	if enabled {
		cmd = "OUTP:MAST ON"
	} else {
		cmd = "OUTP:MAST OFF"
	}
	if err := nt.TCPSend(nt.target, cmd); err != nil {
		return err
	}
	return nil
}

func (nt *HMC804) SetBeep(enabled bool) error {
	return opennetzteil.ErrNotImplemented
}

func (nt *HMC804) GetChannels() (int, error) {
	return 3, nil
}

func (nt *HMC804) GetCurrent(channel int) (float64, error) {
	cmd := fmt.Sprintf("INST OUT%d", channel)
	if err := nt.TCPSend(nt.target, cmd); err != nil {
		return 0, err
	}
	resp, err := nt.TCPRequest(nt.target, "CURR?")
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(string(resp), 32)
}

func (nt *HMC804) SetCurrent(channel int, current float64) error {
	var cmds []string
	cmd := fmt.Sprintf("INST OUT%d", channel)
	cmds = append(cmds, cmd)
	cmd = fmt.Sprintf("CURR %.3f", current)
	cmds = append(cmds, cmd)
	if err := nt.TCPSendBatched(nt.target, cmds); err != nil {
		return err
	}
	return nil
}

func (nt *HMC804) GetVoltage(channel int) (float64, error) {
	cmd := fmt.Sprintf("INST OUT%d", channel)
	if err := nt.TCPSend(nt.target, cmd); err != nil {
		return 0, err
	}
	resp, err := nt.TCPRequest(nt.target, "VOLT?")
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(string(resp), 32)
}

func (nt *HMC804) SetVoltage(channel int, voltage float64) error {
	var cmds []string
	cmd := fmt.Sprintf("INST OUT%d", channel)
	cmds = append(cmds, cmd)
	cmd = fmt.Sprintf("VOLT %.3f", voltage)
	cmds = append(cmds, cmd)
	if err := nt.TCPSendBatched(nt.target, cmds); err != nil {
		return err
	}
	return nil
}

func (nt *HMC804) GetOut(channel int) (bool, error) {
	cmd := fmt.Sprintf("INST OUT%d", channel)
	if err := nt.TCPSend(nt.target, cmd); err != nil {
		return false, err
	}
	resp, err := nt.TCPRequest(nt.target, "OUTP:STAT?")
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(string(resp))
}

func (nt *HMC804) SetOut(channel int, enabled bool) error {
	var cmds []string
	cmd := fmt.Sprintf("INST OUT%d", channel)
	if err := nt.TCPSend(nt.target, cmd); err != nil {
		return err
	}
	cmds = append(cmds, cmd)
	if enabled {
		cmd = "OUTP:CHAN ON"
	} else {
		cmd = "OUTP:CHAN OFF"
	}
	cmds = append(cmds, cmd)
	if err := nt.TCPSendBatched(nt.target, cmds); err != nil {
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
