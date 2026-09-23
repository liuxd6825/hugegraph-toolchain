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

func newAllShortestPathsFunc(t api.Transport) AllShortestPaths {
	return func(o ...func(*AllShortestPathsRequest)) (*AllShortestPathsResponse, error) {
		var r = AllShortestPathsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type AllShortestPaths func(o ...func(*AllShortestPathsRequest)) (*AllShortestPathsResponse, error)

type AllShortestPathsRequest struct {
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

type AllShortestPathsResponse struct {
	StatusCode int                          `json:"-"`
	Header     http.Header                  `json:"-"`
	Body       io.ReadCloser                `json:"-"`
	Data       AllShortestPathsResponseData `json:"-"`
}

type AllShortestPathsResponseData struct {
	Paths []AllShortestPathsPath `json:"paths"`
}

type AllShortestPathsPath struct {
	Objects []interface{} `json:"objects"`
}

func (r AllShortestPathsRequest) Do(ctx context.Context, transport api.Transport) (*AllShortestPathsResponse, error) {
	if len(r.source) == 0 {
		return nil, errors.New("all_shortest_paths: source is required")
	}
	if len(r.target) == 0 {
		return nil, errors.New("all_shortest_paths: target is required")
	}
	if r.maxDepth <= 0 {
		return nil, errors.New("all_shortest_paths: max_depth must be > 0")
	}

	params := &url.Values{}
	params.Add("source", quoteVertexID(r.source))
	params.Add("target", quoteVertexID(r.target))
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

	url := buildTraverserURL(transport, "allshortestpaths")
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

	data := AllShortestPathsResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &AllShortestPathsResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (a AllShortestPaths) WithContext(ctx context.Context) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.ctx = ctx }
}
func (a AllShortestPaths) WithSource(v string) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.source = v }
}
func (a AllShortestPaths) WithTarget(v string) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.target = v }
}
func (a AllShortestPaths) WithDirection(v string) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.direction = v }
}
func (a AllShortestPaths) WithMaxDepth(v int) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.maxDepth = v }
}
func (a AllShortestPaths) WithLabel(v string) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.label = v }
}
func (a AllShortestPaths) WithMaxDegree(v int64) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.maxDegree = v }
}
func (a AllShortestPaths) WithSkipDegree(v int64) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.skipDegree = v }
}
func (a AllShortestPaths) WithCapacity(v int64) func(*AllShortestPathsRequest) {
	return func(r *AllShortestPathsRequest) { r.capacity = v }
}
