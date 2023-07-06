package labeler

import (
	"github.com/prometheus/prometheus/prompb"
	"strings"
)

type PeerRelationLabelBuilder struct{}

func (pe PeerRelationLabelBuilder) MatchDeviceTagPair(ipAddress string) (matched bool, dtp DeviceTagPair) {
	for _, v := range REDIS_TAGS {
		if ipAddress == v.IP {
			matched = true
			dtp = v
			return
		}
	}
	return
}

func (pe PeerRelationLabelBuilder) Build(ipAddress string, pt *prompb.TimeSeries) {
	matched, dtp := pe.MatchDeviceTagPair(ipAddress)
	if !matched {
		return
	}
	ifDescrPt := ""
	for _, v := range pt.Labels {
		if v.Name == string(IF_DESCR) {
			ifDescrPt = v.Value
			break
		}
	}

	if ifDescrPt == "" {
		return
	}
	for _, dt := range dtp.Tags {
		tagSplitArr := strings.Split(dt.TagName, "->")
		if dt.Dimension != string(TO_DEVICE) {
			label := prompb.Label{Name: dt.TagLabel, Value: dt.TagName}
			pt.Labels = append(pt.Labels, &label)
		} else if strings.Contains(dt.TagName, "->") && tagSplitArr[0] == ifDescrPt {
			peerPortLabel := prompb.Label{Name: string(PEER_PORT), Value: tagSplitArr[1]}
			pt.Labels = append(pt.Labels, &peerPortLabel)
			peerDeviceLabel := prompb.Label{Name: string(PEER_DEVICE), Value: tagSplitArr[2]}
			pt.Labels = append(pt.Labels, &peerDeviceLabel)
		}
	}
}
