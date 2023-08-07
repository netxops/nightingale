package router

import (
	"encoding/json"
	"fmt"
	"github.com/ccfos/nightingale/v6/pushgw/labeler"
	"github.com/toolkits/pkg/logger"
)

func (rt *Router) EnrichLabelsFromRedis() map[string]labeler.DeviceTagPair {
	ct := rt.Ctx
	rds := rt.TargetCache.GetRedis()
	labelMap, err := rds.HGetAll(ct.GetContext(), labeler.DEVICE_TAG_REDIS_KEY).Result()
	if err != nil {
		logger.Errorf("get enrichLabels from redis has error: %v", err)
		return nil
	}

	logger.Info("获取redis数据，数据长度：", len(labelMap))
	if len(labelMap) == 0 {
		return nil
	}

	result := map[string]labeler.DeviceTagPair{}
	for _, v := range labelMap {
		dtp := labeler.DeviceTagPair{}
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

		result[dtp.IP] = dtp
	}

	return result
}
