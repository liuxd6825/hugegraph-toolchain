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

// EdgesPair groups the batch and shard-based edge query APIs.
//
// 3.2.23.1 Edges (List by ids)
// 3.2.23.2 Edges (Shards)
// 3.2.23.3 Edges (Scan)
type EdgesPair struct {
	List   EdgesList
	Shards EdgesShards
	Scan   EdgesScan
}

func newEdgesPair(t api.Transport) EdgesPair {
	return EdgesPair{
		List:   newEdgesListFunc(t),
		Shards: newEdgesShardsFunc(t),
		Scan:   newEdgesScanFunc(t),
	}
}

// ----- List ---------------------------------------------------------------

// EdgesList is the fluent method binding for the
// "GET /traversers/edges" endpoint (batch-by-id query).
type EdgesList func(o ...func(*EdgesListRequest)) (*EdgesListResponse, error)

func newEdgesListFunc(t api.Transport) EdgesList {
	return func(o ...func(*EdgesListRequest)) (*EdgesListResponse, error) {
		var r = EdgesListRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type EdgesListRequest struct {
	ctx context.Context
	ids []string
}

type EdgesListResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       graph.Edges
}

func (e EdgesList) WithIDs(ids []string) func(*EdgesListRequest) {
	return func(r *EdgesListRequest) { r.ids = ids }
}
func (e EdgesList) WithContext(ctx context.Context) func(*EdgesListRequest) {
	return func(r *EdgesListRequest) { r.ctx = ctx }
}

func (r EdgesListRequest) Do(ctx context.Context, transport api.Transport) (*EdgesListResponse, error) {
	if len(r.ids) == 0 {
		return nil, errors.New("ids is required (at least one edge id)")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	for _, id := range r.ids {
		params.Add("ids", id)
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/edges", params, nil)
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

	data := graph.Edges{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &EdgesListResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}

// ----- Shards -------------------------------------------------------------

// EdgesShards is the fluent method binding for the
// "GET /traversers/edges/shards" endpoint.
type EdgesShards func(o ...func(*EdgesShardsRequest)) (*EdgesShardsResponse, error)

func newEdgesShardsFunc(t api.Transport) EdgesShards {
	return func(o ...func(*EdgesShardsRequest)) (*EdgesShardsResponse, error) {
		var r = EdgesShardsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type EdgesShardsRequest struct {
	ctx       context.Context
	splitSize int64
}

type EdgesShardsResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       EdgesShardsResponseData
}

type EdgesShardsResponseData struct {
	Shards []graph.Shard `json:"shards"`
}

func (e EdgesShards) WithSplitSize(v int64) func(*EdgesShardsRequest) {
	return func(r *EdgesShardsRequest) { r.splitSize = v }
}
func (e EdgesShards) WithContext(ctx context.Context) func(*EdgesShardsRequest) {
	return func(r *EdgesShardsRequest) { r.ctx = ctx }
}

func (r EdgesShardsRequest) Do(ctx context.Context, transport api.Transport) (*EdgesShardsResponse, error) {
	if r.splitSize <= 0 {
		return nil, errors.New("split_size must be > 0")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	params.Set("split_size", strconv.FormatInt(r.splitSize, 10))

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/edges/shards", params, nil)
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

	data := EdgesShardsResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &EdgesShardsResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}

// ----- Scan ---------------------------------------------------------------

// EdgesScan is the fluent method binding for the
// "GET /traversers/edges/scan" endpoint.
type EdgesScan func(o ...func(*EdgesScanRequest)) (*EdgesScanResponse, error)

func newEdgesScanFunc(t api.Transport) EdgesScan {
	return func(o ...func(*EdgesScanRequest)) (*EdgesScanResponse, error) {
		var r = EdgesScanRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type EdgesScanRequest struct {
	ctx       context.Context
	start     string
	end       string
	page      string
	pageLimit int64
}

type EdgesScanResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       graph.Edges
}

func (e EdgesScan) WithStart(s string) func(*EdgesScanRequest) {
	return func(r *EdgesScanRequest) { r.start = s }
}
func (e EdgesScan) WithEnd(s string) func(*EdgesScanRequest) {
	return func(r *EdgesScanRequest) { r.end = s }
}
func (e EdgesScan) WithPage(s string) func(*EdgesScanRequest) {
	return func(r *EdgesScanRequest) { r.page = s }
}
func (e EdgesScan) WithPageLimit(v int64) func(*EdgesScanRequest) {
	return func(r *EdgesScanRequest) { r.pageLimit = v }
}
func (e EdgesScan) WithContext(ctx context.Context) func(*EdgesScanRequest) {
	return func(r *EdgesScanRequest) { r.ctx = ctx }
}

func (r EdgesScanRequest) Do(ctx context.Context, transport api.Transport) (*EdgesScanResponse, error) {
	if r.start == "" || r.end == "" {
		return nil, errors.New("start and end are required")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	params.Set("start", r.start)
	params.Set("end", r.end)
	if r.page != "" {
		params.Set("page", r.page)
	}
	if r.pageLimit > 0 {
		params.Set("page_limit", strconv.FormatInt(r.pageLimit, 10))
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/edges/scan", params, nil)
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

	data := graph.Edges{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &EdgesScanResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
