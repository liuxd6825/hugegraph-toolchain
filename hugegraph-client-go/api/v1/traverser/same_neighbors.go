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

func newSameNeighborsFunc(t api.Transport) SameNeighbors {
	return func(o ...func(*SameNeighborsRequest)) (*SameNeighborsResponse, error) {
		var r = SameNeighborsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type SameNeighbors func(o ...func(*SameNeighborsRequest)) (*SameNeighborsResponse, error)

type SameNeighborsRequest struct {
	ctx       context.Context
	vertex    string
	other     string
	direction string
	label     string
	maxDegree int64
	limit     int64
}

type SameNeighborsResponse struct {
	StatusCode int                       `json:"-"`
	Header     http.Header               `json:"-"`
	Body       io.ReadCloser             `json:"-"`
	Data       SameNeighborsResponseData `json:"-"`
}

type SameNeighborsResponseData struct {
	SameNeighbors []interface{} `json:"same_neighbors"`
}

func (r SameNeighborsRequest) Do(ctx context.Context, transport api.Transport) (*SameNeighborsResponse, error) {
	if len(r.vertex) == 0 {
		return nil, errors.New("same_neighbors: vertex is required")
	}
	if len(r.other) == 0 {
		return nil, errors.New("same_neighbors: other is required")
	}

	params := &url.Values{}
	params.Add("vertex", quoteVertexID(r.vertex))
	params.Add("other", quoteVertexID(r.other))
	if r.direction != "" {
		params.Add("direction", r.direction)
	}
	if r.label != "" {
		params.Add("label", r.label)
	}
	if r.maxDegree > 0 {
		params.Add("max_degree", int64ToString(r.maxDegree))
	}
	if r.limit > 0 {
		params.Add("limit", int64ToString(r.limit))
	}

	url := buildTraverserURL(transport, "sameneighbors")
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

	data := SameNeighborsResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &SameNeighborsResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (s SameNeighbors) WithContext(ctx context.Context) func(*SameNeighborsRequest) {
	return func(r *SameNeighborsRequest) { r.ctx = ctx }
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
func (s SameNeighbors) WithMaxDegree(v int64) func(*SameNeighborsRequest) {
	return func(r *SameNeighborsRequest) { r.maxDegree = v }
}
func (s SameNeighbors) WithLimit(v int64) func(*SameNeighborsRequest) {
	return func(r *SameNeighborsRequest) { r.limit = v }
}
