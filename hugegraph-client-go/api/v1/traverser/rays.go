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

func newRaysFunc(t api.Transport) Rays {
	return func(o ...func(*RaysRequest)) (*RaysResponse, error) {
		var r = RaysRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type Rays func(o ...func(*RaysRequest)) (*RaysResponse, error)

type RaysRequest struct {
	ctx       context.Context
	source    string
	maxDepth  int
	direction string
	label     string
	maxDegree int64
	capacity  int64
	limit     int64
}

type RaysResponse struct {
	StatusCode int              `json:"-"`
	Header     http.Header      `json:"-"`
	Body       io.ReadCloser    `json:"-"`
	Data       RaysResponseData `json:"-"`
}

type RaysResponseData struct {
	Rays []RaysPath `json:"rays"`
}

type RaysPath struct {
	Objects []interface{} `json:"objects"`
}

func (r RaysRequest) Do(ctx context.Context, transport api.Transport) (*RaysResponse, error) {
	if len(r.source) == 0 {
		return nil, errors.New("rays: source is required")
	}
	if r.maxDepth <= 0 {
		return nil, errors.New("rays: max_depth must be > 0")
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
	if r.maxDegree > 0 {
		params.Add("max_degree", int64ToString(r.maxDegree))
	}
	if r.capacity > 0 {
		params.Add("capacity", int64ToString(r.capacity))
	}
	if r.limit > 0 {
		params.Add("limit", int64ToString(r.limit))
	}

	url := getURL(transport, "rays")
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

	data := RaysResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &RaysResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (r Rays) WithContext(ctx context.Context) func(*RaysRequest) {
	return func(req *RaysRequest) { req.ctx = ctx }
}
func (r Rays) WithSource(v string) func(*RaysRequest) {
	return func(req *RaysRequest) { req.source = v }
}
func (r Rays) WithMaxDepth(v int) func(*RaysRequest) {
	return func(req *RaysRequest) { req.maxDepth = v }
}
func (r Rays) WithDirection(v string) func(*RaysRequest) {
	return func(req *RaysRequest) { req.direction = v }
}
func (r Rays) WithLabel(v string) func(*RaysRequest) {
	return func(req *RaysRequest) { req.label = v }
}
func (r Rays) WithMaxDegree(v int64) func(*RaysRequest) {
	return func(req *RaysRequest) { req.maxDegree = v }
}
func (r Rays) WithCapacity(v int64) func(*RaysRequest) {
	return func(req *RaysRequest) { req.capacity = v }
}
func (r Rays) WithLimit(v int64) func(*RaysRequest) {
	return func(req *RaysRequest) { req.limit = v }
}
