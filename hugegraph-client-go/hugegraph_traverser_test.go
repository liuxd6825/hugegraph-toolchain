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

package hugegraph

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/constant"
	structtraverser "github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/traverser"
)

// This file provides end-to-end unit tests for every traverser API method
// implemented by api/v1/traverser. It complements the wiring/end-to-end tests
// in hugegraph_test.go and the per-API tests in api/v1/traverser/traverser_test.go
// by covering all 25 methods from the root package, exercising:
//
//   - HTTP method and URL path
//   - Query parameter handling (including graph.FormatVertexID quoting)
//   - POST JSON body serialisation
//   - Response unmarshal into the documented wire types
//
// All tests run against an httptest mock server that mimics HugeGraph and
// share the helper functions defined in hugegraph_test.go (mustNewClientWithServer,
// urlParse, portFromURL, redirectTransport, readAll, defaultConfig) because
// this file lives in the same package.

// -----------------------------------------------------------------------------
// Shortest Path (POST)
//
// Note: TestCommonClientTraverserShortestPathEndToEnd in hugegraph_test.go
// already exercises ShortestPath.GET. This test provides an independent
// minimal sanity check on the same method from the root package.
// -----------------------------------------------------------------------------

func TestTraverserShortestPathBasic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/graphs/hugegraph/traversers/shortestpath" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "source=%22marko%22") {
			t.Errorf("query missing source: %s", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "target=%22peter%22") {
			t.Errorf("query missing target: %s", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "max_depth=3") {
			t.Errorf("query missing max_depth: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"path":["\"marko\"","\"josh\"","\"peter\""]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.ShortestPath(
		c.Traverser.ShortestPath.WithSource("marko"),
		c.Traverser.ShortestPath.WithTarget("peter"),
		c.Traverser.ShortestPath.WithMaxDepth(3),
	)
	if err != nil {
		t.Fatalf("ShortestPath: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		body := readAll(t, resp.Body)
		t.Errorf("status = %d, body=%s", resp.StatusCode, string(body))
	}
	if len(resp.Data.Path) != 3 {
		t.Errorf("len(Path) = %d, want 3", len(resp.Data.Path))
	}
}

// -----------------------------------------------------------------------------
// All Shortest Paths (GET)
// -----------------------------------------------------------------------------

func TestTraverserAllShortestPaths(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/allshortestpaths" {
			t.Errorf("path = %s", r.URL.Path)
		}
		for _, want := range []string{
			"source=%22marko%22",
			"target=%22peter%22",
			"max_depth=5",
			"direction=out",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"a\"","\"b\""]},{"objects":["\"a\"","\"c\""]}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.AllShortestPaths(
		c.Traverser.AllShortestPaths.WithSource("marko"),
		c.Traverser.AllShortestPaths.WithTarget("peter"),
		c.Traverser.AllShortestPaths.WithDirection(constant.OUT.String()),
		c.Traverser.AllShortestPaths.WithMaxDepth(5),
	)
	if err != nil {
		t.Fatalf("AllShortestPaths: %v", err)
	}
	if len(resp.Data.Paths) != 2 {
		t.Errorf("paths = %+v", resp.Data.Paths)
	}
}

// -----------------------------------------------------------------------------
// Multi Node Shortest Path (POST)
// -----------------------------------------------------------------------------

