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

// WeightedShortestPath is the fluent method binding for the
// "GET /traversers/weightedshortestpath" endpoint.
//
// 3.2.10 Weighted Shortest Path
type WeightedShortestPath func(o ...func(*WeightedShortestPathRequest)) (*WeightedShortestPathResponse, error)

func newWeightedShortestPathFunc(t api.Transport) WeightedShortestPath {
	return func(o ...func(*WeightedShortestPathRequest)) (*WeightedShortestPathResponse, error) {
		var r = WeightedShortestPathRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type WeightedShortestPathRequest struct {
	ctx        context.Context
	source     string
	target     string
	direction  string
	label      string
	weight     string
	maxDegree  int
	skipDegree int
	capacity   int
	withVertex bool
}

type WeightedShortestPathResponse struct {
	StatusCode int                              `json:"-"`
	Header     http.Header                      `json:"-"`
	Body       io.ReadCloser                    `json:"-"`
	Data       WeightedShortestPathResponseData `json:"-"`
}

// WeightedShortestPathResponseData mirrors the actual response body shape
// documented at 3.2.10:
//
//	{ "path": { "weight": 2.0, "vertices": [...] }, "vertices": [...] }
type WeightedShortestPathResponseData struct {
	Path     WeightedPathInline       `json:"path"`
	Vertices []map[string]interface{} `json:"vertices,omitempty"`
}

// WeightedPathInline is the inline representation of a single weighted path.
type WeightedPathInline struct {
	Weight   float64  `json:"weight"`
	Vertices []string `json:"vertices,omitempty"`
}

func (s WeightedShortestPath) WithSource(v string) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.source = v }
}
func (s WeightedShortestPath) WithTarget(v string) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.target = v }
}
func (s WeightedShortestPath) WithDirection(v string) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.direction = v }
}
func (s WeightedShortestPath) WithLabel(v string) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.label = v }
}
func (s WeightedShortestPath) WithWeight(v string) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.weight = v }
}
func (s WeightedShortestPath) WithDegree(v int) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.maxDegree = v }
}
func (s WeightedShortestPath) WithSkipDegree(v int) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.skipDegree = v }
}
func (s WeightedShortestPath) WithCapacity(v int) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.capacity = v }
}
func (s WeightedShortestPath) WithVertex(v bool) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.withVertex = v }
}
func (s WeightedShortestPath) WithContext(ctx context.Context) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.ctx = ctx }
}

func (r WeightedShortestPathRequest) Do(ctx context.Context, transport api.Transport) (*WeightedShortestPathResponse, error) {
	if r.source == "" {
		return nil, errors.New("source is required")
	}
	if r.target == "" {
		return nil, errors.New("target is required")
	}
	if r.weight == "" {
		return nil, errors.New("weight is required (numeric property name)")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	params.Set("source", graph.FormatVertexID(r.source))
	params.Set("target", graph.FormatVertexID(r.target))
	params.Set("weight", r.weight)
	if r.direction != "" {
		params.Set("direction", r.direction)
	}
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
	if r.withVertex {
		params.Set("with_vertex", "true")
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/weightedshortestpath", params, nil)
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

	data := WeightedShortestPathResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &WeightedShortestPathResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
