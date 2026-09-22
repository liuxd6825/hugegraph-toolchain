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
)

// Crosspoints is the fluent method binding for the
// "GET /traversers/crosspoints" endpoint.
//
// 3.2.17 Crosspoints
type Crosspoints func(o ...func(*CrosspointsRequest)) (*CrosspointsResponse, error)

func newCrosspointsFunc(t api.Transport) Crosspoints {
	return func(o ...func(*CrosspointsRequest)) (*CrosspointsResponse, error) {
		var r = CrosspointsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type CrosspointsRequest struct {
	ctx       context.Context
	source    string
	target    string
	direction string
	label     string
	maxDepth  int
	maxDegree int
	capacity  int
	limit     int
}

type CrosspointsResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       CrosspointsResponseData
}

// CrosspointsResponseData mirrors the documented response shape:
//
//	{
//	  "crosspoints": [
//	    { "crosspoint": "1:josh", "objects": ["2:lop","1:josh","2:ripple"] }
//	  ]
//	}
type CrosspointsResponseData struct {
	Crosspoints []CrosspointEntry `json:"crosspoints"`
}

// CrosspointEntry is a single crosspoint match.
type CrosspointEntry struct {
	Crosspoint string        `json:"crosspoint"`
	Objects    []interface{} `json:"objects"`
}

func (c Crosspoints) WithSource(v string) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.source = v }
}
func (c Crosspoints) WithTarget(v string) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.target = v }
}
func (c Crosspoints) WithDirection(v string) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.direction = v }
}
func (c Crosspoints) WithLabel(v string) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.label = v }
}
func (c Crosspoints) WithMaxDepth(v int) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.maxDepth = v }
}
func (c Crosspoints) WithDegree(v int) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.maxDegree = v }
}
func (c Crosspoints) WithCapacity(v int) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.capacity = v }
}
func (c Crosspoints) WithLimit(v int) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.limit = v }
}
func (c Crosspoints) WithContext(ctx context.Context) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.ctx = ctx }
}

func (r CrosspointsRequest) Do(ctx context.Context, transport api.Transport) (*CrosspointsResponse, error) {
	if r.source == "" {
		return nil, errors.New("source is required")
	}
	if r.target == "" {
		return nil, errors.New("target is required")
	}
	if r.maxDepth <= 0 {
		return nil, errors.New("max_depth must be > 0")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	params.Set("source", graph.FormatVertexID(r.source))
	params.Set("target", graph.FormatVertexID(r.target))
	if r.direction != "" {
		params.Set("direction", r.direction)
	}
	if r.label != "" {
		params.Set("label", r.label)
	}
	params.Set("max_depth", strconv.Itoa(r.maxDepth))
	if r.maxDegree > 0 {
		params.Set("max_degree", strconv.Itoa(r.maxDegree))
	}
	if r.capacity > 0 {
		params.Set("capacity", strconv.Itoa(r.capacity))
	}
	if r.limit > 0 {
		params.Set("limit", strconv.Itoa(r.limit))
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/crosspoints", params, nil)
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

	data := CrosspointsResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &CrosspointsResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
