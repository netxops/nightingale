package router

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prometheus/prometheus/prompb"
	"github.com/toolkits/pkg/logger"
)

const DEVICE_TAG_REDIS_KEY = "DEVICE_TAG_REDIS_KEY"
const (
	IDENT       SpecifyLabel = "ident"
	TARGET      SpecifyLabel = "target"
	IF_DESCR    SpecifyLabel = "ifDescr"
	TO_DEVICE   SpecifyLabel = "toDevice"
	PEER_DEVICE SpecifyLabel = "peerDevice"
	PEER_PORT   SpecifyLabel = "peerPort"
)

var REDIS_TAGS map[string]DeviceTagPair

type SpecifyLabel string

type DeviceTagPair struct {
	ResourceKey string       `json:"resource_key"`
	ResourceId  string       `json:"resource_id"`
	DeviceName  string       `json:"device_name"`
	IPV4        string       `json:"ipv4"`
	IPV6        string       `json:"ipv6"`
	IP          string       `json:"ip"`
	Tags        []*DeviceTag `json:"tags"`
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

//func (rt *Router) remakeWriteRemoteEnrichLabels(pt *prompb.TimeSeries) {
//	if identVal, identExist := Has(string(IDENT), pt); identExist {
//		peerRelationBuilder := labeler.PeerRelationLabelBuilder{}
//		peerRelationBuilder.Build(identVal+"/24", pt)
//	}
//
//	if targetVal, targetExist := Has(string(TARGET), pt); targetExist {
//		targetBuilder := labeler.TargetLabelBuilder{}
//		targetBuilder.Build(targetVal+"/24", pt)
//	}
//}

func (rt *Router) EnrichLabelsFromRedis() map[string]DeviceTagPair {
	ct := rt.Ctx
	rds := rt.TargetCache.GetRedis()
	labelMap, err := rds.HGetAll(ct.GetContext(), DEVICE_TAG_REDIS_KEY).Result()
	if err != nil {
		logger.Errorf("get enrichLabels from redis has error: %v", err)
		return nil
	}

	logger.Info("获取redis数据，数据长度：", len(labelMap))
	if len(labelMap) == 0 {
		return nil
	}

	result := map[string]DeviceTagPair{}
	for _, v := range labelMap {
		dtp := DeviceTagPair{}
		if err = json.Unmarshal([]byte(v), &dtp); err != nil {
			logger.Errorf("parse DeviceTagPair error: %v", err)
			continue
		}
		if dtp.IPV4 != "" && dtp.IPV6 != "" {
			logger.Warning(fmt.Sprintf("parse DeviceTagPair error: %s has mutil ip address", dtp.DeviceName))
			continue
		}
		if dtp.IPV4 != "" {
			dtp.IP = dtp.IPV4
		}
		if dtp.IPV6 != "" {
			dtp.IP = dtp.IPV6
		}
		if dtp.IP == "" {
			logger.Warning(fmt.Sprintf("%s has not ip address", dtp.DeviceName))
			continue
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
			case "toDevice":
				tag.TagLabel = "toDevice"
			default:
				break
			}
		}
		result[dtp.IP] = dtp
	}

	return result
}

func Has(key string, pt *prompb.TimeSeries) (keyValue string, matched bool) {
	for _, v := range pt.Labels {
		if strings.Contains(v.Name, key) {
			return v.Value, true
		}
	}
	return
}
