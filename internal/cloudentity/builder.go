package cloudentity

import "fmt"

type IdToAPI map[string]*APIGroups

type Builder struct {
	upstreamMap map[string]IdToAPI
}

func NewBuilder() *Builder {
	return &Builder{
		upstreamMap: make(map[string]IdToAPI),
	}
}

func (b *Builder) Len() int {
	return len(b.upstreamMap)
}

func (b *Builder) AddAPI(host string, port uint32, id, name, path, method string) {
	upstream := fmt.Sprintf("%s:%d", host, port)
	idToAPI, ok := b.upstreamMap[upstream]
	if !ok {
		b.upstreamMap[upstream] = make(IdToAPI)
		idToAPI = b.upstreamMap[upstream]
	}
	m, ok := idToAPI[id]
	if !ok {
		idToAPI[id] = &APIGroups{
			Name: name,
			ID:   id,
			Apis: []API{{Path: path, Method: method}},
		}
		return
	}
	m.Apis = append(m.Apis, API{Path: path, Method: method})
}

func (b *Builder) BuildRequest() map[string]*APIGroupsRequest {
	res := make(map[string]*APIGroupsRequest)
	for upstream, idToAPI := range b.upstreamMap {
		req := &APIGroupsRequest{}
		for _, v := range idToAPI {
			req.APIGroups = append(req.APIGroups, *v)
		}
		res[upstream] = req
	}

	return res
}
