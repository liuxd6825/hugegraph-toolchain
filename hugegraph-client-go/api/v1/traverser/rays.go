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

// Rays is the fluent method binding for the
// "GET /traversers/rays" endpoint.
//
// 3.2.20 Rays
type Rays func(o ...func(*RaysRequest)) (*RaysResponse, error)

func newRaysFunc(t api.Transport) Rays {
	return func(o ...func(*RaysRequest)) (*RaysResponse, error) {
		var r = RaysRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type RaysRequest struct {
	ctx       context.Context
	source    string
	direction string
	label     string
	maxDepth  int
	maxDegree int
	capacity  int
	limit     int
}

type RaysResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       RaysResponseData
}

type RaysResponseData struct {
	Rays []PathsGetPathEntry `json:"rays"`
}

func (r Rays) WithSource(v string) func(*RaysRequest) {
	return func(req *RaysRequest) { req.source = v }
}
func (r Rays) WithDirection(v string) func(*RaysRequest) {
	return func(req *RaysRequest) { req.direction = v }
}
func (r Rays) WithLabel(v string) func(*RaysRequest) {
	return func(req *RaysRequest) { req.label = v }
}
func (r Rays) WithMaxDepth(v int) func(*RaysRequest) {
	return func(req *RaysRequest) { req.maxDepth = v }
}
func (r Rays) WithDegree(v int) func(*RaysRequest) {
	return func(req *RaysRequest) { req.maxDegree = v }
}
func (r Rays) WithCapacity(v int) func(*RaysRequest) {
	return func(req *RaysRequest) { req.capacity = v }
}
func (r Rays) WithLimit(v int) func(*RaysRequest) {
	return func(req *RaysRequest) { req.limit = v }
}
func (r Rays) WithContext(ctx context.Context) func(*RaysRequest) {
	return func(req *RaysRequest) { req.ctx = ctx }
}

func (req RaysRequest) Do(ctx context.Context, transport api.Transport) (*RaysResponse, error) {
	if req.source == "" {
		return nil, errors.New("source is required")
	}
	if req.maxDepth <= 0 {
		return nil, errors.New("max_depth must be > 0")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	params.Set("source", graph.FormatVertexID(req.source))
	if req.direction != "" {
		params.Set("direction", req.direction)
	}
	if req.label != "" {
		params.Set("label", req.label)
	}
	params.Set("max_depth", strconv.Itoa(req.maxDepth))
	if req.maxDegree > 0 {
		params.Set("max_degree", strconv.Itoa(req.maxDegree))
	}
	if req.capacity > 0 {
		params.Set("capacity", strconv.Itoa(req.capacity))
	}
	if req.limit > 0 {
		params.Set("limit", strconv.Itoa(req.limit))
	}

	r, err := api.NewRequest("GET", basePath(cfg)+"/traversers/rays", params, nil)
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		r = r.WithContext(ctx)
	}

	res, err := transport.Perform(r)
	if err != nil {
		return nil, err
	}

	data := RaysResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &RaysResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
