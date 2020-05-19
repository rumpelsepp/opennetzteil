package opennetzteil

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

var (
	ErrNotImplemented = errors.New("endpoint not implemented")
)

type Netzteil interface {
	Probe() error
	Status() (interface{}, error)
	GetMaster() (bool, error)
	SetMaster(enabled bool) error
	GetIdent() (string, error)
	SetBeep(enabled bool) error
	GetChannels() (int, error)
	GetCurrent(channel int) (float64, error)
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

func (nt *NetzteilBase) SendCommand(handle io.Writer, cmd []byte) error {
	nt.mutex.Lock()
	defer nt.mutex.Unlock()
	_, err := io.Copy(handle, bytes.NewReader(cmd))
	if err != nil {
		return err
	}
	return nil
}

func (nt *NetzteilBase) SendCommandLine(handle io.Writer, cmd []byte) error {
	return nt.SendCommand(handle, append(cmd, '\n'))
}

func (nt *NetzteilBase) RequestWithTimeout(handle io.ReadWriter, cmd []byte, timeout time.Duration) ([]byte, error) {
	nt.mutex.Lock()
	defer nt.mutex.Unlock()
	_, err := io.Copy(handle, bytes.NewReader(cmd))
	if err != nil {
		return nil, err
	}
	var (
		n    = 0
		read = 0
		buf  = make([]byte, 4*1024)
	)
	for {
		dl := time.Now().Add(timeout)
		switch r := handle.(type) {
		case *os.File:
			err = r.SetReadDeadline(dl)
			if err != nil {
				return nil, err
			}
		case net.Conn:
			err = r.SetReadDeadline(dl)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported reader: %t", r)
		}
		n, err = handle.Read(buf[read:])
		read += n
		if err != nil {
			if os.IsTimeout(err) {
				return buf[:read], nil
			}
			return nil, err
		}
	}
}

func (nt *NetzteilBase) Request(handle io.ReadWriter, cmd []byte) ([]byte, error) {
	nt.mutex.Lock()
	defer nt.mutex.Unlock()
	_, err := io.Copy(handle, bytes.NewReader(cmd))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, handle)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (nt *NetzteilBase) RequestLine(handle io.ReadWriter, cmd []byte) ([]byte, error) {
	nt.mutex.Lock()
	defer nt.mutex.Unlock()
	_, err := io.Copy(handle, bytes.NewReader(append(cmd, '\n')))
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(handle)
	line, _, err := reader.ReadLine()
	if err != nil {
		return nil, err
	}
	return line, nil
}
