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
	Sources    SourcesTargets `json:"sources"`
	Targets    SourcesTargets `json:"targets"`
	Step       Step           `json:"step"`
	MaxDepth   int            `json:"max_depth"`
	Nearest    *bool          `json:"nearest,omitempty"`
	Capacity   *int64         `json:"capacity,omitempty"`
	Limit      *int64         `json:"limit,omitempty"`
	WithVertex *bool          `json:"with_vertex,omitempty"`
}

type PathsAdvancedResponse struct {
	StatusCode int                       `json:"-"`
	Header     http.Header               `json:"-"`
	Body       io.ReadCloser             `json:"-"`
	Data       PathsAdvancedResponseData `json:"-"`
}

type PathsAdvancedResponseData struct {
	Paths    []KoutPath `json:"paths"`
	Vertices []any      `json:"vertices,omitempty"`
}

func (r PathsAdvancedRequest) Do(ctx context.Context, transport api.Transport) (*PathsAdvancedResponse, error) {
	if len(r.reqData.Sources.Ids) == 0 {
		return nil, errors.New("paths_advanced: sources.ids is required")
	}
	if len(r.reqData.Targets.Ids) == 0 {
		return nil, errors.New("paths_advanced: targets.ids is required")
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

// CustomizedPaths
// 根据一批起始顶点、边规则（包括方向、边的类型和属性过滤）和最大深度等条件查找符合条件的所有的路径
type CustomizedPaths func(o ...func(*CustomizedPathsRequest)) (*CustomizedPathsResponse, error)

type CustomizedPathsRequest struct {
	ctx     context.Context
	body    io.Reader
	reqData CustomizedPathsRequestData
}

type CustomizedPathsRequestData struct {
	// 定义起始顶点，必填项，指定方式包括：
	Sources SourcesTargets `json:"sources"`
	// 表示从起始顶点走过的路径规则，是一组 Step 的列表。必填项。每个 Step 的结构如下
	Steps Steps `json:"steps"`
	// sort_by：根据路径的权重排序，选填项，默认为 NONE：
	SortBy CustomizedPathsSortBy `json:"sort_by,omitempty"`
	// capacity: 遍历过程中最大的访问的顶点数目，选填项，默认为 10000000
	Capacity *int64 `json:"capacity"`
	// limit：返回的路径的最大数目，选填项，默认为 10
	Limit *int64 `json:"limit"`
	// with_vertex：true 表示返回结果包含完整的顶点信息（路径中的全部顶点），false 时表示只返回顶点 id，选填项，默认为 false
	WithVertex bool `json:"with_vertex"`
}

type CustomizedPathsSortBy string

const (
	NONE CustomizedPathsSortBy = "NONE"
	INCR CustomizedPathsSortBy = "INCR"
	DECR CustomizedPathsSortBy = "DECR"
)

func (c CustomizedPathsSortBy) NONE() CustomizedPathsSortBy {
	return NONE
}

func (c CustomizedPathsSortBy) INCR() CustomizedPathsSortBy {
	return INCR
}

func (c CustomizedPathsSortBy) DECR() CustomizedPathsSortBy {
	return DECR
}

func (c CustomizedPathsSortBy) String() string {
	return string(c)
}

type SourcesTargets struct {
	// 通过顶点 id 列表提供起始顶点
	Ids []string `json:"ids,omitempty"`
	// 如果没有指定 ids，则使用 label 和 properties 的联合条件查询起始顶点
	// 顶点的类型
	Label string `json:"label,omitempty"`
	// 通过属性的值查询起始顶点
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

// TemplatePaths
// 适合查找各种复杂的模板路径，比如 personA -(朋友)-> personB -(同学)-> personC，
// 其中"朋友"和"同学"边可以分别是最多 3 层和 4 层的情况
type TemplatePaths func(o ...func(*TemplatePathsRequest)) (*TemplatePathsResponse, error)

type TemplatePathsReqData struct {
	Sources    SourcesTargets      `json:"sources"`
	Targets    SourcesTargets      `json:"targets"`
	Steps      []TemplatePathsStep `json:"steps"`
	WithRing   bool                `json:"with_ring"`
	Capacity   int64               `json:"capacity"`
	Limit      int64               `json:"limit"`
	WithVertex bool                `json:"with_vertex"`
}

type TemplatePathsStep struct {
	// 表示边的方向（OUT,IN,BOTH），默认是 BOTH
	Direction Direction `json:"direction"`
	// 边的类型列表
	Labels []string `json:"labels"`
	// 通过属性的值过滤边
	Properties map[string]any `json:"properties"`
	// 当前 step 可以重复的次数，当为 N 时，表示从起始顶点可以经过当前 step 1-N 次
	MaxDegree *int `json:"max_degree"`
	// 用于设置查询过程中舍弃超级顶点的最小边数，即当某个顶点的邻接边数目大于
	// skip_degree 时，完全舍弃该顶点。选填项，如果开启时，需满足 skip_degree >= max_degree 约束，默认为 0 (不启用)，
	// 表示不跳过任何点 ( 	注意：开启此配置后，遍历时会尝试访问一个顶点的 skip_degree 条边，而不仅仅是 max_degree 条边，
	// 这样有额外的遍历开销，对查询性能影响可能有较大影响，请确认理解后再开启)
	SkipDegree *int `json:"skip_degree"`
	// 当前 step 可以重复的次数，当为 N 时，表示从起始顶点可以经过当前 step 1-N 次
	MaxTime *int `json:"max_time"`
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
	if len(r.reqData.Sources.Ids) == 0 {
		return nil, errors.New("template_paths: sources is required")
	}
	if len(r.reqData.Targets.Ids) == 0 {
		return nil, errors.New("template_paths: targets is required")
	}
	if len(r.reqData.Steps) == 0 {
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
