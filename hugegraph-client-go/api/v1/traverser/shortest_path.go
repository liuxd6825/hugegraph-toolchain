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

// ShortestPath is the fluent method binding for the
// "GET /traversers/shortestpath" endpoint. See full documentation at
// https://hugegraph.apache.org/docs/clients/restful-api/traverser/
//
// 3.2.8 Shortest Path
type ShortestPath func(o ...func(*ShortestPathRequest)) (*ShortestPathResponse, error)

func newShortestPathFunc(t api.Transport) ShortestPath {
	return func(o ...func(*ShortestPathRequest)) (*ShortestPathResponse, error) {
		var r = ShortestPathRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type ShortestPathRequest struct {
	ctx        context.Context
	source     string
	target     string
	direction  string
	maxDepth   int
	label      string
	maxDegree  int
	skipDegree int
	capacity   int
}

type ShortestPathResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       structtraverser.PathOfVertices
}

func (s ShortestPath) WithSource(v string) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.source = v }
}
func (s ShortestPath) WithTarget(v string) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.target = v }
}
func (s ShortestPath) WithDirection(v string) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.direction = v }
}
func (s ShortestPath) WithMaxDepth(v int) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.maxDepth = v }
}
func (s ShortestPath) WithLabel(v string) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.label = v }
}
func (s ShortestPath) WithDegree(v int) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.maxDegree = v }
}
func (s ShortestPath) WithSkipDegree(v int) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.skipDegree = v }
}
func (s ShortestPath) WithCapacity(v int) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.capacity = v }
}
func (s ShortestPath) WithContext(ctx context.Context) func(*ShortestPathRequest) {
	return func(r *ShortestPathRequest) { r.ctx = ctx }
}

func (r ShortestPathRequest) Do(ctx context.Context, transport api.Transport) (*ShortestPathResponse, error) {
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
	params.Set("max_depth", strconv.Itoa(r.maxDepth))
	if r.label != "" {
		params.Set("label", r.label)
	}
	if r.maxDegree > 0 {
		params.Set("max_degree", strconv.Itoa(r.maxDegree))
	}
	if r.skipDegree > 0 {
		params.Set("skip_degree", strconv.Itoa(r.skipDegree))
	}
	if r.capacity > 0 {
		params.Set("capacity", strconv.Itoa(r.capacity))
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/shortestpath", params, nil)
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

	data := structtraverser.PathOfVertices{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &ShortestPathResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
