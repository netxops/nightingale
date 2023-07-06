package labeler

import (
	"github.com/ccfos/nightingale/v6/pushgw/router"
	"github.com/prometheus/prometheus/prompb"
)

type LabelBuilder interface {
	MatchDeviceTagPair(keyValue string) (matched bool, dtp router.DeviceTagPair)

	Build(tagValue string, pt *prompb.TimeSeries)
}

func RemakeWriteRemoteEnrichLabels(pt *prompb.TimeSeries) {
	if identVal, identExist := router.Has(string(router.IDENT), pt); identExist {
		peerRelationBuilder := PeerRelationLabelBuilder{}
		peerRelationBuilder.Build(identVal+"/24", pt)
	}

	if targetVal, targetExist := router.Has(string(router.TARGET), pt); targetExist {
		targetBuilder := TargetLabelBuilder{}
		targetBuilder.Build(targetVal+"/24", pt)
	}
}
