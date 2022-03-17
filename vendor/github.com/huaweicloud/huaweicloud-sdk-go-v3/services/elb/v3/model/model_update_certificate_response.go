package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateCertificateResponse struct {
	// 请求ID。  注：自动生成 。

	RequestId *string `json:"request_id,omitempty"`

	Certificate    *CertificateInfo `json:"certificate,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o UpdateCertificateResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateCertificateResponse struct{}"
	}

	return strings.Join([]string{"UpdateCertificateResponse", string(data)}, " ")
}
