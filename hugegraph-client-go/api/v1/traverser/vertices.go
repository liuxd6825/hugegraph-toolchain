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
	"github.com/apache/hugegraph-toolchain/hugegraph-client-go/internal/model"
)

func newVerticesByIDFunc(t api.Transport) VerticesByID {
	return func(o ...func(*VerticesByIDRequest)) (*VerticesByIDResponse, error) {
		var r = VerticesByIDRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type VerticesByID func(o ...func(*VerticesByIDRequest)) (*VerticesByIDResponse, error)

type VerticesByIDRequest struct {
	ctx context.Context
	ids []string
}

type VerticesByIDResponse struct {
	StatusCode int                      `json:"-"`
	Header     http.Header              `json:"-"`
	Body       io.ReadCloser            `json:"-"`
	Data       VerticesByIDResponseData `json:"-"`
}

type VerticesByIDResponseData struct {
	Vertices []model.Vertex[any] `json:"vertices"`
}

func (r VerticesByIDRequest) Do(ctx context.Context, transport api.Transport) (*VerticesByIDResponse, error) {
	if len(r.ids) == 0 {
		return nil, errors.New("vertices_by_id: ids is required")
	}

	params := &url.Values{}
	for _, id := range r.ids {
		params.Add("ids", quoteVertexID(id))
	}

	url := buildTraverserURL(transport, "vertices")
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

	data := VerticesByIDResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &VerticesByIDResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (v VerticesByID) WithContext(ctx context.Context) func(*VerticesByIDRequest) {
	return func(r *VerticesByIDRequest) { r.ctx = ctx }
}
func (v VerticesByID) WithIDs(ids []string) func(*VerticesByIDRequest) {
	return func(r *VerticesByIDRequest) { r.ids = ids }
}

func newVerticesShardsFunc(t api.Transport) VerticesShards {
	return func(o ...func(*VerticesShardsRequest)) (*VerticesShardsResponse, error) {
		var r = VerticesShardsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type VerticesShards func(o ...func(*VerticesShardsRequest)) (*VerticesShardsResponse, error)

type VerticesShardsRequest struct {
	ctx       context.Context
	splitSize int64
}

type VerticesShardsResponse struct {
	StatusCode int                        `json:"-"`
	Header     http.Header                `json:"-"`
	Body       io.ReadCloser              `json:"-"`
	Data       VerticesShardsResponseData `json:"-"`
}

type VerticesShardsResponseData struct {
	Shards []Shard `json:"shards"`
}

type Shard struct {
	Start  string `json:"start"`
	End    string `json:"end"`
	Length int64  `json:"length"`
}

func (r VerticesShardsRequest) Do(ctx context.Context, transport api.Transport) (*VerticesShardsResponse, error) {
	if r.splitSize <= 0 {
		return nil, errors.New("vertices_shards: split_size must be > 0")
	}

	params := &url.Values{}
	params.Add("split_size", int64ToString(r.splitSize))

	url := buildTraverserURL(transport, "vertices/shards")
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

	data := VerticesShardsResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &VerticesShardsResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (vs VerticesShards) WithContext(ctx context.Context) func(*VerticesShardsRequest) {
	return func(r *VerticesShardsRequest) { r.ctx = ctx }
}
func (vs VerticesShards) WithSplitSize(v int64) func(*VerticesShardsRequest) {
	return func(r *VerticesShardsRequest) { r.splitSize = v }
}

func newVerticesScanFunc(t api.Transport) VerticesScan {
	return func(o ...func(*VerticesScanRequest)) (*VerticesScanResponse, error) {
		var r = VerticesScanRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type VerticesScan func(o ...func(*VerticesScanRequest)) (*VerticesScanResponse, error)

type VerticesScanRequest struct {
	ctx       context.Context
	start     string
	end       string
	page      string
	pageLimit int
}

type VerticesScanResponse struct {
	StatusCode int                      `json:"-"`
	Header     http.Header              `json:"-"`
	Body       io.ReadCloser            `json:"-"`
	Data       VerticesScanResponseData `json:"-"`
}

type VerticesScanResponseData struct {
	Vertices []model.Vertex[any] `json:"vertices"`
	Page     string              `json:"page,omitempty"`
}

func (r VerticesScanRequest) Do(ctx context.Context, transport api.Transport) (*VerticesScanResponse, error) {
	if len(r.start) == 0 {
		return nil, errors.New("vertices_scan: start is required")
	}
	if len(r.end) == 0 {
		return nil, errors.New("vertices_scan: end is required")
	}

	params := &url.Values{}
	params.Add("start", r.start)
	params.Add("end", r.end)
	if r.page != "" {
		params.Add("page", r.page)
	}
	if r.pageLimit > 0 {
		params.Add("page_limit", intToString(r.pageLimit))
	}

	url := buildTraverserURL(transport, "vertices/scan")
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

	data := VerticesScanResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &VerticesScanResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (vs VerticesScan) WithContext(ctx context.Context) func(*VerticesScanRequest) {
	return func(r *VerticesScanRequest) { r.ctx = ctx }
}
func (vs VerticesScan) WithStart(v string) func(*VerticesScanRequest) {
	return func(r *VerticesScanRequest) { r.start = v }
}
func (vs VerticesScan) WithEnd(v string) func(*VerticesScanRequest) {
	return func(r *VerticesScanRequest) { r.end = v }
}
func (vs VerticesScan) WithPage(v string) func(*VerticesScanRequest) {
	return func(r *VerticesScanRequest) { r.page = v }
}
func (vs VerticesScan) WithPageLimit(v int) func(*VerticesScanRequest) {
	return func(r *VerticesScanRequest) { r.pageLimit = v }
}