func TestTraverserMultiNodeShortestPath(t *testing.T) {
	var captured []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = readAll(t, r.Body)
		if r.URL.Path != "/graphs/hugegraph/traversers/multinodeshortestpath" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"a\"","\"c\""]}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.MultiNodeShortestPath(
		c.Traverser.MultiNodeShortestPath.WithSourceIDs([]interface{}{"marko", "josh"}),
		c.Traverser.MultiNodeShortestPath.WithTargetIDs([]interface{}{"lop"}),
		c.Traverser.MultiNodeShortestPath.WithStep(&structtraverser.StepPattern{
			Direction: constant.OUT.String(),
		}),
		c.Traverser.MultiNodeShortestPath.WithMaxDepth(3),
	)
	if err != nil {
		t.Fatalf("MultiNodeShortestPath: %v", err)
	}
	if len(resp.Data.GetPaths()) != 1 {
		t.Errorf("paths = %+v", resp.Data)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(captured, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	vertsNode, ok := body["vertices"].(map[string]interface{})
	if !ok {
		t.Errorf("vertices missing in body: %s", captured)
	}
	if _, ok := vertsNode["sources"]; !ok {
		t.Errorf("vertices.sources missing")
	}
	if _, ok := vertsNode["targets"]; !ok {
		t.Errorf("vertices.targets missing")
	}
	if body["max_depth"].(float64) != 3 {
		t.Errorf("max_depth = %v, want 3", body["max_depth"])
	}
}

// -----------------------------------------------------------------------------
// Single Source Shortest Path (GET)
// -----------------------------------------------------------------------------

func TestTraverserSingleSourceShortest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/singlesourceshortestpath" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "source=%22marko%22") {
			t.Errorf("query missing source: %s", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "limit=5") {
			t.Errorf("query missing limit: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"paths":{"weights":{"1:josh":{"weight":1.0,"vertices":["\"marko\"","\"josh\""]}}}}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.SingleSourceShortestPath(
		c.Traverser.SingleSourceShortestPath.WithSource("marko"),
		c.Traverser.SingleSourceShortestPath.WithLimit(5),
	)
	if err != nil {
		t.Fatalf("SingleSourceShortestPath: %v", err)
	}
	if len(resp.Paths.Weights) != 1 {
		t.Errorf("paths = %+v", resp.Paths)
	}
}

// -----------------------------------------------------------------------------
// Weighted Shortest Path (GET)
// -----------------------------------------------------------------------------

func TestTraverserWeightedShortestPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/weightedshortestpath" {
			t.Errorf("path = %s", r.URL.Path)
		}
		for _, want := range []string{
			"source=%22marko%22",
			"target=%22ripple%22",
			"weight=time",
			"direction=out",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"path":{"weight":2.5,"vertices":["\"marko\"","\"josh\"","\"ripple\""]}}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.WeightedShortestPath(
		c.Traverser.WeightedShortestPath.WithSource("marko"),
		c.Traverser.WeightedShortestPath.WithTarget("ripple"),
		c.Traverser.WeightedShortestPath.WithWeight("time"),
		c.Traverser.WeightedShortestPath.WithDirection(constant.OUT.String()),
	)
	if err != nil {
		t.Fatalf("WeightedShortestPath: %v", err)
	}
	if resp.Data.Path.Weight != 2.5 {
		t.Errorf("weight = %v, want 2.5", resp.Data.Path.Weight)
	}
}

// -----------------------------------------------------------------------------
// K-out (GET, basic)
// -----------------------------------------------------------------------------

func TestTraverserKoutGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/kout" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		for _, want := range []string{
			"source=%22marko%22",
			"max_depth=2",
			"direction=out",
			"label=knows",
			"nearest=true",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"vertices":["1:vadas","2:lop"]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Kout.Get(
		c.Traverser.Kout.Get.WithSource("marko"),
		c.Traverser.Kout.Get.WithMaxDepth(2),
		c.Traverser.Kout.Get.WithDirection(constant.OUT.String()),
		c.Traverser.Kout.Get.WithLabel("knows"),
		c.Traverser.Kout.Get.WithNearest(true),
	)
	if err != nil {
		t.Fatalf("Kout.Get: %v", err)
	}
	if len(resp.Data.Vertices) != 2 {
		t.Errorf("vertices = %+v", resp.Data.Vertices)
	}
}

// -----------------------------------------------------------------------------
// K-out (POST, advanced) - basic sanity check; full body assertion lives in
// hugegraph_test.go (TestCommonClientTraverserKoutPostEndToEnd).
// -----------------------------------------------------------------------------

func TestTraverserKoutPostBasic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/graphs/hugegraph/traversers/kout" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"size":2,"ids":["1:vadas","2:lop"]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	b := structtraverser.NewKoutRequestBuilder().
		Source("marko").
		MaxDepth(2).
		Capacity(100).
		Limit(10)
	if _, err := b.Steps().Direction(constant.OUT).Build(); err != nil {
		t.Fatalf("Steps Build: %v", err)
	}
	resp, err := c.Traverser.Kout.Post(
		c.Traverser.Kout.Post.WithRequestBuilder(b),
	)
	if err != nil {
		t.Fatalf("Kout.Post: %v", err)
	}
	if resp.Data.GetSize() != 2 {
		t.Errorf("size = %d", resp.Data.GetSize())
	}
}

