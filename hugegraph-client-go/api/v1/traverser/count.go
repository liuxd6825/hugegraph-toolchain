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

func newCountFunc(t api.Transport) Count {
	return func(o ...func(*CountRequest)) (*CountResponse, error) {
		var r = CountRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type Count func(o ...func(*CountRequest)) (*CountResponse, error)

type CountRequest struct {
	ctx     context.Context
	body    io.Reader
	reqData CountRequestData
}

type CountRequestData struct {
	Source            interface{} `json:"source"`
	Steps             interface{} `json:"steps"`
	ContainsTraversed bool        `json:"contains_traversed,omitempty"`
	DedupSize         int64       `json:"dedup_size,omitempty"`
}

type CountResponse struct {
	StatusCode int               `json:"-"`
	Header     http.Header       `json:"-"`
	Body       io.ReadCloser     `json:"-"`
	Data       CountResponseData `json:"-"`
}

type CountResponseData struct {
	Count int64 `json:"count"`
}

func (r CountRequest) Do(ctx context.Context, transport api.Transport) (*CountResponse, error) {
	if r.reqData.Source == nil {
		return nil, errors.New("count: source is required")
	}
	if r.reqData.Steps == nil {
		return nil, errors.New("count: steps is required")
	}

	if r.body == nil {
		byteBody, err := json.Marshal(&r.reqData)
		if err != nil {
			return nil, err
		}
		r.body = strings.NewReader(string(byteBody))
	}

	url := buildTraverserURL(transport, "count")
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

	data := CountResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &CountResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (c Count) WithContext(ctx context.Context) func(*CountRequest) {
	return func(r *CountRequest) { r.ctx = ctx }
}
func (c Count) WithBody(body io.Reader) func(*CountRequest) {
	return func(r *CountRequest) { r.body = body }
}
func (c Count) WithReqData(data CountRequestData) func(*CountRequest) {
	return func(r *CountRequest) { r.reqData = data }
}
