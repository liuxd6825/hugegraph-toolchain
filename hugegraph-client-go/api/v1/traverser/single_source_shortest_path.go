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

// SingleSourceShortestPath is the fluent method binding for the
// "GET /traversers/singlesourceshortestpath" endpoint.
//
// 3.2.11 Single Source Shortest Path
type SingleSourceShortestPath func(o ...func(*SingleSourceShortestPathRequest)) (*SingleSourceShortestPathResponse, error)

func newSingleSourceShortestPathFunc(t api.Transport) SingleSourceShortestPath {
	return func(o ...func(*SingleSourceShortestPathRequest)) (*SingleSourceShortestPathResponse, error) {
		var r = SingleSourceShortestPathRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type SingleSourceShortestPathRequest struct {
	ctx        context.Context
	source     string
	direction  string
	label      string
	weight     string
	maxDegree  int
	skipDegree int
	capacity   int
	limit      int
	withVertex bool
}

type SingleSourceShortestPathResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	// The API returns {"paths":{vertexId: WeightedPath, ...}, "vertices":[...]}
	Paths    SingleSourceShortestPaths `json:"paths"`
	Vertices []map[string]interface{}  `json:"vertices,omitempty"`
}

// SingleSourceShortestPaths mirrors org.apache.hugegraph.structure.traverser.WeightedPaths.
type SingleSourceShortestPaths struct {
	Source  interface{}                         `json:"source"`
	Weights map[string]SingleSourceWeightedPath `json:"weights"`
}

// SingleSourceWeightedPath mirrors a single WeightedPath response item.
type SingleSourceWeightedPath struct {
	Weight   float64  `json:"weight"`
	Vertices []string `json:"vertices,omitempty"`
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
func (s SingleSourceShortestPath) WithDegree(v int) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.maxDegree = v }
}
func (s SingleSourceShortestPath) WithSkipDegree(v int) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.skipDegree = v }
}
func (s SingleSourceShortestPath) WithCapacity(v int) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.capacity = v }
}
func (s SingleSourceShortestPath) WithLimit(v int) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.limit = v }
}
func (s SingleSourceShortestPath) WithVertex(v bool) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.withVertex = v }
}
func (s SingleSourceShortestPath) WithContext(ctx context.Context) func(*SingleSourceShortestPathRequest) {
	return func(r *SingleSourceShortestPathRequest) { r.ctx = ctx }
}

func (r SingleSourceShortestPathRequest) Do(ctx context.Context, transport api.Transport) (*SingleSourceShortestPathResponse, error) {
	if r.source == "" {
		return nil, errors.New("source is required")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	params.Set("source", graph.FormatVertexID(r.source))
	if r.direction != "" {
		params.Set("direction", r.direction)
	}
	if r.label != "" {
		params.Set("label", r.label)
	}
	if r.weight != "" {
		params.Set("weight", r.weight)
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
	if r.limit > 0 {
		params.Set("limit", strconv.Itoa(r.limit))
	}
	if r.withVertex {
		params.Set("with_vertex", "true")
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/singlesourceshortestpath", params, nil)
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

	raw := struct {
		Paths    SingleSourceShortestPaths `json:"paths"`
		Vertices []map[string]interface{}  `json:"vertices,omitempty"`
	}{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return nil, err
	}

	resp := &SingleSourceShortestPathResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Paths:      raw.Paths,
		Vertices:   raw.Vertices,
	}
	return resp, nil
}
