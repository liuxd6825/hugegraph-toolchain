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

// JaccardSimilarityPair exposes the basic GET and advanced POST (single-source)
// variants of the Jaccard Similarity API.
//
// 3.2.6 Jaccard Similarity (GET)
// 3.2.7 Jaccard Similarity (POST)
type JaccardSimilarityPair struct {
	Get  JaccardSimilarityGet
	Post JaccardSimilarityPost
}

func newJaccardSimilarityPair(t api.Transport) JaccardSimilarityPair {
	return JaccardSimilarityPair{
		Get:  newJaccardSimilarityGetFunc(t),
		Post: newJaccardSimilarityPostFunc(t),
	}
}

// ----- GET ----------------------------------------------------------------

// JaccardSimilarityGet is the fluent method binding for the
// "GET /traversers/jaccardsimilarity" endpoint.
type JaccardSimilarityGet func(o ...func(*JaccardSimilarityGetRequest)) (*JaccardSimilarityGetResponse, error)

func newJaccardSimilarityGetFunc(t api.Transport) JaccardSimilarityGet {
	return func(o ...func(*JaccardSimilarityGetRequest)) (*JaccardSimilarityGetResponse, error) {
		var r = JaccardSimilarityGetRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type JaccardSimilarityGetRequest struct {
	ctx       context.Context
	vertex    string
	other     string
	direction string
	label     string
	maxDegree int
}

type JaccardSimilarityGetResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       JaccardSimilarityGetResponseData
}

// JaccardSimilarityGetResponseData mirrors the actual 3.2.6 response shape:
//
//	{ "jaccard_similarity": 0.25 }
//
// The internal structtraverser.JaccardSimilarity models this as a map keyed by
// vertex pair, but the real HugeGraph server returns a single scalar value.
// We define a dedicated type here to capture the documented behaviour.
type JaccardSimilarityGetResponseData struct {
	Similarity float64 `json:"jaccard_similarity"`
}

func (j JaccardSimilarityGet) WithVertex(v string) func(*JaccardSimilarityGetRequest) {
	return func(r *JaccardSimilarityGetRequest) { r.vertex = v }
}
func (j JaccardSimilarityGet) WithOther(v string) func(*JaccardSimilarityGetRequest) {
	return func(r *JaccardSimilarityGetRequest) { r.other = v }
}
func (j JaccardSimilarityGet) WithDirection(v string) func(*JaccardSimilarityGetRequest) {
	return func(r *JaccardSimilarityGetRequest) { r.direction = v }
}
func (j JaccardSimilarityGet) WithLabel(v string) func(*JaccardSimilarityGetRequest) {
	return func(r *JaccardSimilarityGetRequest) { r.label = v }
}
func (j JaccardSimilarityGet) WithDegree(v int) func(*JaccardSimilarityGetRequest) {
	return func(r *JaccardSimilarityGetRequest) { r.maxDegree = v }
}
func (j JaccardSimilarityGet) WithContext(ctx context.Context) func(*JaccardSimilarityGetRequest) {
	return func(r *JaccardSimilarityGetRequest) { r.ctx = ctx }
}

func (r JaccardSimilarityGetRequest) Do(ctx context.Context, transport api.Transport) (*JaccardSimilarityGetResponse, error) {
	if r.vertex == "" {
		return nil, errors.New("vertex is required")
	}
	if r.other == "" {
		return nil, errors.New("other is required")
	}

	cfg := transport.GetConfig()
	params := &url.Values{}
	params.Set("vertex", graph.FormatVertexID(r.vertex))
	params.Set("other", graph.FormatVertexID(r.other))
	if r.direction != "" {
		params.Set("direction", r.direction)
	}
	if r.label != "" {
		params.Set("label", r.label)
	}
	if r.maxDegree > 0 {
		params.Set("max_degree", strconv.Itoa(r.maxDegree))
	}

	req, err := api.NewRequest("GET", basePath(cfg)+"/traversers/jaccardsimilarity", params, nil)
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

	data := JaccardSimilarityGetResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &JaccardSimilarityGetResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}

// ----- POST ---------------------------------------------------------------

// JaccardSimilarityPost is the fluent method binding for the
// "POST /traversers/jaccardsimilarity" endpoint.
type JaccardSimilarityPost func(o ...func(*JaccardSimilarityPostRequest)) (*JaccardSimilarityPostResponse, error)

func newJaccardSimilarityPostFunc(t api.Transport) JaccardSimilarityPost {
	return func(o ...func(*JaccardSimilarityPostRequest)) (*JaccardSimilarityPostResponse, error) {
		var r = JaccardSimilarityPostRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type JaccardSimilarityPostRequest struct {
	ctx      context.Context
	body     io.Reader
	request  *structtraverser.SingleSourceJaccardSimilarityRequest
	vertex   interface{}
	step     *structtraverser.StepPattern
	top      int
	capacity int
}

type JaccardSimilarityPostResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       JaccardSimilarityPostResponseData
}

// JaccardSimilarityPostResponseData mirrors the documented response shape
// {"<vertex-id>": <similarity>, ...}.
type JaccardSimilarityPostResponseData struct {
	Similars map[interface{}]float64 `json:"-"`
}

// rawJSONResponse is used internally to capture the dynamic-keyed map.
type rawJSONResponse struct {
	Raw map[string]float64
}

func (j JaccardSimilarityPost) WithRequestBody(b io.Reader) func(*JaccardSimilarityPostRequest) {
	return func(r *JaccardSimilarityPostRequest) { r.body = b }
}

func (j JaccardSimilarityPost) WithRequest(req *structtraverser.SingleSourceJaccardSimilarityRequest) func(*JaccardSimilarityPostRequest) {
	return func(r *JaccardSimilarityPostRequest) { r.request = req }
}

func (j JaccardSimilarityPost) WithVertex(v string) func(*JaccardSimilarityPostRequest) {
	return func(r *JaccardSimilarityPostRequest) { r.vertex = v }
}

func (j JaccardSimilarityPost) WithStep(s *structtraverser.StepPattern) func(*JaccardSimilarityPostRequest) {
	return func(r *JaccardSimilarityPostRequest) { r.step = s }
}

func (j JaccardSimilarityPost) WithTop(v int) func(*JaccardSimilarityPostRequest) {
	return func(r *JaccardSimilarityPostRequest) { r.top = v }
}

func (j JaccardSimilarityPost) WithCapacity(v int) func(*JaccardSimilarityPostRequest) {
	return func(r *JaccardSimilarityPostRequest) { r.capacity = v }
}

func (j JaccardSimilarityPost) WithContext(ctx context.Context) func(*JaccardSimilarityPostRequest) {
	return func(r *JaccardSimilarityPostRequest) { r.ctx = ctx }
}

func (r JaccardSimilarityPostRequest) Do(ctx context.Context, transport api.Transport) (*JaccardSimilarityPostResponse, error) {
	cfg := transport.GetConfig()
	body := r.body
	if body == nil {
		if r.request != nil {
			b, err := json.Marshal(r.request)
			if err != nil {
				return nil, err
			}
			body = strings.NewReader(string(b))
		} else {
			if r.vertex == nil {
				return nil, errors.New("vertex is required")
			}
			if r.step == nil {
				return nil, errors.New("step is required")
			}
			payload := map[string]interface{}{
				"vertex":   r.vertex,
				"step":     r.step,
				"top":      r.top,
				"capacity": r.capacity,
			}
			b, err := json.Marshal(payload)
			if err != nil {
				return nil, err
			}
			body = strings.NewReader(string(b))
		}
	}

	req, err := api.NewRequest("POST", basePath(cfg)+"/traversers/jaccardsimilarity", nil, body)
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
	raw := rawJSONResponse{}
	if err := json.Unmarshal(bytes, &raw.Raw); err != nil {
		return nil, err
	}
	similars := make(map[interface{}]float64, len(raw.Raw))
	for k, v := range raw.Raw {
		similars[k] = v
	}

	resp := &JaccardSimilarityPostResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       JaccardSimilarityPostResponseData{Similars: similars},
	}
	return resp, nil
}
