package labeler

import "encoding/json"

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
