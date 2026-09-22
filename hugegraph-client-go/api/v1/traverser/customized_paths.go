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

	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/api"
	structtraverser "github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/traverser"
)

// CustomizedPaths is the fluent method binding for the
// "POST /traversers/customizedpaths" endpoint.
//
// 3.2.15 Customized Paths
type CustomizedPaths func(o ...func(*CustomizedPathsRequest)) (*CustomizedPathsResponse, error)

func newCustomizedPathsFunc(t api.Transport) CustomizedPaths {
	return func(o ...func(*CustomizedPathsRequest)) (*CustomizedPathsResponse, error) {
		var r = CustomizedPathsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type CustomizedPathsRequest struct {
	ctx     context.Context
	request *structtraverser.CustomizedPathsRequest
	body    io.Reader
}

type CustomizedPathsResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       CustomizedPathsResponseData
}

type CustomizedPathsResponseData struct {
	Paths    []CustomizedPathEntry    `json:"paths"`
	Vertices []map[string]interface{} `json:"vertices,omitempty"`
	Edges    []map[string]interface{} `json:"edges,omitempty"`
}

// CustomizedPathEntry mirrors {"objects":[...], "weights":[...]} response item.
type CustomizedPathEntry struct {
	Objects []interface{} `json:"objects"`
	Weights []float64     `json:"weights,omitempty"`
}

func (c CustomizedPaths) WithRequestBody(b io.Reader) func(*CustomizedPathsRequest) {
	return func(r *CustomizedPathsRequest) { r.body = b }
}

func (c CustomizedPaths) WithRequest(req *structtraverser.CustomizedPathsRequest) func(*CustomizedPathsRequest) {
	return func(r *CustomizedPathsRequest) { r.request = req }
}

func (c CustomizedPaths) WithContext(ctx context.Context) func(*CustomizedPathsRequest) {
	return func(r *CustomizedPathsRequest) { r.ctx = ctx }
}

func (r CustomizedPathsRequest) Do(ctx context.Context, transport api.Transport) (*CustomizedPathsResponse, error) {
	cfg := transport.GetConfig()
	body := r.body
	if body == nil {
		if r.request == nil {
			return nil, errors.New("WithRequest or WithRequestBody must be set")
		}
		if len(r.request.Sources) == 0 {
			return nil, errors.New("sources is required")
		}
		if len(r.request.Steps) == 0 {
			return nil, errors.New("steps is required")
		}
		b, err := json.Marshal(r.request)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(b))
	}

	req, err := api.NewRequest("POST", basePath(cfg)+"/traversers/customizedpaths", nil, body)
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

	data := CustomizedPathsResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &CustomizedPathsResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