// -----------------------------------------------------------------------------
// K-neighbor (GET, basic)
// -----------------------------------------------------------------------------

func TestTraverserKneighborGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/kneighbor" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		for _, want := range []string{
			"source=%22marko%22",
			"max_depth=3",
			"direction=both",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"vertices":["1:josh","2:lop","2:ripple"]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Kneighbor.Get(
		c.Traverser.Kneighbor.Get.WithSource("marko"),
		c.Traverser.Kneighbor.Get.WithMaxDepth(3),
		c.Traverser.Kneighbor.Get.WithDirection(constant.BOTH.String()),
	)
	if err != nil {
		t.Fatalf("Kneighbor.Get: %v", err)
	}
	if len(resp.Data.Vertices) != 3 {
		t.Errorf("vertices = %+v", resp.Data.Vertices)
	}
}

// -----------------------------------------------------------------------------
// K-neighbor (POST, advanced)
// -----------------------------------------------------------------------------

func TestTraverserKneighborPost(t *testing.T) {
	var captured []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = readAll(t, r.Body)
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/graphs/hugegraph/traversers/kneighbor" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"size":2,"ids":["1:josh","2:lop"]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	b := structtraverser.NewKneighborRequestBuilder().
		Source("marko").
		MaxDepth(3).
		Limit(100).
		WithVertex(true)
	if _, err := b.Steps().Direction(constant.BOTH).Build(); err != nil {
		t.Fatalf("Steps Build: %v", err)
	}
	resp, err := c.Traverser.Kneighbor.Post(
		c.Traverser.Kneighbor.Post.WithRequestBuilder(b),
	)
	if err != nil {
		t.Fatalf("Kneighbor.Post: %v", err)
	}
	if resp.Data.GetSize() != 2 {
		t.Errorf("size = %d", resp.Data.GetSize())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(captured, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if src, _ := body["source"].(string); src != `"marko"` {
		t.Errorf("source = %q, want %q", src, `"marko"`)
	}
}

// -----------------------------------------------------------------------------
// Crosspoints (GET)
// -----------------------------------------------------------------------------

func TestTraverserCrosspoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/crosspoints" {
			t.Errorf("path = %s", r.URL.Path)
		}
		for _, want := range []string{
			"source=%22marko%22",
			"target=%22ripple%22",
			"max_depth=4",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"crosspoints":[{"crosspoint":"1:josh","objects":["2:lop","1:josh","2:ripple"]}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Crosspoints(
		c.Traverser.Crosspoints.WithSource("marko"),
		c.Traverser.Crosspoints.WithTarget("ripple"),
		c.Traverser.Crosspoints.WithMaxDepth(4),
	)
	if err != nil {
		t.Fatalf("Crosspoints: %v", err)
	}
	if len(resp.Data.Crosspoints) != 1 {
		t.Errorf("crosspoints = %+v", resp.Data.Crosspoints)
	}
	if resp.Data.Crosspoints[0].Crosspoint != "1:josh" {
		t.Errorf("crosspoint = %s", resp.Data.Crosspoints[0].Crosspoint)
	}
}

// -----------------------------------------------------------------------------
// Customized Crosspoints (POST)
// -----------------------------------------------------------------------------

func TestTraverserCustomizedCrosspoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/graphs/hugegraph/traversers/customizedcrosspoints" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"crosspoints":["1:josh"],"paths":[{"objects":["\"marko\"","\"josh\""]}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.CustomizedCrosspoints(
		c.Traverser.CustomizedCrosspoints.WithSources([]interface{}{"marko", "josh"}),
		c.Traverser.CustomizedCrosspoints.WithPathPatterns([][]structtraverser.CrosspointStep{
			{{Direction: constant.OUT.String(), Labels: []string{"knows"}}},
		}),
		c.Traverser.CustomizedCrosspoints.WithPath(true),
	)
	if err != nil {
		t.Fatalf("CustomizedCrosspoints: %v", err)
	}
	if len(resp.Data.GetCrosspoints()) != 1 {
		t.Errorf("crosspoints = %+v", resp.Data.GetCrosspoints())
	}
}

// -----------------------------------------------------------------------------
// Paths (GET, basic)
// -----------------------------------------------------------------------------

func TestTraverserPathsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/paths" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		for _, want := range []string{
			"source=%22marko%22",
			"target=%22peter%22",
			"max_depth=3",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"marko\"","\"josh\""]},{"objects":["\"marko\"","\"lop\"","\"peter\""]}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Paths.Get(
		c.Traverser.Paths.Get.WithSource("marko"),
		c.Traverser.Paths.Get.WithTarget("peter"),
		c.Traverser.Paths.Get.WithMaxDepth(3),
	)
	if err != nil {
		t.Fatalf("Paths.Get: %v", err)
	}
	if len(resp.Data.Paths) != 2 {
		t.Errorf("paths = %+v", resp.Data.Paths)
	}
}

// -----------------------------------------------------------------------------
// Paths (POST, advanced)
// -----------------------------------------------------------------------------

func TestTraverserPathsPost(t *testing.T) {
	var captured []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = readAll(t, r.Body)
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/graphs/hugegraph/traversers/paths" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"a\"","\"b\""]}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Paths.Post(
		c.Traverser.Paths.Post.WithRequest(&structtraverser.PathsRequest{
			Source:    "marko",
			Target:    "peter",
			Direction: constant.OUT.String(),
			MaxDepth:  3,
			Capacity:  100,
		}),
	)
	if err != nil {
		t.Fatalf("Paths.Post: %v", err)
	}
	if len(resp.Data.Paths) != 1 {
		t.Errorf("paths = %+v", resp.Data.Paths)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(captured, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if src, _ := body["source"].(string); src != "marko" {
		t.Errorf("source = %q", src)
	}
}

// -----------------------------------------------------------------------------
// Customized Paths (POST)
// -----------------------------------------------------------------------------

func TestTraverserCustomizedPaths(t *testing.T) {
	var captured []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = readAll(t, r.Body)
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/graphs/hugegraph/traversers/customizedpaths" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"marko\"","\"josh\"","\"lop\""]}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.CustomizedPaths(
		c.Traverser.CustomizedPaths.WithRequest(&structtraverser.CustomizedPathsRequest{
			Sources: []interface{}{"marko"},
			Steps: []structtraverser.StepPattern{
				{Direction: constant.OUT.String(), Labels: []string{"knows"}},
			},
			Capacity: 1000,
			Limit:    10,
		}),
	)
	if err != nil {
		t.Fatalf("CustomizedPaths: %v", err)
	}
	if len(resp.Data.Paths) != 1 {
		t.Errorf("paths = %+v", resp.Data.Paths)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(captured, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if steps, ok := body["steps"].([]interface{}); !ok || len(steps) != 1 {
		t.Errorf("steps = %+v", body["steps"])
	}
}

// -----------------------------------------------------------------------------
// Template Paths (POST)
// -----------------------------------------------------------------------------

func TestTraverserTemplatePaths(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/graphs/hugegraph/traversers/templatepaths" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"vadas\"","\"marko\"","\"lop\""]}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.TemplatePaths(
		c.Traverser.TemplatePaths.WithRequest(&structtraverser.TemplatePathsRequest{
			Source: "vadas",
			Target: "lop",
			Steps: []structtraverser.StepPattern{
				{Direction: constant.OUT.String()},
			},
		}),
	)
	if err != nil {
		t.Fatalf("TemplatePaths: %v", err)
	}
	if len(resp.Data.Paths) != 1 {
		t.Errorf("paths = %+v", resp.Data.Paths)
	}
}

// -----------------------------------------------------------------------------
// Rings (GET)
// -----------------------------------------------------------------------------

