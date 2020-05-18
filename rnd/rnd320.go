package rnd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"git.sr.ht/~rumpelsepp/opennetzteil"
	"git.sr.ht/~rumpelsepp/opennetzteil/serial"
	"git.sr.ht/~rumpelsepp/rlog"
)

// TODO: This power supply does not have any delimiters
// on the wire protocol. Thus, we must set a deadline or
// something. Otherwise the application can deadlock.

// TODO: mutex

type RND320 struct {
	opennetzteil.NetzteilBase
	terminal io.ReadWriteCloser
	writer   *bufio.Writer
	reader   *bufio.Reader
	ident    string
}

func NewRND320(path string) *RND320 {
	options := serial.OpenOptions{
		PortName:        path,
		BaudRate:        19200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	port, err := serial.Open(options)
	if err != nil {
		panic(err)
	}

	return &RND320{
		terminal: port,
		writer:   bufio.NewWriter(port),
		reader:   bufio.NewReader(port),
	}
}

func (nt *RND320) Probe() error {
	var (
		cmd = []byte("*IDN?")
		buf = make([]byte, 20) // This power supply is so broken…
	)
	err := nt.SendCommand(nt.writer, cmd)
	if err != nil {
		return err
	}
	if _, err := io.ReadFull(nt.reader, buf); err != nil {
		rlog.Info(err)
		return err
	}
	if err != nil {
		return err
	}

	nt.ident = string(buf)

	return nil
}

func (nt *RND320) GetMaster() (bool, error) {
	return true, nil
}

func (nt *RND320) SetMaster(enabled bool) error {
	var cmd []byte
	if enabled {
		cmd = []byte("OUT1")
	} else {
		cmd = []byte("OUT0")
	}

	if err := nt.SendCommand(nt.writer, cmd); err != nil {
		return err
	}
	return nil
}

func (nt *RND320) GetIdent() (string, error) {
	return nt.ident, nil
}

func (nt *RND320) SetBeep(enabled bool) error {
	return nil
}

func (nt *RND320) GetChannels() ([]int, error) {
	return []int{1}, nil
}

func (nt *RND320) GetCurrent(channel int) (float64, error) {
	cmd := []byte(fmt.Sprintf("IOUT%d?", channel))
	err := nt.SendCommand(nt.writer, cmd)
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 64)
	n, err := nt.reader.Read(buf)
	if err != nil {
		return 0, err
	}

	current, err := strconv.ParseFloat(string(buf[:n]), 32)
	if err != nil {
		return 0, err
	}
	return current, nil
}

func (nt *RND320) SetCurrent(channel int, current float64) error {
	cmd := []byte(fmt.Sprintf("ISET%d:%.2f", channel, current))
	err := nt.SendCommand(nt.writer, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (nt *RND320) GetVoltage(channel int) (float64, error) {
	cmd := []byte(fmt.Sprintf("VOUT%d?", channel))
	err := nt.SendCommand(nt.writer, cmd)
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 64)
	n, err := nt.reader.Read(buf)
	if err != nil {
		return 0, err
	}

	voltage, err := strconv.ParseFloat(string(buf[:n]), 32)
	if err != nil {
		return 0, err
	}
	return voltage, nil
}

func (nt *RND320) SetVoltage(channel int, voltage float64) error {
	cmd := []byte(fmt.Sprintf("VSET%d:%.2f", channel, voltage))
	err := nt.SendCommand(nt.writer, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (nt *RND320) GetOut(channel int) (bool, error) {
	return true, nil
}

func (nt *RND320) SetOut(channel int, enabled bool) error {
	return nil
}

func (nt *RND320) GetOCP(channel int) (bool, error) {
	return true, nil
}

func (nt *RND320) SetOCP(channel int, enabled bool) error {
	return nil
}

func (nt *RND320) GetOVP(channel int) (bool, error) {
	return true, nil
}

func (nt *RND320) SetOVP(channel int, enabled bool) error {
	return nil
}
