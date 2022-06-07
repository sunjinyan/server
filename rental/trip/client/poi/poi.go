package poi

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"github.com/golang/protobuf/proto"
	"hash/fnv"
)

var poi = []string{
	"中关村",
	"天安门",
	"陆家嘴",
	"迪士尼",
	"广州塔",
	"天河体育中心",
}

type Manager struct {
	
}

func (m *Manager)Resolve(ctx context.Context,lc *rentalpb.Location)(string,error) {
	b,err := proto.Marshal(lc)
	if err != nil {
		return "", err
	}

	h := fnv.New32()
	_, err = h.Write(b)
	if err != nil {

	}
	//return poi[rand.Intn(len(poi))], nil
	return poi[int(h.Sum32()) % len(poi)], nil
}