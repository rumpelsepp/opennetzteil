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

func (nt *NetzteilBase) RequestWithTimeout(handle io.ReadWriter, cmd []byte, timeout time.Duration) ([]byte, error) {
	nt.mutex.Lock()
	defer nt.mutex.Unlock()
	_, err := io.Copy(handle, bytes.NewReader(cmd))
	if err != nil {
		return nil, err
	}
	var (
		outBuf []byte
		n      = 0
		buf    = make([]byte, 32*1024)
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
		n, err = handle.Read(buf)
		outBuf = append(outBuf, buf[:n]...)
		if err != nil {
			if os.IsTimeout(err) {
				return outBuf, nil
			}
			return nil, err
		}
	}
}

// TCPSend creates a connection, sends a commend and closes the connection.
// Useful if the relevant powersupply only supports one TCP connection at a
// time. To avoid deadlocks a HTTP/1 pattern is used. One request maps to
// one TCP connection. A HTTP keep-alive equivalent is avail with
// TCPSendBatched().
func (nt *NetzteilBase) TCPSend(target, cmd string) error {
	conn, err := net.Dial("tcp", target)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nt.SendCommandLine(conn, []byte(cmd))
}

// TCPSendBatched creates a connection, sends multiple commands and
// closes the connection.
func (nt *NetzteilBase) TCPSendBatched(target string, cmd []string) error {
	conn, err := net.Dial("tcp", target)
	if err != nil {
		return err
	}
	defer conn.Close()
	for _, cmd := range cmd {
		err = nt.SendCommandLine(conn, []byte(cmd))
		if err != nil {
			return err
		}
	}
	return nil
}

// TCPRequest creates a connection, sends the command, reads
// back the response, and closes the connection.
func (nt *NetzteilBase) TCPRequest(target, cmd string) ([]byte, error) {
	conn, err := net.Dial("tcp", target)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return nt.RequestLine(conn, []byte(cmd))
}
