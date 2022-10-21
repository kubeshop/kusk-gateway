// MIT License
//
// Copyright (c) 2022 Kubeshop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
