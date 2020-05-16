package opennetzteil

import (
	"bufio"
	"sync"
)

type Netzteil interface {
	Probe() error
	GetMaster() (bool, error)
	SetMaster(enabled bool) error
	GetIdent() (string, error)
	SetBeep(enabled bool) error
	GetChannels() ([]int, error)
	GetCurrent(channel float64) (float64, error)
	SetCurrent(channel int, current float64) error
	GetVoltage(channel int) (float64, error)
	SetVoltage(channel int, voltage float64) error
	GetOut(channel int) (bool, error)
	SetOut(channel int, enabled bool) error
	GetOCP(channel int) (bool, error)
	SetOCP(channel int, enabled bool) error
	GetOVP(channel int) (bool, error)
	SetOVP(channel int, enabled bool) error
}

type NetzteilBase struct {
	mutex sync.Mutex
}

func (nt *NetzteilBase) SendCommand(w *bufio.Writer, cmd []byte) error {
	nt.mutex.Lock()
	defer nt.mutex.Unlock()

	_, err := w.Write(cmd)
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}
