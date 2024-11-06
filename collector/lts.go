package collector

import (
	"errors"
	"fmt"
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

var ltsInfo serversInfo

type LTSInfo struct{}

type ListLogGroupRequest struct {
}

type ListLogStreamRequest struct {
	LogGroupID string `json:"log_group_id"`
}

type ListLogGroupResponse struct {
	LogGroups      []LogGroup `json:"log_groups"`
	HttpStatusCode int        `json:"-"`
}

type ListLogStreamResponse struct {
	LogStreams     []LogStream `json:"log_streams"`
	HttpStatusCode int         `json:"-"`
}

type LogGroup struct {
	LogGroupName  string            `json:"log_group_name"`
	Tag           map[string]string `json:"tag"`
	LogGroupID    string            `json:"log_group_id"`
	LogGroupAlias string            `json:"log_group_alias"`
}

type LogStream struct {
	LogStreamName      string            `json:"log_stream_name"`
	LogStreamNameAlias string            `json:"log_stream_name_alias"`
	LogStreamID        string            `json:"log_stream_id"`
	Tag                map[string]string `json:"tag"`
}

func (getter LTSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	filterMetrics := make([]model.MetricInfoList, 0, 0)
	resourceInfos := map[string]labelInfo{}

	ltsInfo.Lock()
	defer ltsInfo.Unlock()
	if ltsInfo.LabelInfo == nil || time.Now().Unix() > ltsInfo.TTL {
		logGroups, err := getAllLogGroup()
		if err != nil {
			logs.Logger.Errorf("Get all log group error, error is %s", err.Error())
			return nil, nil
		}
		buildLogGroupInfos(logGroups, &filterMetrics, resourceInfos)
		buildLogStreamInfos(logGroups, &filterMetrics, resourceInfos)

		ltsInfo.LabelInfo = resourceInfos
		ltsInfo.FilterMetrics = filterMetrics
		ltsInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return ltsInfo.LabelInfo, ltsInfo.FilterMetrics
}

func buildLogGroupInfos(logGroups []LogGroup, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	sysConfigMap := getMetricConfigMap("SYS.LTS")
	groupIDMetrics := sysConfigMap["log_group_id"]
	for _, group := range logGroups {
		metrics := buildSingleDimensionMetrics(groupIDMetrics, "SYS.LTS", "log_group_id", group.LogGroupID)
		*filterMetrics = append(*filterMetrics, metrics...)
		info := labelInfo{
			Name:  []string{"log_group_name"},
			Value: []string{group.LogGroupName},
		}
		keys, values := getTags(group.Tag)
		info.Name = append(info.Name, keys...)
		info.Value = append(info.Value, values...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
	}
}

func buildLogStreamInfos(logGroups []LogGroup, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	sysConfigMap := getMetricConfigMap("SYS.LTS")
	logStreamMetrics := sysConfigMap["log_group_id,log_stream_id"]
	for _, group := range logGroups {
		logStreams, err := getAllLogStream(group.LogGroupID)
		if err != nil {
			logs.Logger.Errorf("Get all log group error, error is %s", err.Error())
			return
		}

		for _, stream := range logStreams {
			metrics := buildDimensionMetrics(logStreamMetrics, "SYS.LTS", []model.MetricsDimension{
				{
					Name:  "log_group_id",
					Value: group.LogGroupID,
				},
				{
					Name:  "log_stream_id",
					Value: stream.LogStreamID,
				},
			})
			*filterMetrics = append(*filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"log_group_name", "log_stream_name"},
				Value: []string{group.LogGroupName, stream.LogStreamName},
			}
			keys, values := getTags(stream.Tag)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}

	}
}

func getAllLogGroup() ([]LogGroup, error) {
	requestDef := genDefaultReqDefWithOffsetAndLimit("/v2/{project_id}/groups", new(ListLogGroupResponse))
	request := ListLogGroupRequest{}
	resp, err := getHcClient(getEndpoint("lts", "v2")).Sync(request, requestDef)
	if err != nil {
		return nil, err
	}
	jobsInfo, ok := resp.(*ListLogGroupResponse)
	if !ok {
		return nil, errors.New("resp type is not ListFlinkJobsResponse")
	}
	return jobsInfo.LogGroups, nil
}

func getAllLogStream(logGroupID string) ([]LogStream, error) {
	uri := fmt.Sprintf("/v2/{project_id}/groups/%s/streams", logGroupID)
	requestDef := genDefaultReqDefWithOffsetAndLimit(uri, new(ListLogStreamResponse))
	request := ListLogStreamRequest{
		LogGroupID: logGroupID,
	}
	resp, err := getHcClient(getEndpoint("lts", "v2")).Sync(request, requestDef)
	if err != nil {
		return nil, err
	}
	jobsInfo, ok := resp.(*ListLogStreamResponse)
	if !ok {
		return nil, errors.New("resp type is not ListFlinkJobsResponse")
	}
	return jobsInfo.LogStreams, nil
}
