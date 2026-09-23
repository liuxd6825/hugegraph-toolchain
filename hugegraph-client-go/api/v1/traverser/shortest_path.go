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

func newShortestPathFunc(t api.Transport) ShortestPath {
	return func(o ...func(*ShortestPathRequest)) (*ShortestPathResponse, error) {
		var r = ShortestPathRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type ShortestPath func(o ...func(*ShortestPathRequest)) (*ShortestPathResponse, error)

type ShortestPathRequest struct {
	ctx        context.Context
	source     string
	target     string
	direction  string
	maxDepth   int
	label      string
	maxDegree  int64
	skipDegree int64
	capacity   int64
}

type ShortestPathResponse struct {
	StatusCode int                      `json:"-"`
	Header     http.Header              `json:"-"`
	Body       io.ReadCloser            `json:"-"`
	Data       ShortestPathResponseData `json:"-"`
}

type ShortestPathResponseData struct {
	Path     []interface{} `json:"path"`
	Vertices []string      `json:"vertices,omitempty"`
	Edges    []string      `json:"edges,omitempty"`
}

func (r ShortestPathRequest) Do(ctx context.Context, transport api.Transport) (*ShortestPathResponse, error) {
	if len(r.source) == 0 {
		return nil, errors.New("shortest_path: source is required")
	}
	if len(r.target) == 0 {
		return nil, errors.New("shortest_path: target is required")
	}
	if r.maxDepth <= 0 {
		return nil, errors.New("shortest_path: max_depth must be > 0")
	}

	params := &url.Values{}
	params.Add("source", "\""+r.source+"\"")
	params.Add("target", "\""+r.target+"\"")
	params.Add("max_depth", intToString(r.maxDepth))
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

	url := buildTraverserURL(transport, "shortestpath")
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

	data := ShortestPathResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &ShortestPathResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (s ShortestPath) WithContext(ctx context.Context) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.ctx = ctx }
}
func (s ShortestPath) WithSource(v string) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.source = v }
}
func (s ShortestPath) WithTarget(v string) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.target = v }
}
func (s ShortestPath) WithDirection(v string) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.direction = v }
}
func (s ShortestPath) WithMaxDepth(v int) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.maxDepth = v }
}
func (s ShortestPath) WithLabel(v string) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.label = v }
}
func (s ShortestPath) WithMaxDegree(v int64) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.maxDegree = v }
}
func (s ShortestPath) WithSkipDegree(v int64) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.skipDegree = v }
}
func (s ShortestPath) WithCapacity(v int64) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.capacity = v }
}
