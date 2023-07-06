package labeler

import (
	"github.com/ccfos/nightingale/v6/pushgw/router"
	"github.com/prometheus/prometheus/prompb"
)

type LabelBuilder interface {
	MatchDeviceTagPair(keyValue string) (matched bool, dtp router.DeviceTagPair)

	Build(tagValue string, pt *prompb.TimeSeries)
}
