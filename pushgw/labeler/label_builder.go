package labeler

import (
	"github.com/prometheus/prometheus/prompb"
	"strings"
)

type LabelBuilder interface {
	MatchDeviceTagPair(keyValue string) (matched bool, dtp DeviceTagPair)

	Build(tagValue string, pt *prompb.TimeSeries)
}

func has(key string, pt *prompb.TimeSeries) (keyValue string, matched bool) {
	for _, v := range pt.Labels {
		if strings.Contains(v.Name, key) {
			return v.Value, true
		}
	}
	return
}

func RemakeWriteRemoteEnrichLabels(pt *prompb.TimeSeries) {
	if identVal, identExist := has(string(IDENT), pt); identExist {
		peerRelationBuilder := PeerRelationLabelBuilder{}
		peerRelationBuilder.Build(identVal+"/24", pt)
	}

	if targetVal, targetExist := has(string(TARGET), pt); targetExist {
		targetBuilder := TargetLabelBuilder{}
		targetBuilder.Build(targetVal+"/24", pt)
	}
}
