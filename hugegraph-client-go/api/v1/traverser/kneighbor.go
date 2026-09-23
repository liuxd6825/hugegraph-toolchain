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

func newKneighborBasicFunc(t api.Transport) KneighborBasic {
	return func(o ...func(*KneighborBasicRequest)) (*KneighborBasicResponse, error) {
		var r = KneighborBasicRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type KneighborBasic func(o ...func(*KneighborBasicRequest)) (*KneighborBasicResponse, error)

type KneighborBasicRequest struct {
	ctx       context.Context
	source    string
	direction string
	maxDepth  int
	label     string
	maxDegree int64
	limit     int64
}

type KneighborBasicResponse struct {
	StatusCode int                        `json:"-"`
	Header     http.Header                `json:"-"`
	Body       io.ReadCloser              `json:"-"`
	Data       KneighborBasicResponseData `json:"-"`
}

type KneighborBasicResponseData struct {
	Vertices []interface{} `json:"vertices"`
}

func (r KneighborBasicRequest) Do(ctx context.Context, transport api.Transport) (*KneighborBasicResponse, error) {
	if len(r.source) == 0 {
		return nil, errors.New("kneighbor_basic: source is required")
	}
	if r.maxDepth <= 0 {
		return nil, errors.New("kneighbor_basic: max_depth must be > 0")
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
	if r.maxDegree > 0 {
		params.Add("max_degree", int64ToString(r.maxDegree))
	}
	if r.limit > 0 {
		params.Add("limit", int64ToString(r.limit))
	}

	url := buildTraverserURL(transport, "kneighbor")
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

	data := KneighborBasicResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &KneighborBasicResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (k KneighborBasic) WithContext(ctx context.Context) func(*KneighborBasicRequest) {
	return func(r *KneighborBasicRequest) { r.ctx = ctx }
}
func (k KneighborBasic) WithSource(v string) func(*KneighborBasicRequest) {
	return func(r *KneighborBasicRequest) { r.source = v }
}
func (k KneighborBasic) WithDirection(v string) func(*KneighborBasicRequest) {
	return func(r *KneighborBasicRequest) { r.direction = v }
}
func (k KneighborBasic) WithMaxDepth(v int) func(*KneighborBasicRequest) {
	return func(r *KneighborBasicRequest) { r.maxDepth = v }
}
func (k KneighborBasic) WithLabel(v string) func(*KneighborBasicRequest) {
	return func(r *KneighborBasicRequest) { r.label = v }
}
func (k KneighborBasic) WithMaxDegree(v int64) func(*KneighborBasicRequest) {
	return func(r *KneighborBasicRequest) { r.maxDegree = v }
}
func (k KneighborBasic) WithLimit(v int64) func(*KneighborBasicRequest) {
	return func(r *KneighborBasicRequest) { r.limit = v }
}

func newKneighborAdvancedFunc(t api.Transport) KneighborAdvanced {
	return func(o ...func(*KneighborAdvancedRequest)) (*KneighborAdvancedResponse, error) {
		var r = KneighborAdvancedRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type KneighborAdvanced func(o ...func(*KneighborAdvancedRequest)) (*KneighborAdvancedResponse, error)

type KneighborAdvancedRequest struct {
	ctx     context.Context
	body    io.Reader
	reqData KneighborAdvancedRequestData
}

type KneighborAdvancedRequestData struct {
	Source     string `json:"source,omitempty"`
	Steps      Steps  `json:"steps,omitempty"`
	MaxDepth   int    `json:"max_depth,omitempty"`
	CountOnly  *bool  `json:"count_only,omitempty"`
	Limit      *int64 `json:"limit,omitempty"`
	WithVertex *bool  `json:"with_vertex,omitempty"`
	WithPath   *bool  `json:"with_path,omitempty"`
	WithEdge   *bool  `json:"with_edge,omitempty"`
}

type KneighborAdvancedResponse struct {
	StatusCode int                           `json:"-"`
	Header     http.Header                   `json:"-"`
	Body       io.ReadCloser                 `json:"-"`
	Data       KneighborAdvancedResponseData `json:"-"`
}

type KneighborAdvancedResponseData struct {
	Size      int64                    `json:"size,omitempty"`
	Kneighbor []interface{}            `json:"kneighbor,omitempty"`
	Paths     []KoutPath               `json:"paths,omitempty"`
	Vertices  []map[string]interface{} `json:"vertices,omitempty"`
	Edges     []map[string]interface{} `json:"edges,omitempty"`
}

func (r KneighborAdvancedRequest) Do(ctx context.Context, transport api.Transport) (*KneighborAdvancedResponse, error) {
	if r.reqData.Source == "" {
		return nil, errors.New("kneighbor_advanced: source is required")
	}
	if r.reqData.MaxDepth <= 0 {
		return nil, errors.New("kneighbor_advanced: max_depth must be > 0")
	}

	if r.body == nil {
		byteBody, err := json.Marshal(&r.reqData)
		if err != nil {
			return nil, err
		}
		r.body = strings.NewReader(string(byteBody))
	}

	url := buildTraverserURL(transport, "kneighbor")
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

	data := KneighborAdvancedResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &KneighborAdvancedResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (k KneighborAdvanced) WithContext(ctx context.Context) func(*KneighborAdvancedRequest) {
	return func(r *KneighborAdvancedRequest) { r.ctx = ctx }
}
func (k KneighborAdvanced) WithBody(body io.Reader) func(*KneighborAdvancedRequest) {
	return func(r *KneighborAdvancedRequest) { r.body = body }
}
func (k KneighborAdvanced) WithReqData(data KneighborAdvancedRequestData) func(*KneighborAdvancedRequest) {
	return func(r *KneighborAdvancedRequest) { r.reqData = data }
}
