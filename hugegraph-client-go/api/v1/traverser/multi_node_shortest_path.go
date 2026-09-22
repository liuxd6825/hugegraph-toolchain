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
	"strconv"
	"strings"

	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/api"
	structtraverser "github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/traverser"
)

// MultiNodeShortestPath is the fluent method binding for the
// "POST /traversers/multinodeshortestpath" endpoint.
//
// 3.2.12 Multi Node Shortest Path
type MultiNodeShortestPath func(o ...func(*MultiNodeShortestPathRequest)) (*MultiNodeShortestPathResponse, error)

func newMultiNodeShortestPathFunc(t api.Transport) MultiNodeShortestPath {
	return func(o ...func(*MultiNodeShortestPathRequest)) (*MultiNodeShortestPathResponse, error) {
		var r = MultiNodeShortestPathRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type MultiNodeShortestPathRequest struct {
	ctx        context.Context
	body       io.Reader
	sourceIDs  []interface{}
	targetIDs  []interface{}
	source     MultiNodeVertexSet
	target     MultiNodeVertexSet
	step       *structtraverser.StepPattern
	maxDepth   int
	capacity   int
	withVertex bool
}

type MultiNodeVertexSet struct {
	IDs        []interface{}          `json:"ids,omitempty"`
	Label      string                 `json:"label,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type MultiNodeShortestPathResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       structtraverser.PathsWithVertices
}

// WithRequestBody allows callers to fully control the JSON payload (e.g. when
// the source/target vertex sets use label+properties predicates rather than
// direct IDs). When set, body content is used verbatim and all other setters
// are ignored at Do time.
func (m MultiNodeShortestPath) WithRequestBody(b io.Reader) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.body = b }
}

func (m MultiNodeShortestPath) WithSourceIDs(ids []interface{}) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.sourceIDs = ids }
}

func (m MultiNodeShortestPath) WithTargetIDs(ids []interface{}) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.targetIDs = ids }
}

func (m MultiNodeShortestPath) WithSources(s MultiNodeVertexSet) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.source = s }
}

func (m MultiNodeShortestPath) WithTargets(t MultiNodeVertexSet) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.target = t }
}

func (m MultiNodeShortestPath) WithStep(step *structtraverser.StepPattern) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.step = step }
}

func (m MultiNodeShortestPath) WithMaxDepth(v int) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.maxDepth = v }
}

func (m MultiNodeShortestPath) WithCapacity(v int) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.capacity = v }
}

func (m MultiNodeShortestPath) WithVertex(v bool) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.withVertex = v }
}

func (m MultiNodeShortestPath) WithContext(ctx context.Context) func(*MultiNodeShortestPathRequest) {
	return func(r *MultiNodeShortestPathRequest) { r.ctx = ctx }
}

func (r MultiNodeShortestPathRequest) Do(ctx context.Context, transport api.Transport) (*MultiNodeShortestPathResponse, error) {
	cfg := transport.GetConfig()
	body := r.body
	if body == nil {
		if r.maxDepth <= 0 {
			return nil, errors.New("max_depth must be > 0")
		}
		// Build sources/targets from either ids or label/properties.
		sources := r.source
		if len(r.sourceIDs) > 0 {
			sources = MultiNodeVertexSet{IDs: r.sourceIDs}
		}
		targets := r.target
		if len(r.targetIDs) > 0 {
			targets = MultiNodeVertexSet{IDs: r.targetIDs}
		}
		if len(sources.IDs) == 0 && sources.Label == "" {
			return nil, errors.New("sources must be set (via WithSourceIDs or WithSources)")
		}
		if len(targets.IDs) == 0 && targets.Label == "" {
			return nil, errors.New("targets must be set (via WithTargetIDs or WithTargets)")
		}
		if r.step == nil {
			return nil, errors.New("step is required")
		}
		payload := map[string]interface{}{
			"vertices": map[string]interface{}{
				"sources": sources,
				"targets": targets,
			},
			"step":        r.step,
			"max_depth":   r.maxDepth,
			"capacity":    r.capacity,
			"with_vertex": r.withVertex,
		}
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(b))
	}

	req, err := api.NewRequest("POST", basePath(cfg)+"/traversers/multinodeshortestpath", nil, body)
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

	data := structtraverser.PathsWithVertices{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &MultiNodeShortestPathResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}

// suppress unused import lint when strconv is not directly used (kept for future extension).
var _ = strconv.Itoa
