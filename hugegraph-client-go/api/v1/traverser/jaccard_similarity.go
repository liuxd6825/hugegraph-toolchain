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

func newJaccardSimilarityFunc(t api.Transport) JaccardSimilarity {
	return func(o ...func(*JaccardSimilarityRequest)) (*JaccardSimilarityResponse, error) {
		var r = JaccardSimilarityRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type JaccardSimilarity func(o ...func(*JaccardSimilarityRequest)) (*JaccardSimilarityResponse, error)

type JaccardSimilarityRequest struct {
	ctx       context.Context
	vertex    string
	other     string
	direction string
	label     string
	maxDegree int64
}

type JaccardSimilarityResponse struct {
	StatusCode int                           `json:"-"`
	Header     http.Header                   `json:"-"`
	Body       io.ReadCloser                 `json:"-"`
	Data       JaccardSimilarityResponseData `json:"-"`
}

type JaccardSimilarityResponseData struct {
	JaccardSimilarity float64 `json:"jaccard_similarity"`
}

func (r JaccardSimilarityRequest) Do(ctx context.Context, transport api.Transport) (*JaccardSimilarityResponse, error) {
	if len(r.vertex) == 0 {
		return nil, errors.New("jaccard_similarity: vertex is required")
	}
	if len(r.other) == 0 {
		return nil, errors.New("jaccard_similarity: other is required")
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

	url := buildTraverserURL(transport, "jaccardsimilarity")
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

	data := JaccardSimilarityResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &JaccardSimilarityResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (j JaccardSimilarity) WithContext(ctx context.Context) func(*JaccardSimilarityRequest) {
	return func(r *JaccardSimilarityRequest) { r.ctx = ctx }
}
func (j JaccardSimilarity) WithVertex(v string) func(*JaccardSimilarityRequest) {
	return func(r *JaccardSimilarityRequest) { r.vertex = v }
}
func (j JaccardSimilarity) WithOther(v string) func(*JaccardSimilarityRequest) {
	return func(r *JaccardSimilarityRequest) { r.other = v }
}
func (j JaccardSimilarity) WithDirection(v string) func(*JaccardSimilarityRequest) {
	return func(r *JaccardSimilarityRequest) { r.direction = v }
}
func (j JaccardSimilarity) WithLabel(v string) func(*JaccardSimilarityRequest) {
	return func(r *JaccardSimilarityRequest) { r.label = v }
}
func (j JaccardSimilarity) WithMaxDegree(v int64) func(*JaccardSimilarityRequest) {
	return func(r *JaccardSimilarityRequest) { r.maxDegree = v }
}

func newJaccardSimilarityPostFunc(t api.Transport) JaccardSimilarityPost {
	return func(o ...func(*JaccardSimilarityPostRequest)) (*JaccardSimilarityPostResponse, error) {
		var r = JaccardSimilarityPostRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type JaccardSimilarityPost func(o ...func(*JaccardSimilarityPostRequest)) (*JaccardSimilarityPostResponse, error)

type JaccardSimilarityPostRequest struct {
	ctx     context.Context
	body    io.Reader
	reqData JaccardSimilarityPostRequestData
}

type JaccardSimilarityPostRequestData struct {
	Vertex   interface{} `json:"vertex"`
	Step     interface{} `json:"step"`
	Top      int         `json:"top,omitempty"`
	Capacity int64       `json:"capacity,omitempty"`
}

type JaccardSimilarityPostResponse struct {
	StatusCode int                               `json:"-"`
	Header     http.Header                       `json:"-"`
	Body       io.ReadCloser                     `json:"-"`
	Data       JaccardSimilarityPostResponseData `json:"-"`
}

type JaccardSimilarityPostResponseData struct {
	SimilarsMap map[interface{}]float64 `json:"similarsMap"`
}

func (r JaccardSimilarityPostRequest) Do(ctx context.Context, transport api.Transport) (*JaccardSimilarityPostResponse, error) {
	if r.reqData.Vertex == nil {
		return nil, errors.New("jaccard_similarity_post: vertex is required")
	}
	if r.reqData.Step == nil {
		return nil, errors.New("jaccard_similarity_post: step is required")
	}

	if r.body == nil {
		byteBody, err := json.Marshal(&r.reqData)
		if err != nil {
			return nil, err
		}
		r.body = strings.NewReader(string(byteBody))
	}

	url := buildTraverserURL(transport, "jaccardsimilarity")
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

	data := JaccardSimilarityPostResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &JaccardSimilarityPostResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (j JaccardSimilarityPost) WithContext(ctx context.Context) func(*JaccardSimilarityPostRequest) {
	return func(r *JaccardSimilarityPostRequest) { r.ctx = ctx }
}
func (j JaccardSimilarityPost) WithBody(body io.Reader) func(*JaccardSimilarityPostRequest) {
	return func(r *JaccardSimilarityPostRequest) { r.body = body }
}
func (j JaccardSimilarityPost) WithReqData(data JaccardSimilarityPostRequestData) func(*JaccardSimilarityPostRequest) {
	return func(r *JaccardSimilarityPostRequest) { r.reqData = data }
}
