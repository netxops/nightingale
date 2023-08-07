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
		tagSplitArr := strings.Split(dt.TagValue, "->")
		switch dt.TagName {
		case string(SPECIAL_NETWORK_LINE):
			if strings.Contains(dt.TagValue, "->") && tagSplitArr[1] == ifDescrPt {
				specialLineLabel := prompb.Label{Name: string(SPECIAL_NETWORK_LINE), Value: tagSplitArr[2]}
				pt.Labels = append(pt.Labels, &specialLineLabel)
			}
			break
		case string(TO_DEVICE):
			if strings.Contains(dt.TagValue, "->") && tagSplitArr[0] == ifDescrPt {
				peerPortLabel := prompb.Label{Name: string(PEER_PORT), Value: tagSplitArr[1]}
				pt.Labels = append(pt.Labels, &peerPortLabel)
				peerDeviceLabel := prompb.Label{Name: string(PEER_DEVICE), Value: tagSplitArr[2]}
				pt.Labels = append(pt.Labels, &peerDeviceLabel)
				peerDeviceCatalogLabel := prompb.Label{Name: string(PEER_DEVICE_CATALOG), Value: tagSplitArr[3]}
				pt.Labels = append(pt.Labels, &peerDeviceCatalogLabel)
			}
			break
		default:
			label := prompb.Label{Name: dt.TagName, Value: dt.TagValue}
			pt.Labels = append(pt.Labels, &label)
			break
		}
	}
}
