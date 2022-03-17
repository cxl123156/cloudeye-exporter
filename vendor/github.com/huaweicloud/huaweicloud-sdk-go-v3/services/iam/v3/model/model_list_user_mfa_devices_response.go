package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListUserMfaDevicesResponse struct {
	// 虚拟MFA设备信息列表。

	VirtualMfaDevices *[]MfaDeviceResult `json:"virtual_mfa_devices,omitempty"`
	HttpStatusCode    int                `json:"-"`
}

func (o ListUserMfaDevicesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListUserMfaDevicesResponse struct{}"
	}

	return strings.Join([]string{"ListUserMfaDevicesResponse", string(data)}, " ")
}
