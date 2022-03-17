package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ResetPwdRequest struct {
	// 语言

	XLanguage *string `json:"X-Language,omitempty"`
	// 实例ID，严格匹配UUID规则。

	InstanceId string `json:"instance_id"`

	Body *PwdResetRequest `json:"body,omitempty"`
}

func (o ResetPwdRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetPwdRequest struct{}"
	}

	return strings.Join([]string{"ResetPwdRequest", string(data)}, " ")
}
