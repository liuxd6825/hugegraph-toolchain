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

func newPathsBasicFunc(t api.Transport) PathsBasic {
	return func(o ...func(*PathsBasicRequest)) (*PathsBasicResponse, error) {
		var r = PathsBasicRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type PathsBasic func(o ...func(*PathsBasicRequest)) (*PathsBasicResponse, error)

type PathsBasicRequest struct {
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

type PathsBasicResponse struct {
	StatusCode int                    `json:"-"`
	Header     http.Header            `json:"-"`
	Body       io.ReadCloser          `json:"-"`
	Data       PathsBasicResponseData `json:"-"`
}

type PathsBasicResponseData struct {
	Paths []KoutPath `json:"paths"`
}

func (r PathsBasicRequest) Do(ctx context.Context, transport api.Transport) (*PathsBasicResponse, error) {
	if len(r.source) == 0 {
		return nil, errors.New("paths_basic: source is required")
	}
	if len(r.target) == 0 {
		return nil, errors.New("paths_basic: target is required")
	}
	if r.maxDepth <= 0 {
		return nil, errors.New("paths_basic: max_depth must be > 0")
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

	url := buildTraverserURL(transport, "paths")
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

	data := PathsBasicResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &PathsBasicResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (p PathsBasic) WithContext(ctx context.Context) func(*PathsBasicRequest) {
	return func(r *PathsBasicRequest) { r.ctx = ctx }
}
func (p PathsBasic) WithSource(v string) func(*PathsBasicRequest) {
	return func(r *PathsBasicRequest) { r.source = v }
}
func (p PathsBasic) WithTarget(v string) func(*PathsBasicRequest) {
	return func(r *PathsBasicRequest) { r.target = v }
}
func (p PathsBasic) WithDirection(v string) func(*PathsBasicRequest) {
	return func(r *PathsBasicRequest) { r.direction = v }
}
func (p PathsBasic) WithLabel(v string) func(*PathsBasicRequest) {
	return func(r *PathsBasicRequest) { r.label = v }
}
func (p PathsBasic) WithMaxDepth(v int) func(*PathsBasicRequest) {
	return func(r *PathsBasicRequest) { r.maxDepth = v }
}
func (p PathsBasic) WithMaxDegree(v int64) func(*PathsBasicRequest) {
	return func(r *PathsBasicRequest) { r.maxDegree = v }
}
func (p PathsBasic) WithCapacity(v int64) func(*PathsBasicRequest) {
	return func(r *PathsBasicRequest) { r.capacity = v }
}
func (p PathsBasic) WithLimit(v int64) func(*PathsBasicRequest) {
	return func(r *PathsBasicRequest) { r.limit = v }
}

func newPathsAdvancedFunc(t api.Transport) PathsAdvanced {
	return func(o ...func(*PathsAdvancedRequest)) (*PathsAdvancedResponse, error) {
		var r = PathsAdvancedRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type PathsAdvanced func(o ...func(*PathsAdvancedRequest)) (*PathsAdvancedResponse, error)

type PathsAdvancedRequest struct {
	ctx     context.Context
	body    io.Reader
	reqData PathsAdvancedRequestData
}

type PathsAdvancedRequestData struct {
	Sources    []string `json:"sources"`
	Targets    []string `json:"targets"`
	Step       Step     `json:"step"`
	MaxDepth   int      `json:"max_depth"`
	Nearest    *bool    `json:"nearest,omitempty"`
	Capacity   *int64   `json:"capacity,omitempty"`
	Limit      *int64   `json:"limit,omitempty"`
	WithVertex *bool    `json:"with_vertex,omitempty"`
}

type PathsAdvancedResponse struct {
	StatusCode int                       `json:"-"`
	Header     http.Header               `json:"-"`
	Body       io.ReadCloser             `json:"-"`
	Data       PathsAdvancedResponseData `json:"-"`
}

type PathsAdvancedResponseData struct {
	Paths    []KoutPath               `json:"paths"`
	Vertices []map[string]interface{} `json:"vertices,omitempty"`
}

func (r PathsAdvancedRequest) Do(ctx context.Context, transport api.Transport) (*PathsAdvancedResponse, error) {
	if r.reqData.Sources == nil {
		return nil, errors.New("paths_advanced: sources is required")
	}
	if r.reqData.Targets == nil {
		return nil, errors.New("paths_advanced: targets is required")
	}
	if r.reqData.MaxDepth <= 0 {
		return nil, errors.New("paths_advanced: max_depth must be > 0")
	}

	if r.body == nil {
		byteBody, err := json.Marshal(&r.reqData)
		if err != nil {
			return nil, err
		}
		r.body = strings.NewReader(string(byteBody))
	}

	url := buildTraverserURL(transport, "paths")
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

	data := PathsAdvancedResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &PathsAdvancedResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (p PathsAdvanced) WithContext(ctx context.Context) func(*PathsAdvancedRequest) {
	return func(r *PathsAdvancedRequest) { r.ctx = ctx }
}
func (p PathsAdvanced) WithBody(body io.Reader) func(*PathsAdvancedRequest) {
	return func(r *PathsAdvancedRequest) { r.body = body }
}
func (p PathsAdvanced) WithReqData(data PathsAdvancedRequestData) func(*PathsAdvancedRequest) {
	return func(r *PathsAdvancedRequest) { r.reqData = data }
}

func newCustomizedPathsFunc(t api.Transport) CustomizedPaths {
	return func(o ...func(*CustomizedPathsRequest)) (*CustomizedPathsResponse, error) {
		var r = CustomizedPathsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type CustomizedPaths func(o ...func(*CustomizedPathsRequest)) (*CustomizedPathsResponse, error)

type CustomizedPathsRequest struct {
	ctx     context.Context
	body    io.Reader
	reqData CustomizedPathsRequestData
}

type CustomizedPathsRequestData struct {
	Sources    Sources `json:"sources"`
	Steps      Steps   `json:"steps"`
	SortBy     string  `json:"sort_by,omitempty"`
	Capacity   int64   `json:"capacity"`
	Limit      int64   `json:"limit"`
	WithVertex bool    `json:"with_vertex"`
}

type Sources struct {
	Ids        []string       `json:"ids,omitempty"`
	Label      string         `json:"label,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

type CustomizedPathsResponse struct {
	StatusCode int                         `json:"-"`
	Header     http.Header                 `json:"-"`
	Body       io.ReadCloser               `json:"-"`
	Data       CustomizedPathsResponseData `json:"-"`
}

type CustomizedPathsResponseData struct {
	Paths    []CustomizedPath         `json:"paths"`
	Vertices []map[string]interface{} `json:"vertices,omitempty"`
}

type CustomizedPath struct {
	Objects []interface{} `json:"objects"`
	Weights []float64     `json:"weights,omitempty"`
}

func (r CustomizedPathsRequest) Do(ctx context.Context, transport api.Transport) (*CustomizedPathsResponse, error) {

	if r.body == nil {
		byteBody, err := json.Marshal(&r.reqData)
		if err != nil {
			return nil, err
		}
		r.body = strings.NewReader(string(byteBody))
	}

	url := buildTraverserURL(transport, "customizedpaths")
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

	data := CustomizedPathsResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &CustomizedPathsResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (c CustomizedPaths) WithContext(ctx context.Context) func(*CustomizedPathsRequest) {
	return func(r *CustomizedPathsRequest) { r.ctx = ctx }
}
func (c CustomizedPaths) WithBody(body io.Reader) func(*CustomizedPathsRequest) {
	return func(r *CustomizedPathsRequest) { r.body = body }
}
func (c CustomizedPaths) WithReqData(data CustomizedPathsRequestData) func(*CustomizedPathsRequest) {
	return func(r *CustomizedPathsRequest) { r.reqData = data }
}

func newTemplatePathsFunc(t api.Transport) TemplatePaths {
	return func(o ...func(*TemplatePathsRequest)) (*TemplatePathsResponse, error) {
		var r = TemplatePathsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type TemplatePaths func(o ...func(*TemplatePathsRequest)) (*TemplatePathsResponse, error)

type TemplatePathsReqData struct {
	Sources    interface{} `json:"sources"`
	Targets    interface{} `json:"targets"`
	Steps      interface{} `json:"steps"`
	WithRing   bool        `json:"with_ring"`
	Capacity   int64       `json:"capacity"`
	Limit      int64       `json:"limit"`
	WithVertex bool        `json:"with_vertex"`
}

type TemplatePathsRequest struct {
	ctx     context.Context
	body    io.Reader
	reqData TemplatePathsReqData
}

type TemplatePathsResponse struct {
	StatusCode int                       `json:"-"`
	Header     http.Header               `json:"-"`
	Body       io.ReadCloser             `json:"-"`
	Data       TemplatePathsResponseData `json:"-"`
}

type TemplatePathsResponseData struct {
	Paths    []KoutPath               `json:"paths"`
	Vertices []map[string]interface{} `json:"vertices,omitempty"`
}

func (r TemplatePathsRequest) Do(ctx context.Context, transport api.Transport) (*TemplatePathsResponse, error) {
	if r.reqData.Sources == nil {
		return nil, errors.New("template_paths: sources is required")
	}
	if r.reqData.Targets == nil {
		return nil, errors.New("template_paths: targets is required")
	}
	if r.reqData.Steps == nil {
		return nil, errors.New("template_paths: steps is required")
	}

	if r.body == nil {
		byteBody, err := json.Marshal(&r.reqData)
		if err != nil {
			return nil, err
		}
		r.body = strings.NewReader(string(byteBody))
	}

	url := buildTraverserURL(transport, "templatepaths")
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

	data := TemplatePathsResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &TemplatePathsResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (t TemplatePaths) WithContext(ctx context.Context) func(*TemplatePathsRequest) {
	return func(r *TemplatePathsRequest) { r.ctx = ctx }
}
func (t TemplatePaths) WithBody(body io.Reader) func(*TemplatePathsRequest) {
	return func(r *TemplatePathsRequest) { r.body = body }
}
func (t TemplatePaths) WithReqData(data TemplatePathsReqData) func(*TemplatePathsRequest) {
	return func(r *TemplatePathsRequest) { r.reqData = data }
}
