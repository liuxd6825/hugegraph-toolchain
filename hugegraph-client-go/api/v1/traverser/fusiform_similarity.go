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

// FusiformSimilarity is the fluent method binding for the
// "POST /traversers/fusiformsimilarity" endpoint.
//
// 3.2.21 Fusiform Similarity
type FusiformSimilarity func(o ...func(*FusiformSimilarityRequest)) (*FusiformSimilarityResponse, error)

func newFusiformSimilarityFunc(t api.Transport) FusiformSimilarity {
	return func(o ...func(*FusiformSimilarityRequest)) (*FusiformSimilarityResponse, error) {
		var r = FusiformSimilarityRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type FusiformSimilarityRequest struct {
	ctx     context.Context
	body    io.Reader
	request *structtraverser.FusiformSimilarityRequest
}

type FusiformSimilarityResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       FusiformSimilarityResponseData
}

// FusiformSimilarityResponseData mirrors the documented 3.2.21 response:
//
//	{
//	  "similars": {
//	    "<sourceId>": [ { "id": "<similarId>", "score": <float>, "intermediaries": [...] } ]
//	  },
//	  "vertices": [...]
//	}
//
// We use map[string] keys instead of map[interface{}] (which the internal
// model uses but Go's encoding/json rejects for non-comparable interface keys).
type FusiformSimilarityResponseData struct {
	Similars map[string][]FusiformSimilarItem `json:"similars"`
	Vertices []map[string]interface{}         `json:"vertices,omitempty"`
	Measure  map[string]float64               `json:"measure,omitempty"`
}

// FusiformSimilarItem is one scored match in the similars map.
type FusiformSimilarItem struct {
	ID             interface{}   `json:"id"`
	Score          float64       `json:"score"`
	Intermediaries []interface{} `json:"intermediaries,omitempty"`
}

// GetSimilars returns the per-source similars map.
func (f FusiformSimilarityResponseData) GetSimilars() map[string][]FusiformSimilarItem {
	return f.Similars
}

func (f FusiformSimilarity) WithRequestBody(b io.Reader) func(*FusiformSimilarityRequest) {
	return func(r *FusiformSimilarityRequest) { r.body = b }
}

func (f FusiformSimilarity) WithRequest(req *structtraverser.FusiformSimilarityRequest) func(*FusiformSimilarityRequest) {
	return func(r *FusiformSimilarityRequest) { r.request = req }
}

func (f FusiformSimilarity) WithContext(ctx context.Context) func(*FusiformSimilarityRequest) {
	return func(r *FusiformSimilarityRequest) { r.ctx = ctx }
}

func (r FusiformSimilarityRequest) Do(ctx context.Context, transport api.Transport) (*FusiformSimilarityResponse, error) {
	cfg := transport.GetConfig()
	body := r.body
	if body == nil {
		if r.request == nil {
			return nil, errors.New("WithRequest or WithRequestBody must be set")
		}
		if len(r.request.Sources) == 0 {
			return nil, errors.New("sources is required")
		}
		b, err := json.Marshal(r.request)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(b))
	}

	req, err := api.NewRequest("POST", basePath(cfg)+"/traversers/fusiformsimilarity", nil, body)
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

	data := FusiformSimilarityResponseData{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &FusiformSimilarityResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
