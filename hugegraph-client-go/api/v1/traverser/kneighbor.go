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

// KneighborPair exposes the basic GET and advanced POST variants of the
// K-neighbor API.
//
// 3.2.3 K-neighbor (GET, basic)
// 3.2.4 K-neighbor API (POST, advanced)
type KneighborPair struct {
	Get  KneighborGet
	Post KneighborPost
}

func newKneighborPair(t api.Transport) KneighborPair {
	return KneighborPair{
		Get:  newKneighborGetFunc(t),
		Post: newKneighborPostFunc(t),
	}
}

// ----- GET ----------------------------------------------------------------

// KneighborGet is the fluent method binding for the
// "GET /traversers/kneighbor" endpoint.
type KneighborGet func(o ...func(*KneighborGetRequest)) (*KneighborGetResponse, error)

func newKneighborGetFunc(t api.Transport) KneighborGet {
	return func(o ...func(*KneighborGetRequest)) (*KneighborGetResponse, error) {
		var r = KneighborGetRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type KneighborGetRequest struct {
	ctx       context.Context
	source    string
	direction string
	maxDepth  int
	label     string
	maxDegree int
	limit     int
}

type KneighborGetResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       KneighborGetResponseData
}

type KneighborGetResponseData struct {
	Vertices []string `json:"vertices"`
}

func (k KneighborGet) WithSource(v string) func(*KneighborGetRequest) {
	return func(r *KneighborGetRequest) { r.source = v }
}
func (k KneighborGet) WithDirection(v string) func(*KneighborGetRequest) {
	return func(r *KneighborGetRequest) { r.direction = v }
}
func (k KneighborGet) WithMaxDepth(v int) func(*KneighborGetRequest) {
	return func(r *KneighborGetRequest) { r.maxDepth = v }
}
func (k KneighborGet) WithLabel(v string) func(*KneighborGetRequest) {
	return func(r *KneighborGetRequest) { r.label = v }
}
func (k KneighborGet) WithDegree(v int) func(*KneighborGetRequest) {
	return func(r *KneighborGetRequest) { r.maxDegree = v }
}
func (k KneighborGet) WithLimit(v int) func(*KneighborGetRequest) {
	return func(r *KneighborGetRequest) { r.limit = v }
}
func (k KneighborGet) WithContext(ctx context.Context) func(*KneighborGetRequest) {
	return func(r *KneighborGetRequest) { r.ctx = ctx }
}

func (r KneighborGetRequest) Do(ctx context.Context, transport api.Transport) (*KneighborGetResponse, error) {
	if r.source == "" {
		return nil, errors.New("source is required")
	}
	if r.maxDepth <= 0 {
		return nil, errors.New("max_depth must be > 0")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	params.Set("source", graph.FormatVertexID(r.source))
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
	if r.limit > 0 {
		params.Set("limit", strconv.Itoa(r.limit))
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/kneighbor", params, nil)
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

	data := KneighborGetResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &KneighborGetResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}

// ----- POST ---------------------------------------------------------------

// KneighborPost is the fluent method binding for the
// "POST /traversers/kneighbor" endpoint.
type KneighborPost func(o ...func(*KneighborPostRequest)) (*KneighborPostResponse, error)

func newKneighborPostFunc(t api.Transport) KneighborPost {
	return func(o ...func(*KneighborPostRequest)) (*KneighborPostResponse, error) {
		var r = KneighborPostRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type KneighborPostRequest struct {
	ctx     context.Context
	body    io.Reader
	builder *structtraverser.KneighborRequestBuilder
	source  interface{}
}

type KneighborPostResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       structtraverser.Kneighbor
}

func (k KneighborPost) WithRequestBody(b io.Reader) func(*KneighborPostRequest) {
	return func(r *KneighborPostRequest) { r.body = b }
}

func (k KneighborPost) WithRequestBuilder(b *structtraverser.KneighborRequestBuilder) func(*KneighborPostRequest) {
	return func(r *KneighborPostRequest) { r.builder = b }
}

func (k KneighborPost) WithSource(v string) func(*KneighborPostRequest) {
	return func(r *KneighborPostRequest) { r.source = v }
}

func (k KneighborPost) WithContext(ctx context.Context) func(*KneighborPostRequest) {
	return func(r *KneighborPostRequest) { r.ctx = ctx }
}

func (r KneighborPostRequest) Do(ctx context.Context, transport api.Transport) (*KneighborPostResponse, error) {
	cfg := transport.GetConfig()
	body := r.body
	if body == nil {
		if r.builder == nil && r.source == nil {
			return nil, errors.New("either WithRequestBuilder or WithSource must be provided")
		}
		var payload interface{}
		if r.builder != nil {
			built, err := r.builder.Build()
			if err != nil {
				return nil, err
			}
			// Source must be wrapped with double-quotes so the server can
			// distinguish "1:marko" from the literal string "1:marko".
			built.Source = graph.FormatVertexID(built.Source)
			payload = built
		} else {
			payload = map[string]interface{}{"source": graph.FormatVertexID(r.source)}
		}
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(b))
	}

	req, err := api.NewRequest("POST", basePath(cfg)+"/traversers/kneighbor", nil, body)
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

	data := structtraverser.Kneighbor{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &KneighborPostResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
