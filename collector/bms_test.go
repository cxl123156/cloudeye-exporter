package collector

import (
	"testing"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func TestBmsGetResourceInfo(t *testing.T) {
	instances := []EcsInstancesInfo{
		{ResourceBaseInfo: ResourceBaseInfo{ID: "0001-0001-000000001", Name: "host01", EpId: "0"}},
	}
	metricConf = map[string]MetricConf{
		"SYS.BMS": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"instance_id": {"cpu_utils"},
			},
		},
	}

	patches := getPatches()
	logs.InitLog("")
	patches.ApplyFuncReturn(getAllServerFromRMS, instances, nil)
	patches.ApplyFuncReturn(getIPFromEcsInfo, "", "", "")
	defer patches.Reset()

	var bmsGetter BMSInfo
	labels, metrics := bmsGetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))

	var servicesGetter SERVICEBMSInfo
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{
		{
			Dimensions: []model.MetricsDimension{
				{
					Name:  "instance_id",
					Value: "111111",
				},
			},
			Namespace:  "SERVICE.BMS",
			MetricName: "cpu_usage",
		},
	}, nil)

	serviceLabel, _ := servicesGetter.GetResourceInfo()
	assert.Equal(t, 1, len(serviceLabel))
}

func TestGetEvsInfoForBMS(t *testing.T) {
	testInstanceID := "test-instance-id"
	testDiskName := "vda"
	testEvsID := "evs-12345"

	// 初始化日志和 patches
	patches := getPatches()
	logs.InitLog("")
	// mock loadAgentDimensions 避免调用 CES client
	patches.ApplyFunc(loadAgentDimensions, func(instanceID string) error { return nil })
	defer patches.Reset()

	// 清理 agentDimensions 中测试添加的 key
	defer func() {
		agentDimensions.Delete(testDiskName)
	}()

	// 测试用例1: 正常获取 evsId
	t.Run("normal case: get evsId successfully", func(t *testing.T) {
		// 重置 bmsInfo
		bmsInfo = serversInfo{}
		bmsInfo.ExtendInfo = map[string]interface{}{
			testInstanceID: map[string]string{
				testDiskName: testEvsID,
			},
		}

		// 预先存储 agentDimensions，避免调用 loadAgentDimensions
		agentDimensions.Store(testDiskName, testDiskName)

		namespace := "SERVICE.BMS"
		metric := model.BatchMetricData{
			Namespace: &namespace,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: testInstanceID},
				{Name: "disk", Value: testDiskName},
			},
		}

		label := labelInfo{}
		getEvsInfoForBMS(metric, &label)

		assert.Contains(t, label.Name, "evsId")
		evsIdx := -1
		for i, name := range label.Name {
			if name == "evsId" {
				evsIdx = i
				break
			}
		}
		assert.Equal(t, testEvsID, label.Value[evsIdx])
	})

	// 测试用例2: diskName 为空
	t.Run("diskName is empty", func(t *testing.T) {
		bmsInfo = serversInfo{}
		bmsInfo.ExtendInfo = map[string]interface{}{
			testInstanceID: map[string]string{
				testDiskName: testEvsID,
			},
		}

		namespace := "SERVICE.BMS"
		metric := model.BatchMetricData{
			Namespace: &namespace,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: testInstanceID},
				{Name: "disk", Value: ""},
			},
		}

		label := labelInfo{}
		getEvsInfoForBMS(metric, &label)

		assert.NotContains(t, label.Name, "evsId")
	})

	// 测试用例3: instanceID 不在 ExtendInfo 中
	t.Run("instanceID not found in ExtendInfo", func(t *testing.T) {
		bmsInfo = serversInfo{}
		bmsInfo.ExtendInfo = map[string]interface{}{
			"other-instance": map[string]string{
				testDiskName: testEvsID,
			},
		}

		agentDimensions.Store(testDiskName, testDiskName)

		namespace := "SERVICE.BMS"
		metric := model.BatchMetricData{
			Namespace: &namespace,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: testInstanceID},
				{Name: "disk", Value: testDiskName},
			},
		}

		label := labelInfo{}
		getEvsInfoForBMS(metric, &label)

		// instanceID 不存在时，直接返回，不添加 evsId
		assert.NotContains(t, label.Name, "evsId")
	})

	// 测试用例4: extendInfoMap 转换为 map[string]string 失败
	t.Run("extendInfoMap convert failed", func(t *testing.T) {
		bmsInfo = serversInfo{}
		bmsInfo.ExtendInfo = map[string]interface{}{
			testInstanceID: "not a map", // 错误的类型
		}

		agentDimensions.Store(testDiskName, testDiskName)

		namespace := "SERVICE.BMS"
		metric := model.BatchMetricData{
			Namespace: &namespace,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: testInstanceID},
				{Name: "disk", Value: testDiskName},
			},
		}

		label := labelInfo{}
		getEvsInfoForBMS(metric, &label)

		// 类型转换失败，不添加 evsId
		assert.NotContains(t, label.Name, "evsId")
	})

	// 测试用例5: diskName 不在 map 中
	t.Run("diskName not found in map", func(t *testing.T) {
		bmsInfo = serversInfo{}
		bmsInfo.ExtendInfo = map[string]interface{}{
			testInstanceID: map[string]string{
				"other-disk": testEvsID,
			},
		}

		agentDimensions.Store(testDiskName, testDiskName)

		namespace := "SERVICE.BMS"
		metric := model.BatchMetricData{
			Namespace: &namespace,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: testInstanceID},
				{Name: "disk", Value: testDiskName},
			},
		}

		label := labelInfo{}
		getEvsInfoForBMS(metric, &label)

		// diskName 不存在时，添加 evsId 但值为空
		assert.Contains(t, label.Name, "evsId")
		evsIdx := -1
		for i, name := range label.Name {
			if name == "evsId" {
				evsIdx = i
				break
			}
		}
		assert.Equal(t, "", label.Value[evsIdx])
	})
}
