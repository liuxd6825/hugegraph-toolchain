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
)

func newEdgeExistenceFunc(t api.Transport) EdgeExistence {
	return func(o ...func(*EdgeExistenceRequest)) (*EdgeExistenceResponse, error) {
		var r = EdgeExistenceRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type EdgeExistence func(o ...func(*EdgeExistenceRequest)) (*EdgeExistenceResponse, error)

type EdgeExistenceRequest struct {
	ctx        context.Context
	source     string
	target     string
	label      string
	sortValues string
	limit      int
}

type EdgeExistenceResponse struct {
	StatusCode int                       `json:"-"`
	Header     http.Header               `json:"-"`
	Body       io.ReadCloser             `json:"-"`
	Data       EdgeExistenceResponseData `json:"-"`
}

type EdgeExistenceResponseData struct {
	Edges []EdgeExistenceEdge `json:"edges"`
}

type EdgeExistenceEdge struct {
	ID         string                 `json:"id"`
	Label      string                 `json:"label"`
	Typ        string                 `json:"type"`
	OutV       string                 `json:"outV"`
	OutVLabel  string                 `json:"outVLabel"`
	InV        string                 `json:"inV"`
	InVLabel   string                 `json:"inVLabel"`
	Properties map[string]interface{} `json:"properties"`
}

func (r EdgeExistenceRequest) Do(ctx context.Context, transport api.Transport) (*EdgeExistenceResponse, error) {
	if len(r.source) == 0 {
		return nil, errors.New("edge_exist: source is required")
	}
	if len(r.target) == 0 {
		return nil, errors.New("edge_exist: target is required")
	}

	params := &url.Values{}
	params.Add("source", quoteVertexID(r.source))
	params.Add("target", quoteVertexID(r.target))
	if r.label != "" {
		params.Add("label", r.label)
	}
	if r.sortValues != "" {
		params.Add("sort_values", r.sortValues)
	}
	if r.limit > 0 {
		params.Add("limit", intToString(r.limit))
	}

	url := getURL(transport, "edgeexist")
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

	data := EdgeExistenceResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &EdgeExistenceResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (e EdgeExistence) WithContext(ctx context.Context) func(*EdgeExistenceRequest) {
	return func(r *EdgeExistenceRequest) { r.ctx = ctx }
}
func (e EdgeExistence) WithSource(v string) func(*EdgeExistenceRequest) {
	return func(r *EdgeExistenceRequest) { r.source = v }
}
func (e EdgeExistence) WithTarget(v string) func(*EdgeExistenceRequest) {
	return func(r *EdgeExistenceRequest) { r.target = v }
}
func (e EdgeExistence) WithLabel(v string) func(*EdgeExistenceRequest) {
	return func(r *EdgeExistenceRequest) { r.label = v }
}
func (e EdgeExistence) WithSortValues(v string) func(*EdgeExistenceRequest) {
	return func(r *EdgeExistenceRequest) { r.sortValues = v }
}
func (e EdgeExistence) WithLimit(v int) func(*EdgeExistenceRequest) {
	return func(r *EdgeExistenceRequest) { r.limit = v }
}
