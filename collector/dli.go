package collector

import (
	"errors"
	"fmt"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var dliInfo serversInfo

type DLIInfo struct{}

func (getter DLIInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	dliInfo.Lock()
	defer dliInfo.Unlock()
	if dliInfo.LabelInfo == nil || time.Now().Unix() > dliInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.DLI")

		// queues
		buildQueuesInfo(sysConfigMap, &filterMetrics, resourceInfos)

		// flink jobs
		buildFlinkJobsInfo(sysConfigMap, &filterMetrics, resourceInfos)

		// elastic resource pool
		buildElasticPool(sysConfigMap, &filterMetrics, resourceInfos)

		dliInfo.LabelInfo = resourceInfos
		dliInfo.FilterMetrics = filterMetrics
		dliInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return dliInfo.LabelInfo, dliInfo.FilterMetrics
}

func buildElasticPool(configMap map[string][]string, filterMetrics *[]model.MetricInfoList, infos map[string]labelInfo) {
	elasticPoolsMetricNames, ok := configMap["elastic_resource_pool_id"]
	if !ok {
		logs.Logger.Warnf("metric config is empty of elastic_resource_pool_id")
		return
	}

	elasticPoolRes, err := getElasticPoolFromDLI()
	if err != nil {
		logs.Logger.Errorf("Get all elastic pool error: %s", err.Error())
		return
	}

	for _, pool := range elasticPoolRes {
		metrics := buildSingleDimensionMetrics(elasticPoolsMetricNames, "SYS.DLI", "elastic_resource_pool_id", pool.ID)
		*filterMetrics = append(*filterMetrics, metrics...)
		info := labelInfo{
			Name:  []string{"name", "epId"},
			Value: []string{pool.Name, pool.EpId},
		}
		keys, values := getTags(pool.Tags)
		info.Name = append(info.Name, keys...)
		info.Value = append(info.Value, values...)
		infos[GetResourceKeyFromMetricInfo(metrics[0])] = info
	}
}

func buildQueuesInfo(sysConfigMap map[string][]string, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	queueMetricNames, ok := sysConfigMap["queue_id"]
	if !ok {
		logs.Logger.Warnf("metric config is empty of queue_id")
		return
	}
	queues, err := getQueuesFromRMS()
	if err != nil {
		logs.Logger.Errorf("Get all dli queues: %s", err.Error())
		return
	}
	for _, queue := range queues {
		metrics := buildSingleDimensionMetrics(queueMetricNames, "SYS.DLI", "queue_id", queue.ID)
		*filterMetrics = append(*filterMetrics, metrics...)
		info := labelInfo{
			Name:  []string{"name", "epId"},
			Value: []string{queue.Name, queue.EpId},
		}
		keys, values := getTags(queue.Tags)
		info.Name = append(info.Name, keys...)
		info.Value = append(info.Value, values...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
	}
}

func buildFlinkJobsInfo(sysConfigMap map[string][]string, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	jobMetricNames, ok := sysConfigMap["flink_job_id"]
	if !ok {
		logs.Logger.Warnf("metric config is empty of flink_job_id")
		return
	}
	jobs, err := getAllFlinkJobsInfo()
	if err != nil {
		logs.Logger.Errorf("Get all dli flink job: %s", err.Error())
		return
	}
	for _, job := range jobs {
		metrics := buildSingleDimensionMetrics(jobMetricNames, "SYS.DLI", "flink_job_id", job.ID)
		*filterMetrics = append(*filterMetrics, metrics...)
		info := labelInfo{
			Name:  []string{"name", "job_type"},
			Value: []string{job.Name, job.JobType},
		}
		keys, values := getTags(job.Tags)
		info.Name = append(info.Name, keys...)
		info.Value = append(info.Value, values...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
	}
}

func getQueuesFromRMS() ([]ResourceBaseInfo, error) {
	resp, err := listResources("dli", "queues")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of %s.%s, error: %s", "dli", "queues", err.Error())
		return nil, err
	}
	services := make([]ResourceBaseInfo, len(resp))
	for index, resource := range resp {
		var queueProperties DLIQueueProperties
		err := fmtResourceProperties(resource.Properties, &queueProperties)
		if err != nil {
			logs.Logger.Errorf("Fmt dli properties error: %s", err.Error())
			continue
		}
		services[index].ID = fmt.Sprintf("%d", queueProperties.QueueId)
		services[index].Name = *resource.Name
		services[index].EpId = *resource.EpId
		services[index].Tags = resource.Tags
	}
	return services, nil
}

type ListFlinkJobsRequest struct {
	Offset *int32 `json:"offset,omitempty"`
	Limit  *int32 `json:"limit,omitempty"`
}

type ListFlinkJobsResponse struct {
	IsSuccess      string  `json:"is_success"`
	Message        string  `json:"message"`
	JobList        JobList `json:"job_list"`
	HttpStatusCode int     `json:"-"`
}
type Jobs struct {
	JobID              int    `json:"job_id"`
	Name               string `json:"name"`
	Desc               string `json:"desc"`
	UserName           string `json:"user_name"`
	JobType            string `json:"job_type"`
	Status             string `json:"status"`
	StatusDesc         string `json:"status_desc"`
	CreateTime         int64  `json:"create_time"`
	Duration           int    `json:"duration"`
	RootID             int    `json:"root_id"`
	GraphEditorEnabled bool   `json:"graph_editor_enabled"`
	HasSavepoint       bool   `json:"has_savepoint"`
}
type JobList struct {
	TotalCount int    `json:"total_count"`
	Jobs       []Jobs `json:"jobs"`
}

type FlinkJobsInfo struct {
	ResourceBaseInfo
	JobType string
}

type ElasticPool struct {
	EpID     string `json:"enterprise_project_id"`
	ID       int    `json:"id"`
	PoolName string `json:"elastic_resource_pool_name"`
}

type ElasticPoolResponse struct {
	IsSuccess      bool          `json:"is_success"`
	Message        string        `json:"message"`
	Count          int           `json:"count"`
	ElasticPools   []ElasticPool `json:"elastic_resource_pools"`
	HttpStatusCode int           `json:"-"`
}

type DLIQueueProperties struct {
	QueueId int `json:"queue_id"`
}

func getAllFlinkJobsInfo() ([]FlinkJobsInfo, error) {
	var jobs []FlinkJobsInfo
	limit := int32(100)
	offset := int32(0)
	request := &ListFlinkJobsRequest{Limit: &limit, Offset: &offset}
	requestDef := genDefaultReqDefWithOffsetAndLimit("/v1.0/{project_id}/streaming/jobs", new(ListFlinkJobsResponse))
	for {
		resp, err := getHcClient(getEndpoint("dli", "v1.0")).Sync(request, requestDef)
		if err != nil {
			return nil, err
		}
		jobsInfo, ok := resp.(*ListFlinkJobsResponse)
		if !ok {
			return nil, errors.New("resp type is not ListFlinkJobsResponse")
		}
		if len(jobsInfo.JobList.Jobs) == 0 {
			break
		}
		for _, job := range jobsInfo.JobList.Jobs {
			jobs = append(jobs, FlinkJobsInfo{
				ResourceBaseInfo: ResourceBaseInfo{ID: fmt.Sprintf("%d", job.JobID), Name: job.Name},
				JobType:          job.JobType,
			})
		}
		*request.Offset += limit
	}
	return jobs, nil
}

func getElasticPoolFromDLI() ([]ResourceBaseInfo, error) {
	var jobs []ResourceBaseInfo
	limit := int32(100)
	offset := int32(0)
	request := &ListFlinkJobsRequest{Limit: &limit, Offset: &offset}
	requestDef := genDefaultReqDefWithOffsetAndLimit("/v3/{project_id}/elastic-resource-pools", new(ElasticPoolResponse))
	logs.Logger.Infof("get all pools req is %s", requestDef)
	for {
		resp, err := getHcClient(getEndpoint("dli", "v3")).Sync(request, requestDef)
		logs.Logger.Infof("get all pools finish %s", resp)
		if err != nil {
			return nil, err
		}
		poolsInfo, ok := resp.(*ElasticPoolResponse)
		if !ok {
			return nil, errors.New("resp type is not ElasticPoolResponse")
		}
		if len(poolsInfo.ElasticPools) == 0 {
			break
		}
		for _, pool := range poolsInfo.ElasticPools {
			logs.Logger.Infof("current pool is %s", pool)
			jobs = append(jobs, ResourceBaseInfo{
				ID: fmt.Sprintf("%d", pool.ID), Name: pool.PoolName, EpId: pool.EpID,
			})
		}
		*request.Offset += limit
	}
	return jobs, nil
}
