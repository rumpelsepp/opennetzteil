package dummy

import "codeberg.org/rumpelsepp/opennetzteil"

type DummyDevice struct {
	opennetzteil.NetzteilBase
}

func (d *DummyDevice) Probe() error {
	return nil
}

func (d *DummyDevice) Status() (interface{}, error) {
	return nil, nil
}

func (d *DummyDevice) GetMaster() (bool, error) {
	return true, nil
}

func (d *DummyDevice) SetMaster(enabled bool) error {
	return nil
}

func (d *DummyDevice) SetBeep(enabled bool) error {
	return nil
}

func (d *DummyDevice) GetChannels() (int, error) {
	return 1, nil
}

func (d *DummyDevice) GetCurrent(channel int) (float64, error) {
	return 12, nil
}
func (d *DummyDevice) SetOut(channel int, enabled bool) error {
	return nil
}

func (d *DummyDevice) GetOut(channel int) (bool, error) {
	return true, nil
}

func (d *DummyDevice) SetCurrent(channel int, current float64) error {
	return nil
}

func (d *DummyDevice) GetVoltage(channel int) (float64, error) {
	return 15, nil
}

func (d *DummyDevice) SetVoltage(channel int, voltage float64) error {
	return nil
}
func (d *DummyDevice) GetOCP(channel int) (bool, error) {
	return true, nil
}

func (d *DummyDevice) SetOCP(channel int, enabled bool) error {
	return nil
}
func (d *DummyDevice) GetOVP(channel int) (bool, error) {
	return true, nil
}

func (d *DummyDevice) SetOVP(channel int, enabled bool) error {
	return nil
}
