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

func newRingsFunc(t api.Transport) Rings {
	return func(o ...func(*RingsRequest)) (*RingsResponse, error) {
		var r = RingsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type Rings func(o ...func(*RingsRequest)) (*RingsResponse, error)

type RingsRequest struct {
	ctx          context.Context
	source       string
	maxDepth     int
	direction    string
	label        string
	sourceInRing bool
	maxDegree    int64
	capacity     int64
	limit        int64
}

type RingsResponse struct {
	StatusCode int               `json:"-"`
	Header     http.Header       `json:"-"`
	Body       io.ReadCloser     `json:"-"`
	Data       RingsResponseData `json:"-"`
}

type RingsResponseData struct {
	Rings []RingsPath `json:"rings"`
}

type RingsPath struct {
	Objects []interface{} `json:"objects"`
}

func (r RingsRequest) Do(ctx context.Context, transport api.Transport) (*RingsResponse, error) {
	if len(r.source) == 0 {
		return nil, errors.New("rings: source is required")
	}
	if r.maxDepth <= 0 {
		return nil, errors.New("rings: max_depth must be > 0")
	}

	params := &url.Values{}
	params.Add("source", quoteVertexID(r.source))
	params.Add("max_depth", intToString(r.maxDepth))
	if r.direction != "" {
		params.Add("direction", r.direction)
	}
	if r.label != "" {
		params.Add("label", r.label)
	}
	params.Add("source_in_ring", boolToString(r.sourceInRing))
	if r.maxDegree > 0 {
		params.Add("max_degree", int64ToString(r.maxDegree))
	}
	if r.capacity > 0 {
		params.Add("capacity", int64ToString(r.capacity))
	}
	if r.limit > 0 {
		params.Add("limit", int64ToString(r.limit))
	}

	url := getURL(transport, "rings")
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

	data := RingsResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &RingsResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (r Rings) WithContext(ctx context.Context) func(*RingsRequest) {
	return func(req *RingsRequest) { req.ctx = ctx }
}
func (r Rings) WithSource(v string) func(*RingsRequest) {
	return func(req *RingsRequest) { req.source = v }
}
func (r Rings) WithMaxDepth(v int) func(*RingsRequest) {
	return func(req *RingsRequest) { req.maxDepth = v }
}
func (r Rings) WithDirection(v string) func(*RingsRequest) {
	return func(req *RingsRequest) { req.direction = v }
}
func (r Rings) WithLabel(v string) func(*RingsRequest) {
	return func(req *RingsRequest) { req.label = v }
}
func (r Rings) WithSourceInRing(v bool) func(*RingsRequest) {
	return func(req *RingsRequest) { req.sourceInRing = v }
}
func (r Rings) WithMaxDegree(v int64) func(*RingsRequest) {
	return func(req *RingsRequest) { req.maxDegree = v }
}
func (r Rings) WithCapacity(v int64) func(*RingsRequest) {
	return func(req *RingsRequest) { req.capacity = v }
}
func (r Rings) WithLimit(v int64) func(*RingsRequest) {
	return func(req *RingsRequest) { req.limit = v }
}
