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

	"github.com/apache/hugegraph-toolchain/hugegraph-client-go/api"
)

func newSingleSourceShortestPathFunc(t api.Transport) SingleSourceShortestPath {
	return func(o ...func(*SingleSourceShortestPathRequest)) (*SingleSourceShortestPathResponse, error) {
		var r = SingleSourceShortestPathRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type SingleSourceShortestPath func(o ...func(*SingleSourceShortestPathRequest)) (*SingleSourceShortestPathResponse, error)

type SingleSourceShortestPathRequest struct {
	ctx        context.Context
	source     string
	direction  string
	label      string
	weight     string
	maxDegree  int64
	skipDegree int64
	capacity   int64
	limit      int
	withVertex bool
}

type SingleSourceShortestPathResponse struct {
	StatusCode int                                  `json:"-"`
	Header     http.Header                          `json:"-"`
	Body       io.ReadCloser                        `json:"-"`
	Data       SingleSourceShortestPathResponseData `json:"-"`
}

type SingleSourceShortestPathResponseData struct {
	Paths    map[string]SingleSourceTargetPath `json:"paths"`
	Vertices []WeightedShortestPathVertex      `json:"vertices,omitempty"`
}

type SingleSourceTargetPath struct {
	Weight   float64  `json:"weight"`
	Vertices []string `json:"vertices"`
}

func (r SingleSourceShortestPathRequest) Do(ctx context.Context, transport api.Transport) (*SingleSourceShortestPathResponse, error) {
	if len(r.source) == 0 {
		return nil, errors.New("single_source_shortest_path: source is required")
	}

	params := &url.Values{}
	params.Add("source", quoteVertexID(r.source))
	if r.direction != "" {
		params.Add("direction", r.direction)
	}
	if r.label != "" {
		params.Add("label", r.label)
	}
	if r.weight != "" {
		params.Add("weight", r.weight)
	}
	if r.maxDegree > 0 {
		params.Add("max_degree", int64ToString(r.maxDegree))
	}
	if r.skipDegree > 0 {
		params.Add("skip_degree", int64ToString(r.skipDegree))
	}
	if r.capacity > 0 {
		params.Add("capacity", int64ToString(r.capacity))
	}
	if r.limit > 0 {
		params.Add("limit", intToString(r.limit))
	}
	if r.withVertex {
		params.Add("with_vertex", boolToString(r.withVertex))
	}

	url := buildTraverserURL(transport, "singlesourceshortestpath")
	req, err := api.NewRequest("GET", url, params, nil)
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

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := checkHTTPStatus(res, bytes); err != nil {
		return nil, err
	}

	data := SingleSourceShortestPathResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &SingleSourceShortestPathResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (s SingleSourceShortestPath) WithContext(ctx context.Context) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.ctx = ctx }
}
func (s SingleSourceShortestPath) WithSource(v string) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.source = v }
}
func (s SingleSourceShortestPath) WithDirection(v string) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.direction = v }
}
func (s SingleSourceShortestPath) WithLabel(v string) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.label = v }
}
func (s SingleSourceShortestPath) WithWeight(v string) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.weight = v }
}
func (s SingleSourceShortestPath) WithMaxDegree(v int64) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.maxDegree = v }
}
func (s SingleSourceShortestPath) WithSkipDegree(v int64) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.skipDegree = v }
}
func (s SingleSourceShortestPath) WithCapacity(v int64) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.capacity = v }
}
func (s SingleSourceShortestPath) WithLimit(v int) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.limit = v }
}
func (s SingleSourceShortestPath) WithVertex(v bool) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.withVertex = v }
}
