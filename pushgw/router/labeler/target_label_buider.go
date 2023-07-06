package labeler

import (
	"github.com/ccfos/nightingale/v6/pushgw/router"
	"github.com/prometheus/prometheus/prompb"
	"strings"
)

type TargetLabelBuilder struct{}

func (t TargetLabelBuilder) MatchDeviceTagPair(ipAddress string) (matched bool, dtp router.DeviceTagPair) {
	for _, v := range router.REDIS_TAGS {
		if strings.Contains(ipAddress, v.IP) {
			matched = true
			dtp = v
			return
		}
	}
	return
}

func (t TargetLabelBuilder) Build(ipAddress string, pt *prompb.TimeSeries) {
	matched, dtp := t.MatchDeviceTagPair(ipAddress)
	if !matched {
		return
	}
	targetPt := ""
	for _, v := range pt.Labels {
		if v.Name == string(router.TARGET) {
			targetPt = v.Value
			break
		}
	}

	if targetPt == "" {
		return
	}

	if targetPt == ipAddress {
		for _, dt := range dtp.Tags {
			if dt.Dimension != string(router.TO_DEVICE) && validdateDuplication(dt.TagLabel, pt) {
				label := prompb.Label{Name: dt.TagLabel, Value: dt.TagName}
				pt.Labels = append(pt.Labels, &label)
			}
		}
	}
}

func validdateDuplication(key string, pt *prompb.TimeSeries) bool {
	for _, v := range pt.Labels {
		if v.Name == key {
			return false
		}
	}
	return true
}
