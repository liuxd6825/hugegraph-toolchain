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
	"strings"

	"github.com/apache/hugegraph-toolchain/hugegraph-client-go/api"
)

func newMultiNodeShortestPathFunc(t api.Transport) MultiNodeShortestPath {
	return func(o ...func(*MultiNodeShortestPathRequest)) (*MultiNodeShortestPathResponse, error) {
		var r = MultiNodeShortestPathRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type MultiNodeShortestPath func(o ...func(*MultiNodeShortestPathRequest)) (*MultiNodeShortestPathResponse, error)

type MultiNodeShortestPathRequest struct {
	ctx     context.Context
	body    io.Reader
	reqData MultiNodeShortestPathRequestData
}

type MultiNodeShortestPathRequestData struct {
	Vertices   interface{} `json:"vertices"`
	Step       interface{} `json:"step"`
	MaxDepth   int         `json:"max_depth"`
	Capacity   int64       `json:"capacity"`
	WithVertex bool        `json:"with_vertex"`
}

type MultiNodeShortestPathResponse struct {
	StatusCode int                               `json:"-"`
	Header     http.Header                       `json:"-"`
	Body       io.ReadCloser                     `json:"-"`
	Data       MultiNodeShortestPathResponseData `json:"-"`
}

type MultiNodeShortestPathResponseData struct {
	Paths    []KoutPath               `json:"paths"`
	Vertices []map[string]interface{} `json:"vertices,omitempty"`
}

func (r MultiNodeShortestPathRequest) Do(ctx context.Context, transport api.Transport) (*MultiNodeShortestPathResponse, error) {
	if r.reqData.Vertices == nil {
		return nil, errors.New("multi_node_shortest_path: vertices is required")
	}
	if r.reqData.Step == nil {
		return nil, errors.New("multi_node_shortest_path: step is required")
	}
	if r.reqData.MaxDepth <= 0 {
		return nil, errors.New("multi_node_shortest_path: max_depth must be > 0")
	}

	if r.body == nil {
		byteBody, err := json.Marshal(&r.reqData)
		if err != nil {
			return nil, err
		}
		r.body = strings.NewReader(string(byteBody))
	}

	url := buildTraverserURL(transport, "multinodeshortestpath")
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

	data := MultiNodeShortestPathResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &MultiNodeShortestPathResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (m MultiNodeShortestPath) WithContext(ctx context.Context) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.ctx = ctx }
}
func (m MultiNodeShortestPath) WithBody(body io.Reader) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.body = body }
}
func (m MultiNodeShortestPath) WithReqData(data MultiNodeShortestPathRequestData) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.reqData = data }
}
