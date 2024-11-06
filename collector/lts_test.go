package collector

import (
	"testing"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func TestLTSGetResourceInfo(t *testing.T) {

	patches := getPatches()
	defer patches.Reset()
	conf.DomainID = "123"
	logs.InitLog("")
	sysConfig := map[string][]string{
		"log_group_id":               {"packet_loss_rate_avg"},
		"log_group_id,log_stream_id": {"packet_loss_rate_avg"},
	}
	patches.ApplyFuncReturn(getMetricConfigMap, sysConfig)

	logGroupResp := ListLogGroupResponse{
		LogGroups: []LogGroup{
			{
				LogGroupName:  "test-group1",
				LogGroupID:    "xxxxx1",
				LogGroupAlias: "xxx-alias1",
			},
			{
				LogGroupName:  "test-group2",
				LogGroupID:    "xxxxx2",
				LogGroupAlias: "xxx-alias2",
			},
		},
		HttpStatusCode: 200,
	}
	logStreamResp := ListLogStreamResponse{
		LogStreams: []LogStream{
			{
				LogStreamName:      "xxx-name1",
				LogStreamNameAlias: "xxx-alias1",
				LogStreamID:        "xxxx1",
			},
			{
				LogStreamName:      "xxx-name2",
				LogStreamNameAlias: "xxx-alias2",
				LogStreamID:        "xxxx2",
			},
		},
		HttpStatusCode: 200,
	}
	patches.ApplyMethodFunc(getHcClient(getEndpoint("lts", "v2")), "Sync", func(req interface{}, reqDef *def.HttpRequestDef) (interface{}, error) {
		if reqDef.Path == "/v2/{project_id}/groups" {
			return &logGroupResp, nil
		}
		if reqDef.Path == "/v2/{project_id}/groups/xxxxx1/streams" {
			return &logStreamResp, nil
		}
		return nil, nil
	})

	var ltsInfo LTSInfo
	resourceInfo, metrics := ltsInfo.GetResourceInfo()
	assert.Equal(t, 4, len(resourceInfo))
	assert.Equal(t, 4, len(metrics))

}
