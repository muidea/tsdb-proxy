package snapshot

import (
	"sync"

	"supos.ai/data-lake/external/tsdb-proxy/common/model"
)

// Snapshot snapshot
type Snapshot interface {
	Name() string
	UpdateTagValue(val *model.NamedValue) bool
	UpdateTagValues(values *model.ValueSequnce) *model.ValueSequnce

	GetAllCurrentTagValue() *model.ValueSequnce
}

// NodeValueInfo nodeName -> nodeValue
type NodeValueInfo map[string]*model.Value

type nodeRegistry struct {
	name       string
	nodeValue  NodeValueInfo
	routesLock sync.RWMutex
}

// New create snapshot
func New(name string) Snapshot {
	return &nodeRegistry{name: name, nodeValue: NodeValueInfo{}}
}

func (s *nodeRegistry) Name() string {
	return s.name
}

func (s *nodeRegistry) UpdateTagValue(val *model.NamedValue) bool {
	s.routesLock.Lock()
	defer s.routesLock.Unlock()
	name := val.GetName()

	_, ok := s.nodeValue[name]
	s.nodeValue[name] = val.GetValue()

	//log.Printf("current name:%s, value:%s", name, value.String())
	return !ok
}

func (s *nodeRegistry) UpdateTagValues(values *model.ValueSequnce) *model.ValueSequnce {
	s.routesLock.Lock()
	defer s.routesLock.Unlock()

	newValues := &model.ValueSequnce{Value: []*model.NamedValue{}}
	for _, val := range values.GetValue() {
		name := val.GetName()
		_, ok := s.nodeValue[name]
		if !ok {
			newValues.Value = append(newValues.Value, val)
		}

		s.nodeValue[name] = val.GetValue()
	}

	return newValues
}

func (s *nodeRegistry) GetAllCurrentTagValue() *model.ValueSequnce {
	values := &model.ValueSequnce{Value: []*model.NamedValue{}}
	s.routesLock.RLock()
	defer s.routesLock.RUnlock()
	for key, val := range s.nodeValue {
		values.Value = append(values.Value, &model.NamedValue{Name: key, Value: val})
	}

	return values
}
