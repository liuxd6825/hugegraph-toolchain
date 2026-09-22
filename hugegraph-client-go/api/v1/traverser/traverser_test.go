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

package traverser_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go"
	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/constant"
	structtraverser "github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/traverser"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newMockServerClient builds a CommonClient whose Transport rewrites every
// outbound request to point at the supplied httptest.Server.
func newMockServerClient(t *testing.T, srv *httptest.Server) *hugegraph.CommonClient {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	port, _ := strconv.Atoi(u.Port())
	c, err := hugegraph.NewCommonClient(hugegraph.Config{
		Host:  u.Hostname(),
		Port:  port,
		Graph: "hugegraph",
	})
	if err != nil {
		t.Fatalf("NewCommonClient: %v", err)
	}
	return c
}

// readBody drains an HTTP response body into bytes.
func readBody(t *testing.T, body io.ReadCloser) []byte {
	t.Helper()
	all, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return all
}

// ---------------------------------------------------------------------------
// Wiring
// ---------------------------------------------------------------------------

// TestTraverserAllFieldsNonNil verifies that every method on a freshly-built
// Traverser is non-nil. This catches regressions where a new field is added
// to traverser.Traverser but its newXxxFunc helper is not wired up.
func TestTraverserAllFieldsNonNil(t *testing.T) {
	c, err := hugegraph.NewCommonClient(hugegraph.Config{Host: "127.0.0.1", Port: 8080, Graph: "hugegraph"})
	if err != nil {
		t.Fatalf("NewCommonClient: %v", err)
	}
	tr := c.Traverser
	checks := []struct {
		name string
		ok   bool
	}{
		{"ShortestPath", tr.ShortestPath != nil},
		{"AllShortestPaths", tr.AllShortestPaths != nil},
		{"MultiNodeShortestPath", tr.MultiNodeShortestPath != nil},
		{"SingleSourceShortestPath", tr.SingleSourceShortestPath != nil},
		{"WeightedShortestPath", tr.WeightedShortestPath != nil},
		{"Kneighbor.Get", tr.Kneighbor.Get != nil},
		{"Kneighbor.Post", tr.Kneighbor.Post != nil},
		{"Kout.Get", tr.Kout.Get != nil},
		{"Kout.Post", tr.Kout.Post != nil},
		{"Crosspoints", tr.Crosspoints != nil},
		{"CustomizedCrosspoints", tr.CustomizedCrosspoints != nil},
		{"Paths.Get", tr.Paths.Get != nil},
		{"Paths.Post", tr.Paths.Post != nil},
		{"CustomizedPaths", tr.CustomizedPaths != nil},
		{"TemplatePaths", tr.TemplatePaths != nil},
		{"Rings", tr.Rings != nil},
		{"Rays", tr.Rays != nil},
		{"JaccardSimilarity.Get", tr.JaccardSimilarity.Get != nil},
		{"JaccardSimilarity.Post", tr.JaccardSimilarity.Post != nil},
		{"FusiformSimilarity", tr.FusiformSimilarity != nil},
		{"SameNeighbors", tr.SameNeighbors != nil},
		{"Edges.List", tr.Edges.List != nil},
		{"Edges.Shards", tr.Edges.Shards != nil},
		{"Edges.Scan", tr.Edges.Scan != nil},
		{"Vertices.List", tr.Vertices.List != nil},
		{"Vertices.Shards", tr.Vertices.Shards != nil},
		{"Vertices.Scan", tr.Vertices.Scan != nil},
	}
	for _, c := range checks {
		if !c.ok {
			t.Errorf("%s is nil", c.name)
		}
	}
}

// ---------------------------------------------------------------------------
// GET endpoints
// ---------------------------------------------------------------------------

func TestShortestPathGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/graphs/hugegraph/traversers/shortestpath" {
			t.Errorf("path = %s", r.URL.Path)
		}
		for _, want := range []string{
			"source=%22marko%22",
			"target=%22peter%22",
			"direction=out",
			"max_depth=3",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"path":["\"marko\"","\"josh\"","\"peter\""]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.ShortestPath(
		c.Traverser.ShortestPath.WithSource("marko"),
		c.Traverser.ShortestPath.WithTarget("peter"),
		c.Traverser.ShortestPath.WithDirection(constant.OUT.String()),
		c.Traverser.ShortestPath.WithMaxDepth(3),
	)
	if err != nil {
		t.Fatalf("ShortestPath: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d", resp.StatusCode)
	}
	if got := len(resp.Data.Path); got != 3 {
		t.Errorf("len(Path) = %d, want 3", got)
	}
}

func TestAllShortestPathsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/allshortestpaths" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"a\"","\"z\""]}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.AllShortestPaths(
		c.Traverser.AllShortestPaths.WithSource("a"),
		c.Traverser.AllShortestPaths.WithTarget("z"),
		c.Traverser.AllShortestPaths.WithMaxDepth(5),
	)
	if err != nil {
		t.Fatalf("AllShortestPaths: %v", err)
	}
	if len(resp.Data.Paths) != 1 || len(resp.Data.Paths[0].Objects) != 2 {
		t.Errorf("paths = %+v", resp.Data.Paths)
	}
}

func TestSingleSourceShortestPathGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "source=%22marko%22") {
			t.Errorf("query missing source: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"paths":{"weights":{"1:josh":{"weight":1.0,"vertices":["\"marko\"","\"josh\""]}}}}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.SingleSourceShortestPath(
		c.Traverser.SingleSourceShortestPath.WithSource("marko"),
		c.Traverser.SingleSourceShortestPath.WithLimit(10),
	)
	if err != nil {
		t.Fatalf("SingleSourceShortestPath: %v", err)
	}
	if len(resp.Paths.Weights) != 1 {
		t.Errorf("paths = %+v", resp.Paths)
	}
}

func TestWeightedShortestPathGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "weight=time") {
			t.Errorf("missing weight: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"path":{"weight":2.0,"vertices":["\"a\"","\"b\""]}}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.WeightedShortestPath(
		c.Traverser.WeightedShortestPath.WithSource("a"),
		c.Traverser.WeightedShortestPath.WithTarget("b"),
		c.Traverser.WeightedShortestPath.WithWeight("time"),
	)
	if err != nil {
		t.Fatalf("WeightedShortestPath: %v", err)
	}
	if resp.Data.Path.Weight != 2.0 {
		t.Errorf("weight = %v", resp.Data.Path.Weight)
	}
}

func TestKoutGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "max_depth=2") {
			t.Errorf("missing max_depth: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"vertices":["\"vadas\"","\"lop\""]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.Kout.Get(
		c.Traverser.Kout.Get.WithSource("marko"),
		c.Traverser.Kout.Get.WithMaxDepth(2),
	)
	if err != nil {
		t.Fatalf("Kout.Get: %v", err)
	}
	if len(resp.Data.Vertices) != 2 {
		t.Errorf("vertices = %+v", resp.Data.Vertices)
	}
}

func TestKneighborGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "max_depth=2") {
			t.Errorf("missing max_depth: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"vertices":["\"vadas\"","\"josh\"","\"lop\""]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.Kneighbor.Get(
		c.Traverser.Kneighbor.Get.WithSource("marko"),
		c.Traverser.Kneighbor.Get.WithMaxDepth(2),
	)
	if err != nil {
		t.Fatalf("Kneighbor.Get: %v", err)
	}
	if len(resp.Data.Vertices) != 3 {
		t.Errorf("vertices = %+v", resp.Data.Vertices)
	}
}

func TestCrosspointsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"crosspoints":[{"crosspoint":"1:josh","objects":["2:lop","1:josh","2:ripple"]}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.Crosspoints(
		c.Traverser.Crosspoints.WithSource("2:lop"),
		c.Traverser.Crosspoints.WithTarget("2:ripple"),
		c.Traverser.Crosspoints.WithMaxDepth(3),
	)
	if err != nil {
		t.Fatalf("Crosspoints: %v", err)
	}
	if len(resp.Data.Crosspoints) != 1 {
		t.Errorf("crosspoints = %+v", resp.Data.Crosspoints)
	}
}

func TestPathsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"a\"","\"b\""]}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.Paths.Get(
		c.Traverser.Paths.Get.WithSource("a"),
		c.Traverser.Paths.Get.WithTarget("b"),
		c.Traverser.Paths.Get.WithMaxDepth(3),
	)
	if err != nil {
		t.Fatalf("Paths.Get: %v", err)
	}
	if len(resp.Data.Paths) != 1 {
		t.Errorf("paths = %+v", resp.Data.Paths)
	}
}

func TestRingsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"rings":[{"objects":["\"a\"","\"b\"","\"a\""]}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.Rings(
		c.Traverser.Rings.WithSource("a"),
		c.Traverser.Rings.WithMaxDepth(3),
	)
	if err != nil {
		t.Fatalf("Rings: %v", err)
	}
	if len(resp.Data.Rings) != 1 {
		t.Errorf("rings = %+v", resp.Data.Rings)
	}
}

func TestRaysGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"rays":[{"objects":["\"a\"","\"b\""]}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.Rays(
		c.Traverser.Rays.WithSource("a"),
		c.Traverser.Rays.WithMaxDepth(3),
	)
	if err != nil {
		t.Fatalf("Rays: %v", err)
	}
	if len(resp.Data.Rays) != 1 {
		t.Errorf("rays = %+v", resp.Data.Rays)
	}
}

func TestJaccardSimilarityGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jaccard_similarity":0.25}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.JaccardSimilarity.Get(
		c.Traverser.JaccardSimilarity.Get.WithVertex("marko"),
		c.Traverser.JaccardSimilarity.Get.WithOther("josh"),
	)
	if err != nil {
		t.Fatalf("JaccardSimilarity.Get: %v", err)
	}
	if resp.Data.Similarity != 0.25 {
		t.Errorf("similarity = %v, want 0.25", resp.Data.Similarity)
	}
}

func TestSameNeighborsGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"same_neighbors":["2:lop"]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.SameNeighbors(
		c.Traverser.SameNeighbors.WithVertex("marko"),
		c.Traverser.SameNeighbors.WithOther("josh"),
	)
	if err != nil {
		t.Fatalf("SameNeighbors: %v", err)
	}
	if len(resp.Data.GetSameNeighbors()) != 1 {
		t.Errorf("same_neighbors = %+v", resp.Data)
	}
}

