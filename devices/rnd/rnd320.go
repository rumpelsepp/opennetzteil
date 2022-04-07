package rnd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"time"

	"codeberg.org/rumpelsepp/opennetzteil"
)

type RND320 struct {
	opennetzteil.NetzteilBase
	path string
	file *os.File
}

const (
	ChannelModeCC = 1 << iota
	ChannelModeVC
)

type Status struct {
	ChannelMode string
	Output      bool
}

func NewRND320(path, name string) (*RND320, error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &RND320{
		NetzteilBase: opennetzteil.NetzteilBase{Name: name},
		file:         file,
		path:         path,
	}, nil
}

func (nt *RND320) reopenHandeIfNeeded(err error) error {
	// This happens when the power supply itself is
	// powercycled. In this case the handle must be renewed.
	if errors.Is(err, syscall.EIO) {
		file, err := os.OpenFile(nt.path, os.O_RDWR, 0644)
		if err != nil {
			// The filedescriptor could not be refreshed.
			// This is a fatal error.
			return err
		}
		nt.file = file
	}
	return nil
}

func (nt *RND320) request(cmd string, timeout time.Duration) ([]byte, error) {
	var (
		err  error
		resp []byte
	)
	for i := 0; i < 3; i++ {
		resp, err = nt.RequestWithTimeout(nt.file, []byte(cmd), timeout)
		if err != nil {
			if err := nt.reopenHandeIfNeeded(err); err != nil {
				return nil, err
			}
			// This powersupply is crap, retry 3 times.
			time.Sleep(500 * time.Millisecond)
			continue
		}
		break
	}
	return resp, err
}

func (nt *RND320) command(cmd string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = nt.SendCommand(nt.file, []byte(cmd))
		if err != nil {
			if err := nt.reopenHandeIfNeeded(err); err != nil {
				return err
			}
			// This powersupply is crap, retry 3 times.
			time.Sleep(500 * time.Millisecond)
			continue
		}
		break
	}
	return err
}

func (nt *RND320) Probe() error {
	cmd := "*IDN?"
	resp, err := nt.request(cmd, 1000*time.Millisecond)
	if err != nil {
		return err
	}
	nt.NetzteilBase.Ident = string(resp)
	return nil
}

func (nt *RND320) Status() (interface{}, error) {
	cmd := "STATUS?"
	resp, err := nt.request(cmd, 100*time.Millisecond)
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
	var cmd string
	if enabled {
		cmd = "OUT1"
	} else {
		cmd = "OUT0"
	}

	if err := nt.command(cmd); err != nil {
		return err
	}
	return nil
}

func (nt *RND320) SetBeep(enabled bool) error {
	return opennetzteil.ErrNotImplemented
}

func (nt *RND320) GetChannels() (int, error) {
	return 1, nil
}

func (nt *RND320) GetCurrent(channel int) (float64, error) {
	cmd := fmt.Sprintf("IOUT%d?", channel)
	resp, err := nt.request(cmd, 100*time.Millisecond)
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
	cmd := fmt.Sprintf("ISET%d:%.2f", channel, current)
	err := nt.command(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (nt *RND320) GetVoltage(channel int) (float64, error) {
	cmd := fmt.Sprintf("VOUT%d?", channel)
	resp, err := nt.request(cmd, 100*time.Millisecond)
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
	cmd := fmt.Sprintf("VSET%d:%.2f", channel, voltage)
	err := nt.command(cmd)
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
