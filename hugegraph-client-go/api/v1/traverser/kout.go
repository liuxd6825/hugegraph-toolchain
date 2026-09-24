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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/apache/hugegraph-toolchain/hugegraph-client-go/api"
)

func newKoutBasicFunc(t api.Transport) KoutBasic {
	return func(o ...func(*KoutBasicRequest)) (*KoutBasicResponse, error) {
		var r = KoutBasicRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type KoutBasic func(o ...func(*KoutBasicRequest)) (*KoutBasicResponse, error)

type KoutBasicRequest struct {
	ctx       context.Context
	source    string
	direction string
	maxDepth  int
	label     string
	nearest   bool
	maxDegree int64
	capacity  int64
	limit     int64
}

type KoutBasicResponse struct {
	StatusCode int                   `json:"-"`
	Header     http.Header           `json:"-"`
	Body       io.ReadCloser         `json:"-"`
	Data       KoutBasicResponseData `json:"-"`
}

type KoutBasicResponseData struct {
	Vertices []interface{} `json:"vertices"`
}

func (r KoutBasicRequest) Do(ctx context.Context, transport api.Transport) (*KoutBasicResponse, error) {
	if len(r.source) == 0 {
		return nil, errors.New("kout_basic: source is required")
	}
	if r.maxDepth <= 0 {
		return nil, errors.New("kout_basic: max_depth must be > 0")
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
	if !r.nearest {
		params.Add("nearest", boolToString(false))
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

	url := getURL(transport, "kout")
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

	data := KoutBasicResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &KoutBasicResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (k KoutBasic) WithContext(ctx context.Context) func(*KoutBasicRequest) {
	return func(r *KoutBasicRequest) { r.ctx = ctx }
}
func (k KoutBasic) WithSource(v string) func(*KoutBasicRequest) {
	return func(r *KoutBasicRequest) { r.source = v }
}
func (k KoutBasic) WithDirection(v string) func(*KoutBasicRequest) {
	return func(r *KoutBasicRequest) { r.direction = v }
}
func (k KoutBasic) WithMaxDepth(v int) func(*KoutBasicRequest) {
	return func(r *KoutBasicRequest) { r.maxDepth = v }
}
func (k KoutBasic) WithLabel(v string) func(*KoutBasicRequest) {
	return func(r *KoutBasicRequest) { r.label = v }
}
func (k KoutBasic) WithNearest(v bool) func(*KoutBasicRequest) {
	return func(r *KoutBasicRequest) { r.nearest = v }
}
func (k KoutBasic) WithMaxDegree(v int64) func(*KoutBasicRequest) {
	return func(r *KoutBasicRequest) { r.maxDegree = v }
}
func (k KoutBasic) WithCapacity(v int64) func(*KoutBasicRequest) {
	return func(r *KoutBasicRequest) { r.capacity = v }
}
func (k KoutBasic) WithLimit(v int64) func(*KoutBasicRequest) {
	return func(r *KoutBasicRequest) { r.limit = v }
}

func newKoutAdvancedFunc(t api.Transport) KoutAdvanced {
	return func(o ...func(*KoutAdvancedRequest)) (*KoutAdvancedResponse, error) {
		var r = KoutAdvancedRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type KoutAdvanced func(o ...func(*KoutAdvancedRequest)) (*KoutAdvancedResponse, error)

type KoutAdvancedRequest struct {
	ctx     context.Context
	body    io.Reader
	reqData KoutAdvancedRequestData
}

type KoutAdvancedRequestData struct {
	// 起始顶点 id，必填项
	Source string `json:"source,omitempty"`
	// 从起始点出发的 Steps，必填项，结构如下
	Steps Steps `json:"steps,omitempty"`
	// 步数，必填项
	MaxDepth int `json:"max_depth,omitempty"`
	// nearest 为 true 时，代表起始顶点到达结果顶点的最短路径长度为 depth，不存在更短的路径；nearest 为 false 时，
	// 代表起始顶点到结果顶点有一条长度为 depth 的路径（未必最短且可以有环），选填项，默认为 true
	Nearest bool `json:"nearest,omitempty"`
	// true 表示只统计结果的数目，不返回具体结果；false 表示返回具体的结果，默认为 false
	CountOnly bool `json:"count_only,omitempty"`
	// 遍历过程中最大的访问的顶点数目，选填项，默认为 10000000
	Capacity *int64 `json:"capacity,omitempty"`
	// 返回的顶点的最大数目，选填项，默认为 10000000
	Limit *int64 `json:"limit,omitempty"`
	// 选填项，默认为 false： 如果设置为 true，则结果将包含所有顶点的完整信息，即路径中的所有顶点.
	// 当 with_path 为 true 时，将返回所有路径中的顶点的完整信息 	当 with_path 为 false 时，返回所有邻居顶点的完整信息
	// 如果设置为 false，则仅返回顶点的 id
	WithVertex bool `json:"with_vertex,omitempty"`
	// true 表示返回起始点到每个邻居的最短路径，false 表示不返回起始点到每个邻居的最短路径，选填项，默认为 false
	WithPath bool `json:"with_path,omitempty"`
	// 选填项，默认为 false： 如果设置为 true，则结果将包含所有边的完整信息，即路径中的所有边
	// 当 with_path 为 true 时，将返回所有路径中的边的完整信息;  当 with_path 为 false 时，不返回任何信息
	// 如果设置为 false，则仅返回边的 id
	WithEdge bool `json:"with_edge,omitempty"`
	// 遍历方式，可选择“breadth_first_search”或“depth_first_search”作为参数，默认为“breadth_first_search”
	TraverseMode TraverseMode `json:"traverse_mode,omitempty"`
}
type TraverseMode string

const (
	TraverseMode_DepthFirstSearch   TraverseMode = "depth_first_search"
	TraverseMode_BreadthFirstSearch TraverseMode = "breadth_first_search"
)

func NewInt64(v int64) *int64 {
	return &v
}

func NewBool(v bool) *bool {
	return &v
}

type KoutAdvancedResponse struct {
	StatusCode int                      `json:"-"`
	Header     http.Header              `json:"-"`
	Body       io.ReadCloser            `json:"-"`
	Data       KoutAdvancedResponseData `json:"-"`
}

type KoutAdvancedResponseData struct {
	Size     int64         `json:"size,omitempty"`
	Kout     []interface{} `json:"kout,omitempty"`
	Paths    []KoutPath    `json:"paths,omitempty"`
	Vertices []any         `json:"vertices,omitempty"`
	Edges    []any         `json:"edges,omitempty"`
}

type KoutPath struct {
	Objects []interface{} `json:"objects,omitempty"`
}

func (r KoutAdvancedRequest) Do(ctx context.Context, transport api.Transport) (*KoutAdvancedResponse, error) {
	if r.reqData.Source == "" {
		return nil, errors.New("kout_advanced: Source is required")
	}
	if r.reqData.MaxDepth <= 0 {
		return nil, errors.New("kout_advanced: MaxDepth must be > 0")
	}
	if r.reqData.Steps.Direction == "" {
		return nil, errors.New("kout_advanced: Steps.Direction is required")
	}

	if r.body == nil {
		byteBody, err := json.Marshal(&r.reqData)
		if err != nil {
			return nil, err
		}
		body := string(byteBody)
		fmt.Println(body)
		r.body = strings.NewReader(body)
	}

	url := getURL(transport, "kout")
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

	data := KoutAdvancedResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &KoutAdvancedResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (k KoutAdvanced) WithContext(ctx context.Context) func(*KoutAdvancedRequest) {
	return func(r *KoutAdvancedRequest) { r.ctx = ctx }
}
func (k KoutAdvanced) WithBody(body io.Reader) func(*KoutAdvancedRequest) {
	return func(r *KoutAdvancedRequest) { r.body = body }
}
func (k KoutAdvanced) WithReqData(data KoutAdvancedRequestData) func(*KoutAdvancedRequest) {
	return func(r *KoutAdvancedRequest) { r.reqData = data }
}
