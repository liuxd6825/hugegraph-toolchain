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
	"strings"

	"github.com/apache/hugegraph-toolchain/hugegraph-client-go/api"
)

func newCrosspointsFunc(t api.Transport) Crosspoints {
	return func(o ...func(*CrosspointsRequest)) (*CrosspointsResponse, error) {
		var r = CrosspointsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type Crosspoints func(o ...func(*CrosspointsRequest)) (*CrosspointsResponse, error)

type CrosspointsRequest struct {
	ctx       context.Context
	source    string
	target    string
	direction string
	label     string
	maxDepth  int
	maxDegree int64
	capacity  int64
	limit     int64
}

type CrosspointsResponse struct {
	StatusCode int                     `json:"-"`
	Header     http.Header             `json:"-"`
	Body       io.ReadCloser           `json:"-"`
	Data       CrosspointsResponseData `json:"-"`
}

type CrosspointsResponseData struct {
	Crosspoints []Crosspoint `json:"crosspoints"`
}

type Crosspoint struct {
	Crosspoint interface{}   `json:"crosspoint"`
	Objects    []interface{} `json:"objects"`
}

func (r CrosspointsRequest) Do(ctx context.Context, transport api.Transport) (*CrosspointsResponse, error) {
	if len(r.source) == 0 {
		return nil, errors.New("crosspoints: source is required")
	}
	if len(r.target) == 0 {
		return nil, errors.New("crosspoints: target is required")
	}
	if r.maxDepth <= 0 {
		return nil, errors.New("crosspoints: max_depth must be > 0")
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
	if r.capacity > 0 {
		params.Add("capacity", int64ToString(r.capacity))
	}
	if r.limit > 0 {
		params.Add("limit", int64ToString(r.limit))
	}

	url := buildTraverserURL(transport, "crosspoints")
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

	data := CrosspointsResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &CrosspointsResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (c Crosspoints) WithContext(ctx context.Context) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.ctx = ctx }
}
func (c Crosspoints) WithSource(v string) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.source = v }
}
func (c Crosspoints) WithTarget(v string) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.target = v }
}
func (c Crosspoints) WithDirection(v string) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.direction = v }
}
func (c Crosspoints) WithLabel(v string) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.label = v }
}
func (c Crosspoints) WithMaxDepth(v int) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.maxDepth = v }
}
func (c Crosspoints) WithMaxDegree(v int64) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.maxDegree = v }
}
func (c Crosspoints) WithCapacity(v int64) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.capacity = v }
}
func (c Crosspoints) WithLimit(v int64) func(*CrosspointsRequest) {
	return func(r *CrosspointsRequest) { r.limit = v }
}

func newCustomizedCrosspointsFunc(t api.Transport) CustomizedCrosspoints {
	return func(o ...func(*CustomizedCrosspointsRequest)) (*CustomizedCrosspointsResponse, error) {
		var r = CustomizedCrosspointsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type CustomizedCrosspoints func(o ...func(*CustomizedCrosspointsRequest)) (*CustomizedCrosspointsResponse, error)

type CustomizedCrosspointsRequest struct {
	ctx     context.Context
	body    io.Reader
	reqData CustomizedCrosspointsRequestData
}

type CustomizedCrosspointsRequestData struct {
	Sources      interface{} `json:"sources"`
	PathPatterns interface{} `json:"path_patterns"`
	Capacity     int64       `json:"capacity"`
	Limit        int64       `json:"limit"`
	WithPath     bool        `json:"with_path"`
	WithVertex   bool        `json:"with_vertex"`
}

type CustomizedCrosspointsResponse struct {
	StatusCode int                               `json:"-"`
	Header     http.Header                       `json:"-"`
	Body       io.ReadCloser                     `json:"-"`
	Data       CustomizedCrosspointsResponseData `json:"-"`
}

type CustomizedCrosspointsResponseData struct {
	Crosspoints []interface{}            `json:"crosspoints"`
	Paths       []CrosspointsPathList    `json:"paths,omitempty"`
	Vertices    []map[string]interface{} `json:"vertices,omitempty"`
}

type CrosspointsPathList struct {
	Objects []interface{} `json:"objects"`
}

func (r CustomizedCrosspointsRequest) Do(ctx context.Context, transport api.Transport) (*CustomizedCrosspointsResponse, error) {
	if r.reqData.Sources == nil {
		return nil, errors.New("customized_crosspoints: sources is required")
	}
	if r.reqData.PathPatterns == nil {
		return nil, errors.New("customized_crosspoints: path_patterns is required")
	}

	if r.body == nil {
		byteBody, err := json.Marshal(&r.reqData)
		if err != nil {
			return nil, err
		}
		r.body = strings.NewReader(string(byteBody))
	}

	url := buildTraverserURL(transport, "customizedcrosspoints")
	req, err := api.NewRequest("POST", url, nil, r.body)
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

	data := CustomizedCrosspointsResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &CustomizedCrosspointsResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (c CustomizedCrosspoints) WithContext(ctx context.Context) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.ctx = ctx }
}
func (c CustomizedCrosspoints) WithBody(body io.Reader) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.body = body }
}
func (c CustomizedCrosspoints) WithReqData(data CustomizedCrosspointsRequestData) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.reqData = data }
}
