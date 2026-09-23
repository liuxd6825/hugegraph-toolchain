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

func newFusiformSimilarityFunc(t api.Transport) FusiformSimilarity {
	return func(o ...func(*FusiformSimilarityRequest)) (*FusiformSimilarityResponse, error) {
		var r = FusiformSimilarityRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type FusiformSimilarity func(o ...func(*FusiformSimilarityRequest)) (*FusiformSimilarityResponse, error)

type FusiformSimilarityRequest struct {
	ctx     context.Context
	body    io.Reader
	reqData FusiformSimilarityRequestData
}

type FusiformSimilarityRequestData struct {
	Sources          SourcesTargets `json:"sources"`
	Label            string         `json:"label,omitempty"`
	Direction        string         `json:"direction,omitempty"`
	MinNeighbors     int            `json:"min_neighbors"`
	Alpha            float64        `json:"alpha"`
	MinSimilars      int            `json:"min_similars"`
	Top              int            `json:"top,omitempty"`
	GroupProperty    string         `json:"group_property,omitempty"`
	MinGroups        int            `json:"min_groups,omitempty"`
	MaxDegree        int64          `json:"max_degree"`
	Capacity         int64          `json:"capacity"`
	Limit            int64          `json:"limit"`
	WithIntermediary bool           `json:"with_intermediary"`
	WithVertex       bool           `json:"with_vertex"`
}

type FusiformSimilarityResponse struct {
	StatusCode int                            `json:"-"`
	Header     http.Header                    `json:"-"`
	Body       io.ReadCloser                  `json:"-"`
	Data       FusiformSimilarityResponseData `json:"-"`
}

type FusiformSimilarityResponseData struct {
	Similars map[string][]FusiformSimilarItem `json:"similars"`
	Vertices []map[string]interface{}         `json:"vertices,omitempty"`
	Measure  map[string]float64               `json:"measure,omitempty"`
}

type FusiformSimilarItem struct {
	ID             interface{}   `json:"id"`
	Score          float64       `json:"score"`
	Intermediaries []interface{} `json:"intermediaries,omitempty"`
}

func (r FusiformSimilarityRequest) Do(ctx context.Context, transport api.Transport) (*FusiformSimilarityResponse, error) {
	if len(r.reqData.Sources.Ids) == 0 {
		return nil, errors.New("fusiform_similarity: sources is required")
	}
	if r.reqData.MinNeighbors <= 0 {
		return nil, errors.New("fusiform_similarity: min_neighbors must be > 0")
	}
	if r.reqData.Alpha <= 0 {
		return nil, errors.New("fusiform_similarity: alpha must be > 0")
	}

	if r.body == nil {
		byteBody, err := json.Marshal(&r.reqData)
		if err != nil {
			return nil, err
		}
		r.body = strings.NewReader(string(byteBody))
	}

	url := buildTraverserURL(transport, "fusiformsimilarity")
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

	data := FusiformSimilarityResponseData{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &FusiformSimilarityResponse{}
	resp.StatusCode = res.StatusCode
	resp.Header = res.Header
	resp.Body = res.Body
	resp.Data = data
	return resp, nil
}

func (f FusiformSimilarity) WithContext(ctx context.Context) func(*FusiformSimilarityRequest) {
	return func(r *FusiformSimilarityRequest) { r.ctx = ctx }
}
func (f FusiformSimilarity) WithBody(body io.Reader) func(*FusiformSimilarityRequest) {
	return func(r *FusiformSimilarityRequest) { r.body = body }
}
func (f FusiformSimilarity) WithReqData(data FusiformSimilarityRequestData) func(*FusiformSimilarityRequest) {
	return func(r *FusiformSimilarityRequest) { r.reqData = data }
}
