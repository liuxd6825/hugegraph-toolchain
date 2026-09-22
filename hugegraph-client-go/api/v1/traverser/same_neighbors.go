/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements. See the NOTICE file distributed with this
 * work for additional information regarding copyright ownership. The ASF
 * licenses this file to You under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 */

package traverser

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/api"
	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/graph"
	structtraverser "github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/traverser"
)

// SameNeighbors is the fluent method binding for the
// "GET /traversers/sameneighbors" endpoint.
//
// 3.2.5 Same Neighbors
type SameNeighbors func(o ...func(*SameNeighborsRequest)) (*SameNeighborsResponse, error)

func newSameNeighborsFunc(t api.Transport) SameNeighbors {
	return func(o ...func(*SameNeighborsRequest)) (*SameNeighborsResponse, error) {
		var r = SameNeighborsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type SameNeighborsRequest struct {
	ctx       context.Context
	vertex    string
	other     string
	direction string
	label     string
	maxDegree int
	limit     int
}

type SameNeighborsResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       structtraverser.SameNeighbors
}

func (s SameNeighbors) WithVertex(v string) func(*SameNeighborsRequest) {
	return func(r *SameNeighborsRequest) { r.vertex = v }
}
func (s SameNeighbors) WithOther(v string) func(*SameNeighborsRequest) {
	return func(r *SameNeighborsRequest) { r.other = v }
}
func (s SameNeighbors) WithDirection(v string) func(*SameNeighborsRequest) {
	return func(r *SameNeighborsRequest) { r.direction = v }
}
func (s SameNeighbors) WithLabel(v string) func(*SameNeighborsRequest) {
	return func(r *SameNeighborsRequest) { r.label = v }
}
func (s SameNeighbors) WithDegree(v int) func(*SameNeighborsRequest) {
	return func(r *SameNeighborsRequest) { r.maxDegree = v }
}
func (s SameNeighbors) WithLimit(v int) func(*SameNeighborsRequest) {
	return func(r *SameNeighborsRequest) { r.limit = v }
}
func (s SameNeighbors) WithContext(ctx context.Context) func(*SameNeighborsRequest) {
	return func(r *SameNeighborsRequest) { r.ctx = ctx }
}

func (r SameNeighborsRequest) Do(ctx context.Context, transport api.Transport) (*SameNeighborsResponse, error) {
	if r.vertex == "" {
		return nil, errors.New("vertex is required")
	}
	if r.other == "" {
		return nil, errors.New("other is required")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	params.Set("vertex", graph.FormatVertexID(r.vertex))
	params.Set("other", graph.FormatVertexID(r.other))
	if r.direction != "" {
		params.Set("direction", r.direction)
	}
	if r.label != "" {
		params.Set("label", r.label)
	}
	if r.maxDegree > 0 {
		params.Set("max_degree", strconv.Itoa(r.maxDegree))
	}
	if r.limit > 0 {
		params.Set("limit", strconv.Itoa(r.limit))
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/sameneighbors", params, nil)
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	res, err := transport.Perform(req)
	if err != nil {
		return nil, err
	}

	data := structtraverser.SameNeighbors{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &SameNeighborsResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
