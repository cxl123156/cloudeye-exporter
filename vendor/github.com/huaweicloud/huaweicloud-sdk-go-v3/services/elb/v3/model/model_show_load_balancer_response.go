package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowLoadBalancerResponse struct {
	// 请求ID。  注：自动生成 。

	RequestId *string `json:"request_id,omitempty"`

	Loadbalancer   *LoadBalancer `json:"loadbalancer,omitempty"`
	HttpStatusCode int           `json:"-"`
}

func (o ShowLoadBalancerResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowLoadBalancerResponse struct{}"
	}

	return strings.Join([]string{"ShowLoadBalancerResponse", string(data)}, " ")
}
