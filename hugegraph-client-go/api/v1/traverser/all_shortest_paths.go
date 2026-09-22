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

// AllShortestPaths is the fluent method binding for the
// "GET /traversers/allshortestpaths" endpoint.
//
// 3.2.9 All Shortest Paths
type AllShortestPaths func(o ...func(*AllShortestPathsRequest)) (*AllShortestPathsResponse, error)

func newAllShortestPathsFunc(t api.Transport) AllShortestPaths {
	return func(o ...func(*AllShortestPathsRequest)) (*AllShortestPathsResponse, error) {
		var r = AllShortestPathsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type AllShortestPathsRequest struct {
	ctx        context.Context
	source     string
	target     string
	direction  string
	maxDepth   int
	label      string
	maxDegree  int
	skipDegree int
	capacity   int
}

type AllShortestPathsResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	// The API returns a JSON object {"paths":[{objects:...}]}.
	Data AllShortestPathsResponseData `json:"-"`
}

type AllShortestPathsResponseData struct {
	Paths []AllShortestPath `json:"paths"`
}

// AllShortestPath wraps a sequence of vertex IDs.
type AllShortestPath struct {
	Objects []interface{} `json:"objects"`
}

func (s AllShortestPaths) WithSource(v string) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.source = v }
}
func (s AllShortestPaths) WithTarget(v string) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.target = v }
}
func (s AllShortestPaths) WithDirection(v string) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.direction = v }
}
func (s AllShortestPaths) WithMaxDepth(v int) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.maxDepth = v }
}
func (s AllShortestPaths) WithLabel(v string) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.label = v }
}
func (s AllShortestPaths) WithDegree(v int) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.maxDegree = v }
}
func (s AllShortestPaths) WithSkipDegree(v int) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.skipDegree = v }
}
func (s AllShortestPaths) WithCapacity(v int) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.capacity = v }
}
func (s AllShortestPaths) WithContext(ctx context.Context) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.ctx = ctx }
}

func (r AllShortestPathsRequest) Do(ctx context.Context, transport api.Transport) (*AllShortestPathsResponse, error) {
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
	params.Set("max_depth", strconv.Itoa(r.maxDepth))
	if r.label != "" {
		params.Set("label", r.label)
	}
	if r.maxDegree > 0 {
		params.Set("max_degree", strconv.Itoa(r.maxDegree))
	}
	if r.skipDegree > 0 {
		params.Set("skip_degree", strconv.Itoa(r.skipDegree))
	}
	if r.capacity > 0 {
		params.Set("capacity", strconv.Itoa(r.capacity))
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/allshortestpaths", params, nil)
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

	data := AllShortestPathsResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &AllShortestPathsResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
