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

// Rings is the fluent method binding for the
// "GET /traversers/rings" endpoint.
//
// 3.2.19 Rings
type Rings func(o ...func(*RingsRequest)) (*RingsResponse, error)

func newRingsFunc(t api.Transport) Rings {
	return func(o ...func(*RingsRequest)) (*RingsResponse, error) {
		var r = RingsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type RingsRequest struct {
	ctx          context.Context
	source       string
	direction    string
	label        string
	maxDepth     int
	sourceInRing bool
	maxDegree    int
	capacity     int
	limit        int
}

type RingsResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       RingsResponseData
}

type RingsResponseData struct {
	Rings []PathsGetPathEntry `json:"rings"`
}

func (r Rings) WithSource(v string) func(*RingsRequest) {
	return func(req *RingsRequest) { req.source = v }
}
func (r Rings) WithDirection(v string) func(*RingsRequest) {
	return func(req *RingsRequest) { req.direction = v }
}
func (r Rings) WithLabel(v string) func(*RingsRequest) {
	return func(req *RingsRequest) { req.label = v }
}
func (r Rings) WithMaxDepth(v int) func(*RingsRequest) {
	return func(req *RingsRequest) { req.maxDepth = v }
}
func (r Rings) WithSourceInRing(v bool) func(*RingsRequest) {
	return func(req *RingsRequest) { req.sourceInRing = v }
}
func (r Rings) WithDegree(v int) func(*RingsRequest) {
	return func(req *RingsRequest) { req.maxDegree = v }
}
func (r Rings) WithCapacity(v int) func(*RingsRequest) {
	return func(req *RingsRequest) { req.capacity = v }
}
func (r Rings) WithLimit(v int) func(*RingsRequest) {
	return func(req *RingsRequest) { req.limit = v }
}
func (r Rings) WithContext(ctx context.Context) func(*RingsRequest) {
	return func(req *RingsRequest) { req.ctx = ctx }
}

func (req RingsRequest) Do(ctx context.Context, transport api.Transport) (*RingsResponse, error) {
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
	if !req.sourceInRing {
		params.Set("source_in_ring", "false")
	}
	if req.maxDegree > 0 {
		params.Set("max_degree", strconv.Itoa(req.maxDegree))
	}
	if req.capacity > 0 {
		params.Set("capacity", strconv.Itoa(req.capacity))
	}
	if req.limit > 0 {
		params.Set("limit", strconv.Itoa(req.limit))
	}

	r, err := api.NewRequest("GET", basePath(cfg)+"/traversers/rings", params, nil)
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

	data := RingsResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &RingsResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
