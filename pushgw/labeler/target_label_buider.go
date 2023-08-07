package labeler

import (
	"github.com/prometheus/prometheus/prompb"
)

type TargetLabelBuilder struct{}

func (t TargetLabelBuilder) MatchDeviceTagPair(ipAddress string) (matched bool, dtp DeviceTagPair) {
	for _, v := range REDIS_TAGS {
		if ipAddress == v.IP {
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
		if v.Name == string(TARGET) {
			targetPt = v.Value
			break
		}
	}

	if targetPt == "" {
		return
	}

	if targetPt == ipAddress {
		for _, dt := range dtp.Tags {
			if dt.TagName != string(TO_DEVICE) && validdateDuplication(dt.TagName, pt) {
				label := prompb.Label{Name: dt.TagName, Value: dt.TagValue}
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