func TestVerticesList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphs/hugegraph/traversers/vertices" {
			t.Errorf("path = %s", r.URL.Path)
		}
		for _, want := range []string{"ids=%221%3Amarko%22", "ids=%221%3Ajosh%22"} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		_, _ = w.Write([]byte(`{"vertices":[{"id":"1:marko","label":"person","type":"vertex","properties":{"name":"marko"}}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
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

func TestVerticesShards(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "split_size=100") {
			t.Errorf("missing split_size: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"shards":[{"start":"0","end":"100","length":3}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
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

func TestVerticesScan(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "start=0") || !strings.Contains(r.URL.RawQuery, "end=100") {
			t.Errorf("missing start/end: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"vertices":[]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	_, err := c.Traverser.Vertices.Scan(
		c.Traverser.Vertices.Scan.WithStart("0"),
		c.Traverser.Vertices.Scan.WithEnd("100"),
	)
	if err != nil {
		t.Fatalf("Vertices.Scan: %v", err)
	}
}

func TestEdgesList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"edges":[]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.Edges.List(
		c.Traverser.Edges.List.WithIDs([]string{"S1:marko>1>S1:vadas"}),
	)
	if err != nil {
		t.Fatalf("Edges.List: %v", err)
	}
	_ = resp
}

func TestEdgesShards(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"shards":[{"start":"0","end":"100","length":5}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	_, err := c.Traverser.Edges.Shards(
		c.Traverser.Edges.Shards.WithSplitSize(100),
	)
	if err != nil {
		t.Fatalf("Edges.Shards: %v", err)
	}
}

func TestEdgesScan(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"edges":[]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	_, err := c.Traverser.Edges.Scan(
		c.Traverser.Edges.Scan.WithStart("0"),
		c.Traverser.Edges.Scan.WithEnd("100"),
	)
	if err != nil {
		t.Fatalf("Edges.Scan: %v", err)
	}
}

// ---------------------------------------------------------------------------
// POST endpoints
// ---------------------------------------------------------------------------

func TestKoutPost(t *testing.T) {
	var captured []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = readBody(t, r.Body)
		_, _ = w.Write([]byte(`{"size":1,"ids":["\"vadas\""]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	b := structtraverser.NewKoutRequestBuilder().Source("marko").MaxDepth(2).Capacity(100).Limit(10)
	if _, err := b.Steps().Direction(constant.OUT).Build(); err != nil {
		t.Fatalf("Steps Build: %v", err)
	}
	resp, err := c.Traverser.Kout.Post(
		c.Traverser.Kout.Post.WithRequestBuilder(b),
	)
	if err != nil {
		t.Fatalf("Kout.Post: %v", err)
	}
	if resp.Data.GetSize() != 1 {
		t.Errorf("size = %d", resp.Data.GetSize())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(captured, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if src, _ := body["source"].(string); src != `"marko"` {
		t.Errorf("source = %q", src)
	}
}

func TestKneighborPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"size":2,"ids":["\"josh\"","\"lop\""]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	b := structtraverser.NewKneighborRequestBuilder().Source("marko").MaxDepth(3).Limit(100)
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
}

func TestPathsPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"a\"","\"b\""]}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	req := &structtraverser.PathsRequest{
		Source:    "a",
		Target:    "b",
		Direction: constant.OUT.String(),
		MaxDepth:  3,
	}
	resp, err := c.Traverser.Paths.Post(
		c.Traverser.Paths.Post.WithRequest(req),
	)
	if err != nil {
		t.Fatalf("Paths.Post: %v", err)
	}
	if len(resp.Data.Paths) != 1 {
		t.Errorf("paths = %+v", resp.Data.Paths)
	}
}

func TestCustomizedPaths(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"a\"","\"b\""]}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	req := &structtraverser.CustomizedPathsRequest{
		Sources:  []interface{}{"a"},
		Steps:    []structtraverser.StepPattern{{Direction: constant.OUT.String()}},
		Capacity: 1000,
		Limit:    10,
	}
	resp, err := c.Traverser.CustomizedPaths(
		c.Traverser.CustomizedPaths.WithRequest(req),
	)
	if err != nil {
		t.Fatalf("CustomizedPaths: %v", err)
	}
	if len(resp.Data.Paths) != 1 {
		t.Errorf("paths = %+v", resp.Data.Paths)
	}
}

func TestTemplatePaths(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"a\"","\"b\""]}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	req := &structtraverser.TemplatePathsRequest{
		Source: "a",
		Target: "b",
		Steps:  []structtraverser.StepPattern{{Direction: constant.OUT.String()}},
	}
	resp, err := c.Traverser.TemplatePaths(
		c.Traverser.TemplatePaths.WithRequest(req),
	)
	if err != nil {
		t.Fatalf("TemplatePaths: %v", err)
	}
	if len(resp.Data.Paths) != 1 {
		t.Errorf("paths = %+v", resp.Data.Paths)
	}
}

func TestCustomizedCrosspoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"crosspoints":["1:josh"],"paths":[]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.CustomizedCrosspoints(
		c.Traverser.CustomizedCrosspoints.WithSources([]interface{}{"marko", "josh"}),
		c.Traverser.CustomizedCrosspoints.WithPathPatterns([][]structtraverser.CrosspointStep{
			{{Direction: constant.OUT.String()}},
		}),
	)
	if err != nil {
		t.Fatalf("CustomizedCrosspoints: %v", err)
	}
	if len(resp.Data.GetCrosspoints()) != 1 {
		t.Errorf("crosspoints = %+v", resp.Data.GetCrosspoints())
	}
}

func TestMultiNodeShortestPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"paths":[{"objects":["\"a\"","\"b\""]}]}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
	resp, err := c.Traverser.MultiNodeShortestPath(
		c.Traverser.MultiNodeShortestPath.WithSourceIDs([]interface{}{"a", "b"}),
		c.Traverser.MultiNodeShortestPath.WithTargetIDs([]interface{}{"c"}),
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
}

func TestJaccardSimilarityPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"1:peter":0.4,"1:josh":0.2}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
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

func TestFusiformSimilarity(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"similars":{"3:p1":[{"id":"3:p2","score":0.5}]}}`))
	}))
	defer srv.Close()

	c := newMockServerClient(t, srv)
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

// ---------------------------------------------------------------------------
// Graphspace routing
// ---------------------------------------------------------------------------

func TestGraphSpaceRouting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Caller is configured with GraphSpace=DEFAULT, so we expect
		// /graphspaces/DEFAULT/graphs/{graph}/traversers/...
		if !strings.HasPrefix(r.URL.Path, "/graphspaces/DEFAULT/graphs/hugegraph/traversers/") {
			t.Errorf("path = %s, want /graphspaces/DEFAULT/... prefix", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"path":["a","b"]}`))
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL)
	port, _ := strconv.Atoi(u.Port())
	c, err := hugegraph.NewCommonClient(hugegraph.Config{
		Host: u.Hostname(), Port: port,
		Graph: "hugegraph", GraphSpace: "DEFAULT",
	})
	if err != nil {
		t.Fatalf("NewCommonClient: %v", err)
	}
	_, err = c.Traverser.ShortestPath(
		c.Traverser.ShortestPath.WithSource("a"),
		c.Traverser.ShortestPath.WithTarget("b"),
		c.Traverser.ShortestPath.WithMaxDepth(2),
	)
	if err != nil {
		t.Fatalf("ShortestPath with GraphSpace: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Real-server smoke test (skipped by default)
// ---------------------------------------------------------------------------

// TestTraverserRealServer performs a minimal smoke test against the real
// HugeGraph server. The test is skipped unless the REAL_HUGEGRAPH env var
// is set to a non-empty value, since CI usually lacks a live server.
//
// Set HG_REAL_HOST / HG_REAL_PORT / HG_REAL_USER / HG_REAL_PASSWORD to
// override defaults.
//
// Run with:
//
//	REAL_HUGEGRAPH=1 HG_REAL_HOST=192.168.120.200 HG_REAL_PORT=18080 \
//	  HG_REAL_USER=admin HG_REAL_PASSWORD=admin \
//	  go test ./api/v1/traverser/... -run TestTraverserRealServer -v
func TestTraverserRealServer(t *testing.T) {
	if os.Getenv("REAL_HUGEGRAPH") == "" {
		t.Skip("set REAL_HUGEGRAPH=1 to enable real-server smoke test")
	}
	host := envOr("HG_REAL_HOST", "192.168.120.200")
	port := envOrInt("HG_REAL_PORT", 18080)
	user := envOr("HG_REAL_USER", "admin")
	pass := envOr("HG_REAL_PASSWORD", "admin")
	graph := envOr("HG_REAL_GRAPH", "hugegraph")

	c, err := hugegraph.NewCommonClient(hugegraph.Config{
		Host: host, Port: port, Graph: graph,
		Username: user, Password: pass,
	})
	if err != nil {
		t.Fatalf("NewCommonClient: %v", err)
	}

	resp, err := c.Traverser.ShortestPath(
		c.Traverser.ShortestPath.WithSource("1:marko"),
		c.Traverser.ShortestPath.WithTarget("2:ripple"),
		c.Traverser.ShortestPath.WithMaxDepth(3),
	)
	if err != nil {
		t.Fatalf("ShortestPath: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func envOr(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

func envOrInt(name string, def int) int {
	v := os.Getenv(name)
	if v == "" {
		return def
	}
	var n int
	for _, c := range v {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}
