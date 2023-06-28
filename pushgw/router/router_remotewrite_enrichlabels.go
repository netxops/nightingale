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
		//实际匹配到ident
		if strings.Contains(ident, v.IP) {
			target, exist := rt.TargetCache.Get(ident)
			if !exist {
				logger.Errorf(fmt.Sprintf("not found target[%s] device[%s]", ident, v.DeviceName))
				return
			}
			for _, tag := range v.Tags {
				target.TagsMap[tag.Dimension] = tag.TagName
			}
		}
	}
}

func (rt *Router) EnrichLabelsFromRedis() map[string]DeviceTagPair {
	ct := rt.Ctx
	rds := rt.TargetCache.GetRedis()
	labelMap, err := rds.HGetAll(ct.GetContext(), DEVICE_TAG_REDIS_KEY).Result()
	if err != nil {
		logger.Errorf("get enrichLabels from redis has error", err)
		return nil
	}

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

	return result
}