func TestTraverserRings(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/rings" {
			t.Errorf("path = %s", r.URL.Path)
		}
		for _, want := range []string{
			"source=%22marko%22",
			"max_depth=3",
			"source_in_ring=false",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"rings":[{"objects":["\"a\"","\"b\"","\"a\""]}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Rings(
		c.Traverser.Rings.WithSource("marko"),
		c.Traverser.Rings.WithMaxDepth(3),
		c.Traverser.Rings.WithSourceInRing(false),
	)
	if err != nil {
		t.Fatalf("Rings: %v", err)
	}
	if len(resp.Data.Rings) != 1 {
		t.Errorf("rings = %+v", resp.Data.Rings)
	}
}

// -----------------------------------------------------------------------------
// Rays (GET)
// -----------------------------------------------------------------------------

func TestTraverserRays(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/rays" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "source=%22marko%22") {
			t.Errorf("query missing source: %s", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "max_depth=5") {
			t.Errorf("query missing max_depth: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"rays":[{"objects":["\"marko\"","\"josh\"","\"ripple\""]}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Rays(
		c.Traverser.Rays.WithSource("marko"),
		c.Traverser.Rays.WithMaxDepth(5),
	)
	if err != nil {
		t.Fatalf("Rays: %v", err)
	}
	if len(resp.Data.Rays) != 1 {
		t.Errorf("rays = %+v", resp.Data.Rays)
	}
}

// -----------------------------------------------------------------------------
// Jaccard Similarity (GET, two-vertex)
// -----------------------------------------------------------------------------

func TestTraverserJaccardSimilarityGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/jaccardsimilarity" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		for _, want := range []string{
			"vertex=%22marko%22",
			"other=%22josh%22",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"jaccard_similarity":0.2}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.JaccardSimilarity.Get(
		c.Traverser.JaccardSimilarity.Get.WithVertex("marko"),
		c.Traverser.JaccardSimilarity.Get.WithOther("josh"),
	)
	if err != nil {
		t.Fatalf("JaccardSimilarity.Get: %v", err)
	}
	if resp.Data.Similarity != 0.2 {
		t.Errorf("similarity = %v, want 0.2", resp.Data.Similarity)
	}
}

// -----------------------------------------------------------------------------
// Jaccard Similarity (POST, single-source top-N)
// -----------------------------------------------------------------------------

func TestTraverserJaccardSimilarityPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/jaccardsimilarity" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		_, _ = w.Write([]byte(`{"1:josh":0.2,"2:ripple":0.33}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.JaccardSimilarity.Post(
		c.Traverser.JaccardSimilarity.Post.WithVertex("marko"),
		c.Traverser.JaccardSimilarity.Post.WithStep(&structtraverser.StepPattern{
			Direction: constant.BOTH.String(),
		}),
		c.Traverser.JaccardSimilarity.Post.WithTop(5),
	)
	if err != nil {
		t.Fatalf("JaccardSimilarity.Post: %v", err)
	}
	if len(resp.Data.Similars) != 2 {
		t.Errorf("similars = %+v", resp.Data.Similars)
	}
}

// -----------------------------------------------------------------------------
// Fusiform Similarity (POST)
// -----------------------------------------------------------------------------

func TestTraverserFusiformSimilarity(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/fusiformsimilarity" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		_, _ = w.Write([]byte(`{"similars":{"3:p1":[{"id":"3:p2","score":0.5,"intermediaries":[]}]}}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.FusiformSimilarity(
		c.Traverser.FusiformSimilarity.WithRequest(&structtraverser.FusiformSimilarityRequest{
			Sources:      []interface{}{"3:p1"},
			MinNeighbors: 1,
			Alpha:        0.5,
			Top:          5,
		}),
	)
	if err != nil {
		t.Fatalf("FusiformSimilarity: %v", err)
	}
	if len(resp.Data.GetSimilars()) != 1 {
		t.Errorf("similars = %+v", resp.Data.GetSimilars())
	}
}

// -----------------------------------------------------------------------------
// Same Neighbors (GET)
// -----------------------------------------------------------------------------

func TestTraverserSameNeighbors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/sameneighbors" {
			t.Errorf("path = %s", r.URL.Path)
		}
		for _, want := range []string{
			"vertex=%22marko%22",
			"other=%22josh%22",
			"direction=both",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"same_neighbors":["2:lop"]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.SameNeighbors(
		c.Traverser.SameNeighbors.WithVertex("marko"),
		c.Traverser.SameNeighbors.WithOther("josh"),
		c.Traverser.SameNeighbors.WithDirection(constant.BOTH.String()),
	)
	if err != nil {
		t.Fatalf("SameNeighbors: %v", err)
	}
	if len(resp.Data.GetSameNeighbors()) != 1 {
		t.Errorf("same_neighbors = %+v", resp.Data)
	}
}

// -----------------------------------------------------------------------------
// Vertices.List (GET, batch-by-id)
// -----------------------------------------------------------------------------

