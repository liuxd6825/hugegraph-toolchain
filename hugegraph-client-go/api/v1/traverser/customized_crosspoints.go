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

// CustomizedCrosspoints is the fluent method binding for the
// "POST /traversers/customizedcrosspoints" endpoint.
//
// 3.2.18 Customized Crosspoints
type CustomizedCrosspoints func(o ...func(*CustomizedCrosspointsRequest)) (*CustomizedCrosspointsResponse, error)

func newCustomizedCrosspointsFunc(t api.Transport) CustomizedCrosspoints {
	return func(o ...func(*CustomizedCrosspointsRequest)) (*CustomizedCrosspointsResponse, error) {
		var r = CustomizedCrosspointsRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

type CustomizedCrosspointsRequest struct {
	ctx          context.Context
	body         io.Reader
	sources      []interface{}
	labels       []string
	properties   map[string]interface{}
	pathPatterns [][]structtraverser.CrosspointStep
	capacity     int
	limit        int
	withPath     bool
	withVertex   bool
	withEdge     bool
}

type CustomizedCrosspointsResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       structtraverser.CustomizedCrosspoints
}

func (c CustomizedCrosspoints) WithRequestBody(b io.Reader) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.body = b }
}

func (c CustomizedCrosspoints) WithSources(ids []interface{}) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.sources = ids }
}

func (c CustomizedCrosspoints) WithPathPatterns(patterns [][]structtraverser.CrosspointStep) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.pathPatterns = patterns }
}

func (c CustomizedCrosspoints) WithLabels(labels []string) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.labels = labels }
}

func (c CustomizedCrosspoints) WithProperties(props map[string]interface{}) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.properties = props }
}

func (c CustomizedCrosspoints) WithCapacity(v int) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.capacity = v }
}

func (c CustomizedCrosspoints) WithLimit(v int) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.limit = v }
}

func (c CustomizedCrosspoints) WithPath(v bool) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.withPath = v }
}

func (c CustomizedCrosspoints) WithVertex(v bool) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.withVertex = v }
}

func (c CustomizedCrosspoints) WithEdge(v bool) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.withEdge = v }
}

func (c CustomizedCrosspoints) WithContext(ctx context.Context) func(*CustomizedCrosspointsRequest) {
	return func(r *CustomizedCrosspointsRequest) { r.ctx = ctx }
}

func (r CustomizedCrosspointsRequest) Do(ctx context.Context, transport api.Transport) (*CustomizedCrosspointsResponse, error) {
	cfg := transport.GetConfig()
	body := r.body
	if body == nil {
		if len(r.sources) == 0 {
			return nil, errors.New("sources is required")
		}
		if len(r.pathPatterns) == 0 {
			return nil, errors.New("path_patterns is required")
		}
		// Build JSON manually to keep the wire shape consistent with the
		// documented structure: each pattern is a list of steps, each step
		// carries direction/labels/properties.
		patterns := make([]struct {
			Steps []structtraverser.CrosspointStep `json:"steps"`
		}, len(r.pathPatterns))
		for i, steps := range r.pathPatterns {
			for _, s := range steps {
				if len(r.labels) > 0 && len(s.Labels) == 0 {
					s.Labels = r.labels
				}
				if r.properties != nil && len(s.Properties) == 0 {
					s.Properties = r.properties
				}
				patterns[i].Steps = append(patterns[i].Steps, s)
			}
		}
		payload := map[string]interface{}{
			"sources":       r.sources,
			"path_patterns": patterns,
			"capacity":      r.capacity,
			"limit":         r.limit,
			"with_path":     r.withPath,
			"with_vertex":   r.withVertex,
			"with_edge":     r.withEdge,
		}
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(b))
	}

	req, err := api.NewRequest("POST", basePath(cfg)+"/traversers/customizedcrosspoints", nil, body)
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

	data := structtraverser.CustomizedCrosspoints{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	resp := &CustomizedCrosspointsResponse{
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
		Data:       data,
	}
	return resp, nil
}
