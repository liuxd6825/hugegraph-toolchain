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

func newResourceAllocationFunc(t api.Transport) ResourceAllocation {
	return func(o ...func(*ResourceAllocationRequest)) (*ResourceAllocationResponse, error) {
		var r = ResourceAllocationRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type ResourceAllocation func(o ...func(*ResourceAllocationRequest)) (*ResourceAllocationResponse, error)

type ResourceAllocationRequest struct {
	ctx       context.Context
	vertex    string
	other     string
	direction string
	label     string
	maxDegree int64
	limit     int64
}

type ResourceAllocationResponse struct {
	StatusCode int                            `json:"-"`
	Header     http.Header                    `json:"-"`
	Body       io.ReadCloser                  `json:"-"`
	Data       ResourceAllocationResponseData `json:"-"`
}

type ResourceAllocationResponseData struct {
	ResourceAllocation float64 `json:"resource_allocation"`
}

func (r ResourceAllocationRequest) Do(ctx context.Context, transport api.Transport) (*ResourceAllocationResponse, error) {
	if len(r.vertex) == 0 {
		return nil, errors.New("resource_allocation: vertex is required")
	}
	if len(r.other) == 0 {
		return nil, errors.New("resource_allocation: other is required")
	}

	params := &url.Values{}
	params.Add("vertex", quoteVertexID(r.vertex))
	params.Add("other", quoteVertexID(r.other))
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

	url := getURL(transport, "resourceallocation")
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

	data := ResourceAllocationResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &ResourceAllocationResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (a ResourceAllocation) WithContext(ctx context.Context) func(*ResourceAllocationRequest) {
	return func(r *ResourceAllocationRequest) { r.ctx = ctx }
}
func (a ResourceAllocation) WithVertex(v string) func(*ResourceAllocationRequest) {
	return func(r *ResourceAllocationRequest) { r.vertex = v }
}
func (a ResourceAllocation) WithOther(v string) func(*ResourceAllocationRequest) {
	return func(r *ResourceAllocationRequest) { r.other = v }
}
func (a ResourceAllocation) WithDirection(v string) func(*ResourceAllocationRequest) {
	return func(r *ResourceAllocationRequest) { r.direction = v }
}
func (a ResourceAllocation) WithLabel(v string) func(*ResourceAllocationRequest) {
	return func(r *ResourceAllocationRequest) { r.label = v }
}
func (a ResourceAllocation) WithMaxDegree(v int64) func(*ResourceAllocationRequest) {
	return func(r *ResourceAllocationRequest) { r.maxDegree = v }
}
func (a ResourceAllocation) WithLimit(v int64) func(*ResourceAllocationRequest) {
	return func(r *ResourceAllocationRequest) { r.limit = v }
}