func TestTraverserVerticesList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/vertices" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "ids=%221%3Amarko%22") {
			t.Errorf("query missing ids=marko: %s", r.URL.RawQuery)
		}
		if !strings.Contains(r.URL.RawQuery, "ids=%221%3Ajosh%22") {
			t.Errorf("query missing ids=josh: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"vertices":[{"id":"1:marko","label":"person","type":"vertex","properties":{"name":"marko"}}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Vertices.List(
		c.Traverser.Vertices.List.WithIDs([]string{"1:marko", "1:josh"}),
	)
	if err != nil {
		t.Fatalf("Vertices.List: %v", err)
	}
	if len(resp.Data.Vertices) != 1 {
		t.Errorf("vertices = %+v", resp.Data.Vertices)
	}
}

// -----------------------------------------------------------------------------
// Vertices.Shards (GET)
// -----------------------------------------------------------------------------

func TestTraverserVerticesShards(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/vertices/shards" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "split_size=100") {
			t.Errorf("query missing split_size: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"shards":[{"start":"0","end":"100","length":3}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Vertices.Shards(
		c.Traverser.Vertices.Shards.WithSplitSize(100),
	)
	if err != nil {
		t.Fatalf("Vertices.Shards: %v", err)
	}
	if len(resp.Data.Shards) != 1 || resp.Data.Shards[0].Length != 3 {
		t.Errorf("shards = %+v", resp.Data.Shards)
	}
}

// -----------------------------------------------------------------------------
// Vertices.Scan (GET)
// -----------------------------------------------------------------------------

