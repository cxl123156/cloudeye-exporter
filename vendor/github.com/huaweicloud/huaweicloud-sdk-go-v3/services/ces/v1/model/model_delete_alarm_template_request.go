package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type DeleteAlarmTemplateRequest struct {
	// 发送的实体的MIME类型。推荐用户默认使用application/json，如果API是对象、镜像上传等接口，媒体类型可按照流类型的不同进行确定。

	ContentType string `json:"Content-Type"`
	// 需要删除的自定义告警模板ID。

	TemplateId string `json:"template_id"`
}

func (o DeleteAlarmTemplateRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteAlarmTemplateRequest struct{}"
	}

	return strings.Join([]string{"DeleteAlarmTemplateRequest", string(data)}, " ")
}
