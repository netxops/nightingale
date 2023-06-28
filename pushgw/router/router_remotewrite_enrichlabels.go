package router

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/prometheus/prompb"
	"github.com/toolkits/pkg/logger"
	"strings"
)

const DEVICE_TAG_REDIS_KEY = "DEVICE_TAG_REDIS_KEY"

var REDIS_TAGS map[string]DeviceTagPair

type DeviceTagPair struct {
	ResourceKey string      `json:"resource_key"`
	ResourceId  string      `json:"resource_id"`
	DeviceName  string      `json:"device_name"`
	IPV4        string      `json:"ipv4"`
	IPV6        string      `json:"ipv6"`
	IP          string      `json:"ip"`
	Tags        []DeviceTag `json:"tags"`
}

type DeviceTag struct {
	Dimension string `json:"dimension"`
	TagLabel  string `json:"tag_label"`
	TagName   string `json:"tag_name"`
}

func (dtp DeviceTagPair) MarshalBinary() (data []byte, err error) {
	return json.Marshal(dtp)
}

func (dtp *DeviceTagPair) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, dtp)
}

func (rt *Router) remakeWriteRemoteEnrichLabels(pt *prompb.TimeSeries) {
	fmt.Println("======开始执行remakeWriteRemoteEnrichLabels=======")
	for _, v := range REDIS_TAGS {
		ident := ""
		for i := 0; i < len(pt.Labels); i++ {
			if pt.Labels[i].Name == "ident" {
				ident = pt.Labels[i].Value
				break
			}
		}
		if ident == "" {
			break
		}
		fmt.Println("当前ident值为：", ident)
		//实际匹配到ident
		if strings.Contains(ident, v.IP) {
			target, exist := rt.TargetCache.Get(ident)
			if !exist {
				logger.Errorf(fmt.Sprintf("not found target[%s] device[%s]", ident, v.DeviceName))
				return
			}
			fmt.Println("匹配到相应ident：", ident)
			for _, tag := range v.Tags {
				target.TagsMap[tag.Dimension] = tag.TagName
			}
			fmt.Println("当前ident的扩展标签为：", target.TagsMap)
		}
	}
}

func (rt *Router) EnrichLabelsFromRedis() map[string]DeviceTagPair {
	fmt.Println("======开始执行EnrichLabelsFromRedis=====")
	ct := rt.Ctx
	rds := rt.TargetCache.GetRedis()
	labelMap, err := rds.HGetAll(ct.GetContext(), DEVICE_TAG_REDIS_KEY).Result()
	if err != nil {
		logger.Errorf("get enrichLabels from redis has error", err)
		return nil
	}

	fmt.Println("获取redis数据：", labelMap)
	if len(labelMap) == 0 {
		return nil
	}

	result := map[string]DeviceTagPair{}
	for _, v := range labelMap {
		dtp := DeviceTagPair{}
		if err = json.Unmarshal([]byte(v), &dtp); err != nil {
			logger.Errorf("parse DeviceTagPair error:", err)
			return nil
		}
		if dtp.IPV4 != "" && dtp.IPV6 != "" {
			logger.Errorf(fmt.Sprintf("parse DeviceTagPair error: %s has mutil ip address", dtp.DeviceName))
			return nil
		}
		if dtp.IPV4 != "" {
			dtp.IP = dtp.IPV4
		}
		if dtp.IPV6 != "" {
			dtp.IP = dtp.IPV6
		}

		for _, tag := range dtp.Tags {
			switch tag.Dimension {
			case "设备":
				tag.TagLabel = "currentDevice"
			case "环境":
				tag.TagLabel = "env"
			case "区域":
				tag.TagLabel = "region"
			case "机房":
				tag.TagLabel = "machineRoom"
			case "产品线":
				tag.TagLabel = "productLine"
			case "云平台":
				tag.TagLabel = "cloudPlatform"
			}
		}
		result[dtp.IP] = dtp
	}

	fmt.Println("输出map DeviceTagPair： ", result)
	return result
}
