package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	rmsModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type GaussdbNodeInfo struct {
	ResourceBaseInfo
	NodeProperties
}

type NodeProperties struct {
	InstanceName string                   `json:"instance_name"`
	InstanceId   string                   `json:"instanceId"`
	InnerPort    string                   `json:"innerPort"`
	InnerIp      string                   `json:"innerIp"`
	Role         string                   `json:"role"`
	EngineName   string                   `json:"engineName"`
	Dimensions   []model.MetricsDimension `json:"dimensions"`
}

type DbProxyProperties struct {
	Role             string `json:"role"`
	IpAddress        string `json:"ipAddress"`
	MasterInstanceId string `json:"masterInstanceId"`
	EngineName       string `json:"enginName"`
	ProxyId          string `json:"proxyId"`
}

var gaussdbInfo serversInfo

type GAUSSDBInfo struct{}

func (getter GAUSSDBInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	gaussdbInfo.Lock()
	defer gaussdbInfo.Unlock()
	if gaussdbInfo.LabelInfo == nil || time.Now().Unix() > gaussdbInfo.TTL {
		// 获取gaussdbformysql资源标签和指标维度
		gaussdbSysConfigMap := getMetricConfigMap("SYS.GAUSSDB")
		nodes, instanceMap, err := getAllGaussdbNodesFromRMS()
		if err == nil {
			for _, node := range nodes {
				dim0Name := node.Dimensions[0].Name
				var metricNames []string
				ok := false
				if dim0Name == "gaussdb_mysql_ha_node_id" || dim0Name == "gaussdb_mysql_ha_id" {
					metricNames, ok = gaussdbSysConfigMap["gaussdb_mysql_ha_id,gaussdb_mysql_ha_node_id"]
				} else {
					metricNames, ok = gaussdbSysConfigMap["gaussdb_mysql_instance_id,gaussdb_mysql_node_id"]
				}
				if !ok {
					continue
				}
				metrics := buildDimensionMetrics(metricNames, "SYS.GAUSSDB", node.Dimensions)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"instanceName", "name", "epId", "innerPort", "innerIp", "role", "engineName"},
					Value: []string{node.InstanceName, node.Name, node.EpId, node.InnerPort, node.InnerIp, node.Role, node.EngineName},
				}
				keys, values := getTags(node.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		// 获取dbproxy资源标签和指标维度
		dbproxySysConfigMap := getMetricConfigMap("SYS.DBPROXY")
		dbProxyNodes, err := getAllDBProxyNodesFromRMS()
		if err == nil {
			for _, node := range dbProxyNodes {
				metricNames, ok := dbproxySysConfigMap["dbproxy_instance_id,dbproxy_node_id"]
				if !ok {
					continue
				}
				var properties DbProxyProperties
				err = fmtResourceProperties(node.Properties, &properties)
				if err != nil {
					logs.Logger.Errorf("fmt gaussdb db proxy node properties error: %s", err.Error())
					continue
				}
				instanceInfo, ok := instanceMap[properties.MasterInstanceId]
				if !ok {
					logs.Logger.Errorf("gaussdb db proxy node has no instance: %s", node.Name)
					continue
				}
				info := labelInfo{
					Name:  []string{"name", "instanceName", "instanceId", "epId", "role", "ipAddress", "engineName"},
					Value: []string{*node.Name, *instanceInfo.Name, properties.MasterInstanceId, *node.EpId, properties.Role, properties.IpAddress, properties.EngineName},
				}
				keys, values := getTags(node.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)

				metrics := buildDimensionMetrics(metricNames, "SYS.DBPROXY", []model.MetricsDimension{
					{
						Name:  "dbproxy_instance_id",
						Value: properties.ProxyId,
					},
					{
						Name:  "dbproxy_node_id",
						Value: *node.Id,
					},
				})
				filterMetrics = append(filterMetrics, metrics...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		gaussdbInfo.LabelInfo = resourceInfos
		gaussdbInfo.FilterMetrics = filterMetrics
		gaussdbInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return gaussdbInfo.LabelInfo, gaussdbInfo.FilterMetrics
}

func getAllDBProxyNodesFromRMS() ([]rmsModel.ResourceEntity, error) {
	resp, err := listResources("gaussdbformysql", "proxynodes")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of gaussdb.proxynodes, error: %s", err.Error())
		return nil, err
	}
	return resp, nil
}

func getAllGaussdbNodesFromRMS() ([]GaussdbNodeInfo, map[string]rmsModel.ResourceEntity, error) {
	resp, err := listResources("gaussdbformysql", "nodes")
	instanceResp, err := listResources("gaussdbformysql", "instance")

	instanceMap := make(map[string]rmsModel.ResourceEntity)
	for _, entity := range instanceResp {
		instanceMap[*entity.Id] = entity
	}
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of gaussdb.nodes, error: %s", err.Error())
		return nil, nil, err
	}
	nodes := make([]GaussdbNodeInfo, 0, len(resp))
	for _, resource := range resp {
		var properties NodeProperties
		err := fmtResourceProperties(resource.Properties, &properties)
		if err != nil {
			logs.Logger.Errorf("fmt gaussdb node properties error: %s", err.Error())
			continue
		}
		instanceResource, ok := instanceMap[properties.InstanceId]
		if ok {
			properties.InstanceName = *instanceResource.Name
		} else {
			logs.Logger.Errorf("Get gaussdb instance name empty, instance id is %s", properties.InstanceId)
		}
		nodes = append(nodes, GaussdbNodeInfo{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags,
			},
			NodeProperties: properties,
		})

	}
	return nodes, instanceMap, nil
}
