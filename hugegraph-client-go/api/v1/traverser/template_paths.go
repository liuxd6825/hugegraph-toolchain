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

// TemplatePaths is the fluent method binding for the
// "POST /traversers/templatepaths" endpoint.
//
// 3.2.16 Template Paths
type TemplatePaths func(o ...func(*TemplatePathsRequest)) (*TemplatePathsResponse, error)

func newTemplatePathsFunc(t api.Transport) TemplatePaths {
	return func(o ...func(*TemplatePathsRequest)) (*TemplatePathsResponse, error) {
		var r = TemplatePathsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type TemplatePathsRequest struct {
	ctx     context.Context
	body    io.Reader
	request *structtraverser.TemplatePathsRequest
}

type TemplatePathsResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       TemplatePathsResponseData
}

type TemplatePathsResponseData struct {
	Paths    []PathsGetPathEntry      `json:"paths"`
	Vertices []map[string]interface{} `json:"vertices,omitempty"`
	Edges    []map[string]interface{} `json:"edges,omitempty"`
}

func (t TemplatePaths) WithRequestBody(b io.Reader) func(*TemplatePathsRequest) {
	return func(r *TemplatePathsRequest) { r.body = b }
}

func (t TemplatePaths) WithRequest(req *structtraverser.TemplatePathsRequest) func(*TemplatePathsRequest) {
	return func(r *TemplatePathsRequest) { r.request = req }
}

func (t TemplatePaths) WithContext(ctx context.Context) func(*TemplatePathsRequest) {
	return func(r *TemplatePathsRequest) { r.ctx = ctx }
}

func (r TemplatePathsRequest) Do(ctx context.Context, transport api.Transport) (*TemplatePathsResponse, error) {
	cfg := transport.GetConfig()
	body := r.body
	if body == nil {
		if r.request == nil {
			return nil, errors.New("WithRequest or WithRequestBody must be set")
		}
		if r.request.Source == nil {
			return nil, errors.New("source is required")
		}
		if r.request.Target == nil {
			return nil, errors.New("target is required")
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

	req, err := api.NewRequest("POST", basePath(cfg)+"/traversers/templatepaths", nil, body)
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

	data := TemplatePathsResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &TemplatePathsResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
