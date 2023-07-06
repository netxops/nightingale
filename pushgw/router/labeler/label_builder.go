package router

import (
	rt "github.com/ccfos/nightingale/v6/pushgw/router"
	"github.com/prometheus/prometheus/prompb"
)

type LabelBuilder interface {
	MatchDeviceTagPair(keyValue string) (matched bool, dtp rt.DeviceTagPair)

	Build(tagValue string, pt *prompb.TimeSeries)
}
