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

func newWeightedShortestPathFunc(t api.Transport) WeightedShortestPath {
	return func(o ...func(*WeightedShortestPathRequest)) (*WeightedShortestPathResponse, error) {
		var r = WeightedShortestPathRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type WeightedShortestPath func(o ...func(*WeightedShortestPathRequest)) (*WeightedShortestPathResponse, error)

type WeightedShortestPathRequest struct {
	ctx        context.Context
	source     string
	target     string
	weight     string
	direction  string
	label      string
	maxDegree  int64
	skipDegree int64
	capacity   int64
	withVertex bool
}

type WeightedShortestPathResponse struct {
	StatusCode int                              `json:"-"`
	Header     http.Header                      `json:"-"`
	Body       io.ReadCloser                    `json:"-"`
	Data       WeightedShortestPathResponseData `json:"-"`
}

type WeightedShortestPathResponseData struct {
	Path     WeightedShortestPathPath     `json:"path"`
	Vertices []WeightedShortestPathVertex `json:"vertices,omitempty"`
}

type WeightedShortestPathPath struct {
	Weight   float64  `json:"weight"`
	Vertices []string `json:"vertices"`
}

type WeightedShortestPathVertex struct {
	ID         string                 `json:"id"`
	Label      string                 `json:"label"`
	Typ        string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

func (r WeightedShortestPathRequest) Do(ctx context.Context, transport api.Transport) (*WeightedShortestPathResponse, error) {
	if len(r.source) == 0 {
		return nil, errors.New("weighted_shortest_path: source is required")
	}
	if len(r.target) == 0 {
		return nil, errors.New("weighted_shortest_path: target is required")
	}
	if len(r.weight) == 0 {
		return nil, errors.New("weighted_shortest_path: weight is required")
	}

	params := &url.Values{}
	params.Add("source", quoteVertexID(r.source))
	params.Add("target", quoteVertexID(r.target))
	params.Add("weight", r.weight)
	if r.direction != "" {
		params.Add("direction", r.direction)
	}
	if r.label != "" {
		params.Add("label", r.label)
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
	if r.withVertex {
		params.Add("with_vertex", boolToString(r.withVertex))
	}

	url := getURL(transport, "weightedshortestpath")
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

	data := WeightedShortestPathResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &WeightedShortestPathResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (w WeightedShortestPath) WithContext(ctx context.Context) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.ctx = ctx }
}
func (w WeightedShortestPath) WithSource(v string) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.source = v }
}
func (w WeightedShortestPath) WithTarget(v string) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.target = v }
}
func (w WeightedShortestPath) WithWeight(v string) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.weight = v }
}
func (w WeightedShortestPath) WithDirection(v string) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.direction = v }
}
func (w WeightedShortestPath) WithLabel(v string) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.label = v }
}
func (w WeightedShortestPath) WithMaxDegree(v int64) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.maxDegree = v }
}
func (w WeightedShortestPath) WithSkipDegree(v int64) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.skipDegree = v }
}
func (w WeightedShortestPath) WithCapacity(v int64) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.capacity = v }
}
func (w WeightedShortestPath) WithVertex(v bool) func(*WeightedShortestPathRequest) {
	return func(r *WeightedShortestPathRequest) { r.withVertex = v }
}