func TestTraverserVerticesScan(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/vertices/scan" {
			t.Errorf("path = %s", r.URL.Path)
		}
		for _, want := range []string{
			"start=0",
			"end=1000",
			"page_limit=100",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"vertices":[]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Vertices.Scan(
		c.Traverser.Vertices.Scan.WithStart("0"),
		c.Traverser.Vertices.Scan.WithEnd("1000"),
		c.Traverser.Vertices.Scan.WithPageLimit(100),
	)
	if err != nil {
		t.Fatalf("Vertices.Scan: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

// -----------------------------------------------------------------------------
// Edges.List (GET, batch-by-id)
// -----------------------------------------------------------------------------

func TestTraverserEdgesList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/edges" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "ids=") {
			t.Errorf("query missing ids: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"edges":[{"id":"S1:marko>1>20160110>S1:vadas","label":"knows","type":"edge","outV":"1:marko","outVLabel":"person","inV":"1:vadas","inVLabel":"person","properties":{"weight":0.5}}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Edges.List(
		c.Traverser.Edges.List.WithIDs([]string{"S1:marko>1>20160110>S1:vadas"}),
	)
	if err != nil {
		t.Fatalf("Edges.List: %v", err)
	}
	if len(resp.Data.Edges) != 1 {
		t.Errorf("edges = %+v", resp.Data.Edges)
	}
}

// -----------------------------------------------------------------------------
// Edges.Shards (GET)
// -----------------------------------------------------------------------------

func TestTraverserEdgesShards(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/edges/shards" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "split_size=100") {
			t.Errorf("query missing split_size: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"shards":[{"start":"0","end":"100","length":5}]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.Edges.Shards(
		c.Traverser.Edges.Shards.WithSplitSize(100),
	)
	if err != nil {
		t.Fatalf("Edges.Shards: %v", err)
	}
	if len(resp.Data.Shards) != 1 {
		t.Errorf("shards = %+v", resp.Data.Shards)
	}
}

// -----------------------------------------------------------------------------
// Edges.Scan (GET)
// -----------------------------------------------------------------------------

func TestTraverserEdgesScan(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/edges/scan" {
			t.Errorf("path = %s", r.URL.Path)
		}
		for _, want := range []string{
			"start=0",
			"end=1000",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"edges":[]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	_, err := c.Traverser.Edges.Scan(
		c.Traverser.Edges.Scan.WithStart("0"),
		c.Traverser.Edges.Scan.WithEnd("1000"),
	)
	if err != nil {
		t.Fatalf("Edges.Scan: %v", err)
	}
}

// -----------------------------------------------------------------------------
// Custom IP:port binding
//
// Demonstrates how to start the mock httptest server on a deterministic
// "host:port" instead of httptest.NewServer()'s default "127.0.0.1:<random>".
// This is useful when you want a stable URL across runs, when reproducing
// traffic from a non-loopback interface, or when pointing a client at a real
// HugeGraph instance at a known address.
//
// The test binds to "127.0.0.1:0" so the OS picks a free port while we keep
// localhost as the host. Use ":8080" to pin a port (requires it to be
// free), or "192.168.120.200:18080" to target a specific interface (note:
// binding to a non-loopback IP typically requires root or
// CAP_NET_BIND_SERVICE).
// -----------------------------------------------------------------------------

func TestTraverserCustomAddrBinding(t *testing.T) {
	const addr = "127.0.0.1:0" // OS-assigned port on the loopback interface

	srv := mustNewServerOnAddr(t, addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/shortestpath" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"path":["\"a\"","\"b\""]}`))
	}))
	defer srv.Close()

	// srv.URL contains the actual host:port chosen by the OS. Use it to
	// verify the listener is bound where we asked.
	if !strings.HasPrefix(srv.URL, "http://127.0.0.1:") {
		t.Errorf("server URL = %s, want http://127.0.0.1:<port>", srv.URL)
	}
	t.Logf("mock server bound at %s", srv.URL)

	// mustNewClientForAddr parses "host:port" and produces a CommonClient
	// pointing at it. We use srv.URL here so the test still works even if
	// the OS picked a random port.
	c := mustNewClientForAddr(t, strings.TrimPrefix(srv.URL, "http://"))
	resp, err := c.Traverser.ShortestPath(
		c.Traverser.ShortestPath.WithSource("a"),
		c.Traverser.ShortestPath.WithTarget("b"),
		c.Traverser.ShortestPath.WithMaxDepth(1),
	)
	if err != nil {
		t.Fatalf("ShortestPath against %s: %v", srv.URL, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d", resp.StatusCode)
	}
	if len(resp.Data.Path) != 2 {
		t.Errorf("len(Path) = %d, want 2", len(resp.Data.Path))
	}
}

// TestTraverserRealHugeGraphServer shows how to point a client at a real
// HugeGraph instance running on the network. The test is skipped unless the
// HG_REAL_HOST / HG_REAL_PORT / HG_REAL_USER / HG_REAL_PASSWORD env vars are
// set, so it is safe to leave enabled in CI.
//
// Example:
//
//	HG_REAL_HOST=192.168.120.200 HG_REAL_PORT=18080 \
//	  HG_REAL_USER=admin HG_REAL_PASSWORD=admin \
//	  go test -run TestTraverserRealHugeGraphServer ./...
func TestTraverserRealHugeGraphServer(t *testing.T) {
	host := os.Getenv("HG_REAL_HOST")
	if host == "" {
		t.Skip("set HG_REAL_HOST to enable real-server test")
	}
	portStr := os.Getenv("HG_REAL_PORT")
	if portStr == "" {
		portStr = "18080"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("invalid HG_REAL_PORT %q: %v", portStr, err)
	}
	user := os.Getenv("HG_REAL_USER")
	if user == "" {
		user = "admin"
	}
	pass := os.Getenv("HG_REAL_PASSWORD")
	if pass == "" {
		pass = "admin"
	}
	graph := os.Getenv("HG_REAL_GRAPH")
	if graph == "" {
		graph = "hugegraph"
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	cfg := Config{
		Host:     host,
		Port:     port,
		Graph:    graph,
		Username: user,
		Password: pass,
	}
	c := mustNewClient(t, cfg)
	t.Logf("client built for %s (graph=%s)", addr, graph)

	resp, err := c.Traverser.ShortestPath(
		c.Traverser.ShortestPath.WithSource("1:marko"),
		c.Traverser.ShortestPath.WithTarget("2:ripple"),
		c.Traverser.ShortestPath.WithMaxDepth(3),
	)
	if err != nil {
		t.Fatalf("ShortestPath against %s: %v", addr, err)
	}
	if resp.StatusCode != http.StatusOK {
		// Dump headers + body so the operator can diagnose why the real
		// server rejected the call (wrong creds, GraphSpace mismatch,
		// server requires token auth, etc.).
		headers, _ := json.Marshal(resp.Header)
		body, _ := io.ReadAll(resp.Body)
		t.Logf("real server response: status=%d headers=%s body=%s",
			resp.StatusCode, headers, string(body))
		switch resp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			t.Skipf("real HugeGraph server at %s rejected the credentials; "+
				"set HG_REAL_USER/HG_REAL_PASSWORD (or unset HG_REAL_HOST to skip)", addr)
			return
		default:
			t.Errorf("status = %d, want 200", resp.StatusCode)
			return
		}
	}
}
