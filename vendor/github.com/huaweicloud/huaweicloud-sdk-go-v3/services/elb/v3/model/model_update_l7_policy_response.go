package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateL7PolicyResponse struct {
	// 请求ID。  注：自动生成 。

	RequestId *string `json:"request_id,omitempty"`

	L7policy       *L7Policy `json:"l7policy,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o UpdateL7PolicyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateL7PolicyResponse struct{}"
	}

	return strings.Join([]string{"UpdateL7PolicyResponse", string(data)}, " ")
}
