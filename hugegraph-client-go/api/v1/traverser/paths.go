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
	"strings"

	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/api"
	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/graph"
	structtraverser "github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/traverser"
)

// PathsPair exposes the basic GET and advanced POST variants of the Paths API.
//
// 3.2.13 Paths (GET, basic)
// 3.2.14 Paths (POST, advanced)
type PathsPair struct {
	Get  PathsGet
	Post PathsPost
}

func newPathsPair(t api.Transport) PathsPair {
	return PathsPair{
		Get:  newPathsGetFunc(t),
		Post: newPathsPostFunc(t),
	}
}

// ----- GET ----------------------------------------------------------------

// PathsGet is the fluent method binding for the
// "GET /traversers/paths" endpoint.
type PathsGet func(o ...func(*PathsGetRequest)) (*PathsGetResponse, error)

func newPathsGetFunc(t api.Transport) PathsGet {
	return func(o ...func(*PathsGetRequest)) (*PathsGetResponse, error) {
		var r = PathsGetRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type PathsGetRequest struct {
	ctx       context.Context
	source    string
	target    string
	direction string
	label     string
	maxDepth  int
	maxDegree int
	capacity  int
	limit     int
}

type PathsGetResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       PathsGetResponseData
}

type PathsGetResponseData struct {
	Paths []PathsGetPathEntry `json:"paths"`
}

// PathsGetPathEntry mirrors a single path's {objects: [...]} payload.
type PathsGetPathEntry struct {
	Objects []interface{} `json:"objects"`
}

func (p PathsGet) WithSource(v string) func(*PathsGetRequest) {
	return func(r *PathsGetRequest) { r.source = v }
}
func (p PathsGet) WithTarget(v string) func(*PathsGetRequest) {
	return func(r *PathsGetRequest) { r.target = v }
}
func (p PathsGet) WithDirection(v string) func(*PathsGetRequest) {
	return func(r *PathsGetRequest) { r.direction = v }
}
func (p PathsGet) WithLabel(v string) func(*PathsGetRequest) {
	return func(r *PathsGetRequest) { r.label = v }
}
func (p PathsGet) WithMaxDepth(v int) func(*PathsGetRequest) {
	return func(r *PathsGetRequest) { r.maxDepth = v }
}
func (p PathsGet) WithDegree(v int) func(*PathsGetRequest) {
	return func(r *PathsGetRequest) { r.maxDegree = v }
}
func (p PathsGet) WithCapacity(v int) func(*PathsGetRequest) {
	return func(r *PathsGetRequest) { r.capacity = v }
}
func (p PathsGet) WithLimit(v int) func(*PathsGetRequest) {
	return func(r *PathsGetRequest) { r.limit = v }
}
func (p PathsGet) WithContext(ctx context.Context) func(*PathsGetRequest) {
	return func(r *PathsGetRequest) { r.ctx = ctx }
}

func (r PathsGetRequest) Do(ctx context.Context, transport api.Transport) (*PathsGetResponse, error) {
	if r.source == "" {
		return nil, errors.New("source is required")
	}
	if r.target == "" {
		return nil, errors.New("target is required")
	}
	if r.maxDepth <= 0 {
		return nil, errors.New("max_depth must be > 0")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	params.Set("source", graph.FormatVertexID(r.source))
	params.Set("target", graph.FormatVertexID(r.target))
	if r.direction != "" {
		params.Set("direction", r.direction)
	}
	if r.label != "" {
		params.Set("label", r.label)
	}
	params.Set("max_depth", strconv.Itoa(r.maxDepth))
	if r.maxDegree > 0 {
		params.Set("max_degree", strconv.Itoa(r.maxDegree))
	}
	if r.capacity > 0 {
		params.Set("capacity", strconv.Itoa(r.capacity))
	}
	if r.limit > 0 {
		params.Set("limit", strconv.Itoa(r.limit))
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/paths", params, nil)
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

	data := PathsGetResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &PathsGetResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}

// ----- POST ---------------------------------------------------------------

// PathsPost is the fluent method binding for the
// "POST /traversers/paths" endpoint.
type PathsPost func(o ...func(*PathsPostRequest)) (*PathsPostResponse, error)

func newPathsPostFunc(t api.Transport) PathsPost {
	return func(o ...func(*PathsPostRequest)) (*PathsPostResponse, error) {
		var r = PathsPostRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type PathsPostRequest struct {
	ctx     context.Context
	body    io.Reader
	request *structtraverser.PathsRequest
}

type PathsPostResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       PathsPostResponseData
}

type PathsPostResponseData struct {
	Paths    []PathsGetPathEntry      `json:"paths"`
	Vertices []map[string]interface{} `json:"vertices,omitempty"`
	Edges    []map[string]interface{} `json:"edges,omitempty"`
}

func (p PathsPost) WithRequestBody(b io.Reader) func(*PathsPostRequest) {
	return func(r *PathsPostRequest) { r.body = b }
}

func (p PathsPost) WithRequest(req *structtraverser.PathsRequest) func(*PathsPostRequest) {
	return func(r *PathsPostRequest) { r.request = req }
}

func (p PathsPost) WithContext(ctx context.Context) func(*PathsPostRequest) {
	return func(r *PathsPostRequest) { r.ctx = ctx }
}

func (r PathsPostRequest) Do(ctx context.Context, transport api.Transport) (*PathsPostResponse, error) {
	cfg := transport.GetConfig()
	body := r.body
	if body == nil {
		if r.request == nil {
			return nil, errors.New("WithRequest or WithPathsPost must be set")
		}
		if r.request.Source == nil {
			return nil, errors.New("source is required")
		}
		if r.request.Target == nil {
			return nil, errors.New("target is required")
		}
		if r.request.MaxDepth <= 0 {
			return nil, errors.New("max_depth must be > 0")
		}
		b, err := json.Marshal(r.request)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(b))
	}

	req, err := api.NewRequest("POST", basePath(cfg)+"/traversers/paths", nil, body)
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

	data := PathsPostResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &PathsPostResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
