package dvapi_model

import (
	"encoding/json"
	"time"
)

// / DeviceState is a helper just to make the code more idiomatic
// DeviceState.ToString allows DeviceState to be passed as a string as well

const (
	DeviceStateAvailable string = "available"
	DeviceStateInUse     string = "in-use"
	DeviceStateInactive  string = "inactive"
)

// / struct Device is the structure that holds the information for a single Device
type Device struct {
	ID        int       `json:"id,omitempty"`
	Name      string    `json:"name"`
	Brand     string    `json:"brand,omitempty"`
	State     string    `json:"state"`
	CreatedOn time.Time `json:"created_on"`
}

// / Devices is just a helper to use as a array of devices
type Devices []Device

// Device.FromJsonBytes unmarshals 'bytes []byte' into a Device struct
func (d *Device) FromJsonBytes(bytes []byte) (err error) {
	return json.Unmarshal(bytes, d)
}

// Device.ToJsonBytes marshals 'Device' into []byte bytes
func (d *Device) ToJsonBytes() (bytes []byte, err error) {
	return json.Marshal(d)
}

// Devices.FromJsonBytes marshals 'bytes []byte' into a Devices struct
func (da *Devices) FromJsonBytes(bytes []byte) (err error) {
	return json.Unmarshal(bytes, da)
}

// Device.ToJsonBytes marshals 'Devices' into []byte bytes
func (da *Devices) ToJsonBytes() (bytes []byte, err error) {
	return json.Marshal(da)
}
