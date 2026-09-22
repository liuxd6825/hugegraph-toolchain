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

// KoutPair exposes the basic GET and advanced POST variants of the K-out API.
//
// 3.2.1 K-out API (GET, basic)
// 3.2.2 K-out API (POST, advanced)
type KoutPair struct {
	Get  KoutGet
	Post KoutPost
}

func newKoutPair(t api.Transport) KoutPair {
	return KoutPair{
		Get:  newKoutGetFunc(t),
		Post: newKoutPostFunc(t),
	}
}

// ----- GET ----------------------------------------------------------------

// KoutGet is the fluent method binding for the
// "GET /traversers/kout" endpoint.
type KoutGet func(o ...func(*KoutGetRequest)) (*KoutGetResponse, error)

func newKoutGetFunc(t api.Transport) KoutGet {
	return func(o ...func(*KoutGetRequest)) (*KoutGetResponse, error) {
		var r = KoutGetRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type KoutGetRequest struct {
	ctx       context.Context
	source    string
	direction string
	maxDepth  int
	label     string
	nearest   bool
	maxDegree int
	capacity  int
	limit     int
}

type KoutGetResponse struct {
	StatusCode int                 `json:"-"`
	Header     http.Header         `json:"-"`
	Body       io.ReadCloser       `json:"-"`
	Data       KoutGetResponseData `json:"-"`
}

// KoutGetResponseData mirrors the documented response shape:
//
//	{ "vertices": ["..."] }
type KoutGetResponseData struct {
	Vertices []string `json:"vertices"`
}

func (k KoutGet) WithSource(v string) func(*KoutGetRequest) {
	return func(r *KoutGetRequest) { r.source = v }
}
func (k KoutGet) WithDirection(v string) func(*KoutGetRequest) {
	return func(r *KoutGetRequest) { r.direction = v }
}
func (k KoutGet) WithMaxDepth(v int) func(*KoutGetRequest) {
	return func(r *KoutGetRequest) { r.maxDepth = v }
}
func (k KoutGet) WithLabel(v string) func(*KoutGetRequest) {
	return func(r *KoutGetRequest) { r.label = v }
}
func (k KoutGet) WithNearest(v bool) func(*KoutGetRequest) {
	return func(r *KoutGetRequest) { r.nearest = v }
}
func (k KoutGet) WithDegree(v int) func(*KoutGetRequest) {
	return func(r *KoutGetRequest) { r.maxDegree = v }
}
func (k KoutGet) WithCapacity(v int) func(*KoutGetRequest) {
	return func(r *KoutGetRequest) { r.capacity = v }
}
func (k KoutGet) WithLimit(v int) func(*KoutGetRequest) {
	return func(r *KoutGetRequest) { r.limit = v }
}
func (k KoutGet) WithContext(ctx context.Context) func(*KoutGetRequest) {
	return func(r *KoutGetRequest) { r.ctx = ctx }
}

func (r KoutGetRequest) Do(ctx context.Context, transport api.Transport) (*KoutGetResponse, error) {
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
	if r.nearest {
		params.Set("nearest", "true")
	}
	if r.maxDegree > 0 {
		params.Set("max_degree", strconv.Itoa(r.maxDegree))
	}
	if r.capacity > 0 {
		params.Set("capacity", strconv.Itoa(r.capacity))
	}
	if r.limit > 0 {
		params.Set("limit", strconv.Itoa(r.limit))
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/kout", params, nil)
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

	data := KoutGetResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &KoutGetResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}

// ----- POST ---------------------------------------------------------------

// KoutPost is the fluent method binding for the
// "POST /traversers/kout" endpoint.
type KoutPost func(o ...func(*KoutPostRequest)) (*KoutPostResponse, error)

func newKoutPostFunc(t api.Transport) KoutPost {
	return func(o ...func(*KoutPostRequest)) (*KoutPostResponse, error) {
		var r = KoutPostRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type KoutPostRequest struct {
	ctx     context.Context
	body    io.Reader
	source  interface{}
	builder *structtraverser.KoutRequestBuilder
}

type KoutPostResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       structtraverser.Kout
}

// WithRequestBody sets the raw JSON body for advanced callers who want full
// control over the payload.
func (k KoutPost) WithRequestBody(b io.Reader) func(*KoutPostRequest) {
	return func(r *KoutPostRequest) { r.body = b }
}

// WithRequestBuilder takes a *traverser.KoutRequestBuilder produced by
// structtraverser.NewKoutRequestBuilder(). The builder is responsible for
// validation, so Do only marshals its result.
func (k KoutPost) WithRequestBuilder(b *structtraverser.KoutRequestBuilder) func(*KoutPostRequest) {
	return func(r *KoutPostRequest) { r.builder = b }
}

// WithSource is a convenience setter that wraps source via graph.FormatVertexID.
func (k KoutPost) WithSource(v string) func(*KoutPostRequest) {
	return func(r *KoutPostRequest) { r.source = v }
}

func (k KoutPost) WithContext(ctx context.Context) func(*KoutPostRequest) {
	return func(r *KoutPostRequest) { r.ctx = ctx }
}

func (r KoutPostRequest) Do(ctx context.Context, transport api.Transport) (*KoutPostResponse, error) {
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

	req, err := api.NewRequest("POST", basePath(cfg)+"/traversers/kout", nil, body)
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

	data := structtraverser.Kout{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &KoutPostResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
