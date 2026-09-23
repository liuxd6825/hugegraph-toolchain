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

func newEdgesByIDFunc(t api.Transport) EdgesByID {
	return func(o ...func(*EdgesByIDRequest)) (*EdgesByIDResponse, error) {
		var r = EdgesByIDRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type EdgesByID func(o ...func(*EdgesByIDRequest)) (*EdgesByIDResponse, error)

type EdgesByIDRequest struct {
	ctx context.Context
	ids []string
}

type EdgesByIDResponse struct {
	StatusCode int                   `json:"-"`
	Header     http.Header           `json:"-"`
	Body       io.ReadCloser         `json:"-"`
	Data       EdgesByIDResponseData `json:"-"`
}

type EdgesByIDResponseData struct {
	Edges []model.Edge[any] `json:"edges"`
}

func (r EdgesByIDRequest) Do(ctx context.Context, transport api.Transport) (*EdgesByIDResponse, error) {
	if len(r.ids) == 0 {
		return nil, errors.New("edges_by_id: ids is required")
	}

	params := &url.Values{}
	for _, id := range r.ids {
		params.Add("ids", quoteVertexID(id))
	}

	url := buildTraverserURL(transport, "edges")
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

	data := EdgesByIDResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &EdgesByIDResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (e EdgesByID) WithContext(ctx context.Context) func(*EdgesByIDRequest) {
	return func(r *EdgesByIDRequest) { r.ctx = ctx }
}
func (e EdgesByID) WithIDs(ids []string) func(*EdgesByIDRequest) {
	return func(r *EdgesByIDRequest) { r.ids = ids }
}

func newEdgesShardsFunc(t api.Transport) EdgesShards {
	return func(o ...func(*EdgesShardsRequest)) (*EdgesShardsResponse, error) {
		var r = EdgesShardsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type EdgesShards func(o ...func(*EdgesShardsRequest)) (*EdgesShardsResponse, error)

type EdgesShardsRequest struct {
	ctx       context.Context
	splitSize int64
}

type EdgesShardsResponse struct {
	StatusCode int                     `json:"-"`
	Header     http.Header             `json:"-"`
	Body       io.ReadCloser           `json:"-"`
	Data       EdgesShardsResponseData `json:"-"`
}

type EdgesShardsResponseData struct {
	Shards []Shard `json:"shards"`
}

func (r EdgesShardsRequest) Do(ctx context.Context, transport api.Transport) (*EdgesShardsResponse, error) {
	if r.splitSize <= 0 {
		return nil, errors.New("edges_shards: split_size must be > 0")
	}

	params := &url.Values{}
	params.Add("split_size", int64ToString(r.splitSize))

	url := buildTraverserURL(transport, "edges/shards")
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

	data := EdgesShardsResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &EdgesShardsResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (es EdgesShards) WithContext(ctx context.Context) func(*EdgesShardsRequest) {
	return func(r *EdgesShardsRequest) { r.ctx = ctx }
}
func (es EdgesShards) WithSplitSize(v int64) func(*EdgesShardsRequest) {
	return func(r *EdgesShardsRequest) { r.splitSize = v }
}

func newEdgesScanFunc(t api.Transport) EdgesScan {
	return func(o ...func(*EdgesScanRequest)) (*EdgesScanResponse, error) {
		var r = EdgesScanRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type EdgesScan func(o ...func(*EdgesScanRequest)) (*EdgesScanResponse, error)

type EdgesScanRequest struct {
	ctx       context.Context
	start     string
	end       string
	page      string
	pageLimit int
}

type EdgesScanResponse struct {
	StatusCode int                   `json:"-"`
	Header     http.Header           `json:"-"`
	Body       io.ReadCloser         `json:"-"`
	Data       EdgesScanResponseData `json:"-"`
}

type EdgesScanResponseData struct {
	Edges []model.Edge[any] `json:"edges"`
	Page  string            `json:"page,omitempty"`
}

func (r EdgesScanRequest) Do(ctx context.Context, transport api.Transport) (*EdgesScanResponse, error) {
	if len(r.start) == 0 {
		return nil, errors.New("edges_scan: start is required")
	}
	if len(r.end) == 0 {
		return nil, errors.New("edges_scan: end is required")
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

	url := buildTraverserURL(transport, "edges/scan")
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

	data := EdgesScanResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &EdgesScanResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (es EdgesScan) WithContext(ctx context.Context) func(*EdgesScanRequest) {
	return func(r *EdgesScanRequest) { r.ctx = ctx }
}
func (es EdgesScan) WithStart(v string) func(*EdgesScanRequest) {
	return func(r *EdgesScanRequest) { r.start = v }
}
func (es EdgesScan) WithEnd(v string) func(*EdgesScanRequest) {
	return func(r *EdgesScanRequest) { r.end = v }
}
func (es EdgesScan) WithPage(v string) func(*EdgesScanRequest) {
	return func(r *EdgesScanRequest) { r.page = v }
}
func (es EdgesScan) WithPageLimit(v int) func(*EdgesScanRequest) {
	return func(r *EdgesScanRequest) { r.pageLimit = v }
}
