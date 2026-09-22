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
	structtraverser "github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/traverser"
)

// VerticesPair groups the batch and shard-based vertex query APIs.
//
// 3.2.22.1 Vertices (List by ids)
// 3.2.22.2 Vertices (Shards)
// 3.2.22.3 Vertices (Scan)
type VerticesPair struct {
	List   VerticesList
	Shards VerticesShards
	Scan   VerticesScan
}

func newVerticesPair(t api.Transport) VerticesPair {
	return VerticesPair{
		List:   newVerticesListFunc(t),
		Shards: newVerticesShardsFunc(t),
		Scan:   newVerticesScanFunc(t),
	}
}

// ----- List ---------------------------------------------------------------

// VerticesList is the fluent method binding for the
// "GET /traversers/vertices" endpoint (batch-by-id query).
type VerticesList func(o ...func(*VerticesListRequest)) (*VerticesListResponse, error)

func newVerticesListFunc(t api.Transport) VerticesList {
	return func(o ...func(*VerticesListRequest)) (*VerticesListResponse, error) {
		var r = VerticesListRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type VerticesListRequest struct {
	ctx context.Context
	ids []string
}

type VerticesListResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       graph.Vertices
}

func (v VerticesList) WithIDs(ids []string) func(*VerticesListRequest) {
	return func(r *VerticesListRequest) { r.ids = ids }
}

func (v VerticesList) WithContext(ctx context.Context) func(*VerticesListRequest) {
	return func(r *VerticesListRequest) { r.ctx = ctx }
}

func (r VerticesListRequest) Do(ctx context.Context, transport api.Transport) (*VerticesListResponse, error) {
	if len(r.ids) == 0 {
		return nil, errors.New("ids is required (at least one vertex id)")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	for _, id := range r.ids {
		params.Add("ids", graph.FormatVertexID(id))
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/vertices", params, nil)
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

	data := graph.Vertices{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &VerticesListResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}

// ----- Shards -------------------------------------------------------------

// VerticesShards is the fluent method binding for the
// "GET /traversers/vertices/shards" endpoint.
type VerticesShards func(o ...func(*VerticesShardsRequest)) (*VerticesShardsResponse, error)

func newVerticesShardsFunc(t api.Transport) VerticesShards {
	return func(o ...func(*VerticesShardsRequest)) (*VerticesShardsResponse, error) {
		var r = VerticesShardsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type VerticesShardsRequest struct {
	ctx       context.Context
	splitSize int64
}

type VerticesShardsResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       VerticesShardsResponseData
}

type VerticesShardsResponseData struct {
	Shards []graph.Shard `json:"shards"`
}

func (v VerticesShards) WithSplitSize(size int64) func(*VerticesShardsRequest) {
	return func(r *VerticesShardsRequest) { r.splitSize = size }
}
func (v VerticesShards) WithContext(ctx context.Context) func(*VerticesShardsRequest) {
	return func(r *VerticesShardsRequest) { r.ctx = ctx }
}

func (r VerticesShardsRequest) Do(ctx context.Context, transport api.Transport) (*VerticesShardsResponse, error) {
	if r.splitSize <= 0 {
		return nil, errors.New("split_size must be > 0")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	params.Set("split_size", strconv.FormatInt(r.splitSize, 10))

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/vertices/shards", params, nil)
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

	data := VerticesShardsResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &VerticesShardsResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}

// ----- Scan ---------------------------------------------------------------

// VerticesScan is the fluent method binding for the
// "GET /traversers/vertices/scan" endpoint.
type VerticesScan func(o ...func(*VerticesScanRequest)) (*VerticesScanResponse, error)

func newVerticesScanFunc(t api.Transport) VerticesScan {
	return func(o ...func(*VerticesScanRequest)) (*VerticesScanResponse, error) {
		var r = VerticesScanRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type VerticesScanRequest struct {
	ctx       context.Context
	start     string
	end       string
	page      string
	pageLimit int64
	args      *structtraverser.VerticesArgs
}

type VerticesScanResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       graph.Vertices
}

func (v VerticesScan) WithStart(s string) func(*VerticesScanRequest) {
	return func(r *VerticesScanRequest) { r.start = s }
}
func (v VerticesScan) WithEnd(s string) func(*VerticesScanRequest) {
	return func(r *VerticesScanRequest) { r.end = s }
}
func (v VerticesScan) WithPage(s string) func(*VerticesScanRequest) {
	return func(r *VerticesScanRequest) { r.page = s }
}
func (v VerticesScan) WithPageLimit(limit int64) func(*VerticesScanRequest) {
	return func(r *VerticesScanRequest) { r.pageLimit = limit }
}
func (v VerticesScan) WithArgs(args *structtraverser.VerticesArgs) func(*VerticesScanRequest) {
	return func(r *VerticesScanRequest) { r.args = args }
}
func (v VerticesScan) WithContext(ctx context.Context) func(*VerticesScanRequest) {
	return func(r *VerticesScanRequest) { r.ctx = ctx }
}

func (r VerticesScanRequest) Do(ctx context.Context, transport api.Transport) (*VerticesScanResponse, error) {
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

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/vertices/scan", params, nil)
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

	data := graph.Vertices{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &VerticesScanResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
