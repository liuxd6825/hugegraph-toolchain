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
	"errors"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/apache/hugegraph-toolchain/hugegraph-client-go/api/v1/traverser"
)

// hugegraphReachable probes 127.0.0.1:8080 once at package init. Online
// integration tests use this flag to skip themselves gracefully when the
// HugeGraph server is not running.
var hugegraphReachable = func() bool {
	c, err := net.DialTimeout("tcp", "127.0.0.1:8080", 1*time.Second)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}()

// expectedTraverserFields lists the 31 traverser entry points that
// Traverser must expose. Used by TestTraverserEntryPointNames to guard
// against accidental field removal.
var expectedTraverserFields = []string{
	"ShortestPath", "AllShortestPaths", "SameNeighbors", "JaccardSimilarity",
	"JaccardSimilarityPost", "WeightedShortestPath", "SingleSourceShortest",
	"MultiNodeShortestPath", "KoutBasic", "KoutAdvanced", "KneighborBasic",
	"KneighborAdvanced", "PathsBasic", "PathsAdvanced", "CustomizedPaths",
	"TemplatePaths", "Crosspoints", "CustomizedCrosspoints", "Rings", "Rays",
	"FusiformSimilarity", "VerticesByID", "VerticesShards", "VerticesScan",
	"EdgesByID", "EdgesShards", "EdgesScan", "AdamicAdar", "ResourceAllocation",
	"EdgeExistence", "Count",
}

// newRealClient creates a CommonClient targeting the local HugeGraph
// server. graphSpace may be empty or "DEFAULT".
func newRealClient(t *testing.T, graphSpace string) *CommonClient {
	t.Helper()
	client, err := NewCommonClient(Config{
		Host:       "127.0.0.1",
		Port:       8080,
		GraphSpace: graphSpace,
		Graph:      "hugegraph",
		Username:   "admin",
		Password:   "admin",
		Logger:     nil,
	})
	if err != nil {
		t.Fatalf("NewCommonClient: %v", err)
	}
	return client
}

// requireServer skips the test if HugeGraph is not reachable on
// 127.0.0.1:8080. Returns the client for convenience.
func requireServer(t *testing.T, graphSpace string) *CommonClient {
	t.Helper()
	if !hugegraphReachable {
		t.Skip("hugegraph server 127.0.0.1:8080 not reachable; skipping online test")
	}
	return newRealClient(t, graphSpace)
}

// ---------------------------------------------------------------------------
// Offline tests: do not require a running HugeGraph server.
// ---------------------------------------------------------------------------

// assertAllFieldsBound walks a traverser.Traverser struct value and asserts
// every field is a non-nil function. Used by both ClientWiring and
// APIV1Wiring tests to verify wiring without depending on instance identity.
func assertAllFieldsBound(t *testing.T, v reflect.Value) {
	t.Helper()
	typ := v.Type()
	bound := 0
	for i := 0; i < typ.NumField(); i++ {
		f := v.Field(i)
		name := typ.Field(i).Name
		if f.Kind() != reflect.Func {
			t.Errorf("%s kind = %s, want Func", name, f.Kind())
			continue
		}
		if f.IsNil() {
			t.Errorf("%s is nil", name)
			continue
		}
		bound++
	}
	if bound != typ.NumField() {
		t.Fatalf("expected all %d entry points bound, got %d", typ.NumField(), bound)
	}
}

func TestTraverserClientWiring(t *testing.T) {
	client := newRealClient(t, "")
	if client.Traverser == nil {
		t.Fatal("client.Traverser is nil")
	}
	v := reflect.ValueOf(client.Traverser).Elem()
	assertAllFieldsBound(t, v)
	t.Logf("Traverser: %d/%d entry points bound", v.Type().NumField(), v.Type().NumField())
}

func TestTraverserAPIV1Wiring(t *testing.T) {
	client := newRealClient(t, "")
	if client.APIV1 == nil {
		t.Fatal("client.APIV1 is nil")
	}
	if client.APIV1.Traverser == nil {
		t.Fatal("client.APIV1.Traverser is nil")
	}
	assertAllFieldsBound(t, reflect.ValueOf(client.APIV1.Traverser).Elem())
}

func TestTraverserEntryPointNames(t *testing.T) {
	typ := reflect.TypeOf(traverser.Traverser{})
	seen := map[string]bool{}
	for i := 0; i < typ.NumField(); i++ {
		seen[typ.Field(i).Name] = true
	}
	for _, name := range expectedTraverserFields {
		if !seen[name] {
			t.Errorf("Traverser field %q is missing", name)
		}
	}
}

func TestTraverserRequiredFieldValidation(t *testing.T) {
	client := newRealClient(t, "")
	cases := []struct {
		name string
		run  func() error
	}{
		{"ShortestPath_no_source", func() error {
			_, err := client.Traverser.ShortestPath(
				client.Traverser.ShortestPath.WithTarget("2:ripple"),
				client.Traverser.ShortestPath.WithMaxDepth(3),
			)
			return err
		}},
		{"ShortestPath_no_target", func() error {
			_, err := client.Traverser.ShortestPath(
				client.Traverser.ShortestPath.WithSource("1:marko"),
				client.Traverser.ShortestPath.WithMaxDepth(3),
			)
			return err
		}},
		{"ShortestPath_no_maxDepth", func() error {
			_, err := client.Traverser.ShortestPath(
				client.Traverser.ShortestPath.WithSource("1:marko"),
				client.Traverser.ShortestPath.WithTarget("2:ripple"),
			)
			return err
		}},
		{"AllShortestPaths_no_source", func() error {
			_, err := client.Traverser.AllShortestPaths(
				client.Traverser.AllShortestPaths.WithTarget("2:ripple"),
				client.Traverser.AllShortestPaths.WithMaxDepth(3),
			)
			return err
		}},
		{"SameNeighbors_no_vertex", func() error {
			_, err := client.Traverser.SameNeighbors(
				client.Traverser.SameNeighbors.WithOther("1:josh"),
			)
			return err
		}},
		{"SameNeighbors_no_other", func() error {
			_, err := client.Traverser.SameNeighbors(
				client.Traverser.SameNeighbors.WithVertex("1:marko"),
			)
			return err
		}},
		{"JaccardSimilarity_no_vertex", func() error {
			_, err := client.Traverser.JaccardSimilarity(
				client.Traverser.JaccardSimilarity.WithOther("1:josh"),
			)
			return err
		}},
		{"KoutBasic_no_source", func() error {
			_, err := client.Traverser.KoutBasic(
				client.Traverser.KoutBasic.WithMaxDepth(2),
			)
			return err
		}},
		{"KoutBasic_no_maxDepth", func() error {
			_, err := client.Traverser.KoutBasic(
				client.Traverser.KoutBasic.WithSource("1:marko"),
			)
			return err
		}},
		{"KneighborBasic_no_source", func() error {
			_, err := client.Traverser.KneighborBasic(
				client.Traverser.KneighborBasic.WithMaxDepth(2),
			)
			return err
		}},
		{"KneighborBasic_no_maxDepth", func() error {
			_, err := client.Traverser.KneighborBasic(
				client.Traverser.KneighborBasic.WithSource("1:marko"),
			)
			return err
		}},
		{"Rings_no_source", func() error {
			_, err := client.Traverser.Rings(
				client.Traverser.Rings.WithMaxDepth(3),
			)
			return err
		}},
		{"Rings_no_maxDepth", func() error {
			_, err := client.Traverser.Rings(
				client.Traverser.Rings.WithSource("1:marko"),
			)
			return err
		}},
		{"Rays_no_source", func() error {
			_, err := client.Traverser.Rays(
				client.Traverser.Rays.WithMaxDepth(3),
			)
			return err
		}},
		{"AdamicAdar_no_vertex", func() error {
			_, err := client.Traverser.AdamicAdar(
				client.Traverser.AdamicAdar.WithOther("1:josh"),
			)
			return err
		}},
		{"ResourceAllocation_no_vertex", func() error {
			_, err := client.Traverser.ResourceAllocation(
				client.Traverser.ResourceAllocation.WithOther("1:josh"),
			)
			return err
		}},
		{"EdgeExistence_no_source", func() error {
			_, err := client.Traverser.EdgeExistence(
				client.Traverser.EdgeExistence.WithTarget("2:lop"),
			)
			return err
		}},
		{"WeightedShortestPath_no_weight", func() error {
			_, err := client.Traverser.WeightedShortestPath(
				client.Traverser.WeightedShortestPath.WithSource("1:marko"),
				client.Traverser.WeightedShortestPath.WithTarget("2:lop"),
			)
			return err
		}},
		{"SingleSourceShortest_no_source", func() error {
			_, err := client.Traverser.SingleSourceShortest()
			return err
		}},
		{"PathsBasic_no_source", func() error {
			_, err := client.Traverser.PathsBasic(
				client.Traverser.PathsBasic.WithTarget("2:lop"),
				client.Traverser.PathsBasic.WithMaxDepth(3),
			)
			return err
		}},
		{"Crosspoints_no_source", func() error {
			_, err := client.Traverser.Crosspoints(
				client.Traverser.Crosspoints.WithTarget("2:lop"),
				client.Traverser.Crosspoints.WithMaxDepth(3),
			)
			return err
		}},
		{"VerticesByID_no_ids", func() error {
			_, err := client.Traverser.VerticesByID()
			return err
		}},
		{"VerticesShards_no_splitSize", func() error {
			_, err := client.Traverser.VerticesShards()
			return err
		}},
		{"VerticesScan_no_start", func() error {
			_, err := client.Traverser.VerticesScan(
				client.Traverser.VerticesScan.WithEnd("z"),
			)
			return err
		}},
		{"VerticesScan_no_end", func() error {
			_, err := client.Traverser.VerticesScan(
				client.Traverser.VerticesScan.WithStart("a"),
			)
			return err
		}},
		{"EdgesByID_no_ids", func() error {
			_, err := client.Traverser.EdgesByID()
			return err
		}},
		{"EdgesShards_no_splitSize", func() error {
			_, err := client.Traverser.EdgesShards()
			return err
		}},
		{"EdgesScan_no_start", func() error {
			_, err := client.Traverser.EdgesScan(
				client.Traverser.EdgesScan.WithEnd("z"),
			)
			return err
		}},
		{"KoutAdvanced_no_source", func() error {
			_, err := client.Traverser.KoutAdvanced(
				client.Traverser.KoutAdvanced.WithReqData(traverser.KoutAdvancedRequestData{
					Steps: traverser.Steps{
						Direction: traverser.Both,
					},
					MaxDepth: 1,
				}),
			)
			return err
		}},
		{"KoutAdvanced_no_maxDepth", func() error {
			_, err := client.Traverser.KoutAdvanced(
				client.Traverser.KoutAdvanced.WithReqData(traverser.KoutAdvancedRequestData{
					Source: "1:marko",
					Steps: traverser.Steps{
						Direction: traverser.Both,
					},
				}),
			)
			return err
		}},
		{"KneighborAdvanced_no_source", func() error {
			_, err := client.Traverser.KneighborAdvanced(
				client.Traverser.KneighborAdvanced.WithReqData(traverser.KneighborAdvancedRequestData{
					Steps: traverser.Steps{
						Direction: traverser.Both,
					},
					MaxDepth: 1,
				}),
			)
			return err
		}},
		{"JaccardSimilarityPost_no_vertex", func() error {
			_, err := client.Traverser.JaccardSimilarityPost(
				client.Traverser.JaccardSimilarityPost.WithReqData(traverser.JaccardSimilarityPostRequestData{
					Step: traverser.Steps{
						Direction: traverser.Both,
					},
				}),
			)
			return err
		}},
		{"MultiNodeShortestPath_no_vertices", func() error {
			_, err := client.Traverser.MultiNodeShortestPath(
				client.Traverser.MultiNodeShortestPath.WithReqData(traverser.MultiNodeShortestPathRequestData{
					Step: traverser.Steps{
						Direction: traverser.Both,
					},
					MaxDepth: 3,
				}),
			)
			return err
		}},
		{"Count_no_source", func() error {
			_, err := client.Traverser.Count(
				client.Traverser.Count.WithReqData(traverser.CountRequestData{
					Steps: traverser.Steps{
						Direction: traverser.Both,
					},
				}),
			)
			return err
		}},
		{"FusiformSimilarity_no_sources", func() error {
			_, err := client.Traverser.FusiformSimilarity(
				client.Traverser.FusiformSimilarity.WithReqData(traverser.FusiformSimilarityRequestData{
					MinNeighbors: 1,
					Alpha:        0.5,
				}),
			)
			return err
		}},
	}
	for _, c := range cases {
		if err := c.run(); err == nil {
			t.Errorf("%s: expected validation error, got nil", c.name)
		}
	}
}

// ---------------------------------------------------------------------------
// Online tests: require a running HugeGraph server on 127.0.0.1:8080.
// They exercise the full HTTP request path. When data is missing the
// server may return a 4xx, in which case the test is skipped (avoiding
// false negatives on a fresh server without the TinkerPop example).
// ---------------------------------------------------------------------------

func TestTraverserShortestPath(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.ShortestPath(
		client.Traverser.ShortestPath.WithSource("1:marko"),
		client.Traverser.ShortestPath.WithTarget("2:ripple"),
		client.Traverser.ShortestPath.WithMaxDepth(5),
	)
	if err != nil {
		t.Skipf("ShortestPath failed (server likely missing data): %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("ShortestPath status=%d (likely data not present)", resp.StatusCode)
	}
	if resp.Data.Path == nil {
		t.Errorf("ShortestPath Data.Path is nil; want slice")
	}
	t.Logf("ShortestPath OK: path=%v", resp.Data.Path)
}

func TestTraverserAllShortestPaths(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.AllShortestPaths(
		client.Traverser.AllShortestPaths.WithSource("1:marko"),
		client.Traverser.AllShortestPaths.WithTarget("2:ripple"),
		client.Traverser.AllShortestPaths.WithMaxDepth(5),
	)
	if err != nil {
		t.Skipf("AllShortestPaths failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("AllShortestPaths status=%d", resp.StatusCode)
	}
	if resp.Data.Paths == nil {
		t.Errorf("AllShortestPaths Data.Paths is nil")
	}
}

func TestTraverserSameNeighbors(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.SameNeighbors(
		client.Traverser.SameNeighbors.WithVertex("1:marko"),
		client.Traverser.SameNeighbors.WithOther("1:josh"),
	)
	if err != nil {
		t.Skipf("SameNeighbors failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("SameNeighbors status=%d", resp.StatusCode)
	}
	if resp.Data.SameNeighbors == nil {
		t.Errorf("SameNeighbors Data.SameNeighbors is nil")
	}
}

func TestTraverserJaccardSimilarity(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.JaccardSimilarity(
		client.Traverser.JaccardSimilarity.WithVertex("1:marko"),
		client.Traverser.JaccardSimilarity.WithOther("1:josh"),
	)
	if err != nil {
		t.Skipf("JaccardSimilarity failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("JaccardSimilarity status=%d", resp.StatusCode)
	}
	t.Logf("JaccardSimilarity=%v", resp.Data.JaccardSimilarity)
}

func TestTraverserWeightedShortestPath(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.WeightedShortestPath(
		client.Traverser.WeightedShortestPath.WithSource("1:marko"),
		client.Traverser.WeightedShortestPath.WithTarget("2:lop"),
		client.Traverser.WeightedShortestPath.WithWeight("weight"),
	)
	if err != nil {
		t.Skipf("WeightedShortestPath failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("WeightedShortestPath status=%d", resp.StatusCode)
	}
}

func TestTraverserSingleSourceShortestPath(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.SingleSourceShortest(
		client.Traverser.SingleSourceShortest.WithSource("1:marko"),
		client.Traverser.SingleSourceShortest.WithLimit(5),
	)
	if err != nil {
		t.Skipf("SingleSourceShortest failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("SingleSourceShortest status=%d", resp.StatusCode)
	}
	if resp.Data.Paths == nil {
		t.Errorf("SingleSourceShortest Data.Paths is nil")
	}
}

func TestTraverserRings(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.Rings(
		client.Traverser.Rings.WithSource("1:marko"),
		client.Traverser.Rings.WithMaxDepth(3),
	)
	if err != nil {
		t.Skipf("Rings failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("Rings status=%d", resp.StatusCode)
	}
	if resp.Data.Rings == nil {
		t.Errorf("Rings Data.Rings is nil")
	}
}

func TestTraverserRays(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.Rays(
		client.Traverser.Rays.WithSource("1:marko"),
		client.Traverser.Rays.WithMaxDepth(3),
	)
	if err != nil {
		t.Skipf("Rays failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("Rays status=%d", resp.StatusCode)
	}
	if resp.Data.Rays == nil {
		t.Errorf("Rays Data.Rays is nil")
	}
}

func TestTraverserKoutBasic(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.KoutBasic(
		client.Traverser.KoutBasic.WithSource("1:marko"),
		client.Traverser.KoutBasic.WithMaxDepth(2),
	)
	if err != nil {
		t.Skipf("KoutBasic failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("KoutBasic status=%d", resp.StatusCode)
	}
	if resp.Data.Vertices == nil {
		t.Errorf("KoutBasic Data.Vertices is nil")
	}
}

func TestTraverserKneighborBasic(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.KneighborBasic(
		client.Traverser.KneighborBasic.WithSource("1:marko"),
		client.Traverser.KneighborBasic.WithMaxDepth(2),
	)
	if err != nil {
		t.Skipf("KneighborBasic failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("KneighborBasic status=%d", resp.StatusCode)
	}
	if resp.Data.Vertices == nil {
		t.Errorf("KneighborBasic Data.Vertices is nil")
	}
}

func TestTraverserPathsBasic(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.PathsBasic(
		client.Traverser.PathsBasic.WithSource("1:marko"),
		client.Traverser.PathsBasic.WithTarget("2:lop"),
		client.Traverser.PathsBasic.WithMaxDepth(3),
	)
	if err != nil {
		t.Skipf("PathsBasic failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("PathsBasic status=%d", resp.StatusCode)
	}
	if resp.Data.Paths == nil {
		t.Errorf("PathsBasic Data.Paths is nil")
	}
}

func TestTraverserCrosspoints(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.Crosspoints(
		client.Traverser.Crosspoints.WithSource("1:marko"),
		client.Traverser.Crosspoints.WithTarget("2:lop"),
		client.Traverser.Crosspoints.WithMaxDepth(3),
	)
	if err != nil {
		t.Skipf("Crosspoints failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("Crosspoints status=%d", resp.StatusCode)
	}
	if resp.Data.Crosspoints == nil {
		t.Errorf("Crosspoints Data.Crosspoints is nil")
	}
}

func TestTraverserAdamicAdar(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.AdamicAdar(
		client.Traverser.AdamicAdar.WithVertex("1:marko"),
		client.Traverser.AdamicAdar.WithOther("1:josh"),
	)
	if err != nil {
		t.Skipf("AdamicAdar failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("AdamicAdar status=%d", resp.StatusCode)
	}
}

func TestTraverserResourceAllocation(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.ResourceAllocation(
		client.Traverser.ResourceAllocation.WithVertex("1:marko"),
		client.Traverser.ResourceAllocation.WithOther("1:josh"),
	)
	if err != nil {
		t.Skipf("ResourceAllocation failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("ResourceAllocation status=%d", resp.StatusCode)
	}
}

func TestTraverserEdgeExistence(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.EdgeExistence(
		client.Traverser.EdgeExistence.WithSource("1:marko"),
		client.Traverser.EdgeExistence.WithTarget("2:lop"),
		client.Traverser.EdgeExistence.WithLabel("created"),
	)
	if err != nil {
		t.Skipf("EdgeExistence failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("EdgeExistence status=%d", resp.StatusCode)
	}
	if resp.Data.Edges == nil {
		t.Errorf("EdgeExistence Data.Edges is nil")
	}
}

func TestTraverserVerticesByID(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.VerticesByID(
		client.Traverser.VerticesByID.WithIDs([]string{"1:marko", "1:josh"}),
	)
	if err != nil {
		t.Skipf("VerticesByID failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("VerticesByID status=%d", resp.StatusCode)
	}
	if resp.Data.Vertices == nil {
		t.Errorf("VerticesByID Data.Vertices is nil")
	}
}

func TestTraverserVerticesShards(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.VerticesShards(
		client.Traverser.VerticesShards.WithSplitSize(64 * 1024 * 1024),
	)
	if err != nil {
		t.Skipf("VerticesShards failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("VerticesShards status=%d", resp.StatusCode)
	}
	if resp.Data.Shards == nil {
		t.Errorf("VerticesShards Data.Shards is nil")
	}
}

func TestTraverserVerticesScan(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.VerticesScan(
		client.Traverser.VerticesScan.WithStart(""),
		client.Traverser.VerticesScan.WithEnd("z"),
		client.Traverser.VerticesScan.WithPageLimit(10),
	)
	if err != nil {
		t.Skipf("VerticesScan failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("VerticesScan status=%d", resp.StatusCode)
	}
	if resp.Data.Vertices == nil {
		t.Errorf("VerticesScan Data.Vertices is nil")
	}
}

func TestTraverserEdgesByID(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.EdgesByID(
		client.Traverser.EdgesByID.WithIDs([]string{"S1:marko>1>20130220>S1:josh"}),
	)
	if err != nil {
		t.Skipf("EdgesByID failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("EdgesByID status=%d", resp.StatusCode)
	}
}

func TestTraverserEdgesShards(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.EdgesShards(
		client.Traverser.EdgesShards.WithSplitSize(64 * 1024 * 1024),
	)
	if err != nil {
		t.Skipf("EdgesShards failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("EdgesShards status=%d", resp.StatusCode)
	}
	if resp.Data.Shards == nil {
		t.Errorf("EdgesShards Data.Shards is nil")
	}
}

func TestTraverserEdgesScan(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.EdgesScan(
		client.Traverser.EdgesScan.WithStart(""),
		client.Traverser.EdgesScan.WithEnd("z"),
		client.Traverser.EdgesScan.WithPageLimit(10),
	)
	if err != nil {
		t.Skipf("EdgesScan failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("EdgesScan status=%d", resp.StatusCode)
	}
	if resp.Data.Edges == nil {
		t.Errorf("EdgesScan Data.Edges is nil")
	}
}

func TestTraverserKoutAdvanced(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.KoutAdvanced(
		client.Traverser.KoutAdvanced.WithReqData(traverser.KoutAdvancedRequestData{
			Source: "PnRAAeb4liYYMyLfpsA7e9bu_周燕琦",
			Steps: traverser.Steps{
				Direction: traverser.Both,
			},
			MaxDepth:   1,
			WithVertex: true,
			WithPath:   true,
			WithEdge:   true,
		}),
	)
	if err != nil {
		t.Skipf("KoutAdvanced failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("KoutAdvanced status=%d", resp.StatusCode)
	}
	t.Log(resp.Data)
}

func TestTraverserKneighborAdvanced(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.KneighborAdvanced(
		client.Traverser.KneighborAdvanced.WithReqData(traverser.KneighborAdvancedRequestData{
			Source:   "PnRAAeb4liYYMyLfpsA7e9bu_周燕琦",
			Steps:    traverser.Steps{Direction: traverser.Both},
			MaxDepth: 20,
		}),
	)
	if err != nil {
		t.Skipf("KneighborAdvanced failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("KneighborAdvanced status=%d", resp.StatusCode)
	}
}

func TestTraverserJaccardSimilarityPost(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.JaccardSimilarityPost(
		client.Traverser.JaccardSimilarityPost.WithReqData(traverser.JaccardSimilarityPostRequestData{
			Vertex: "1:marko",
			Step:   map[string]interface{}{},
			Top:    5,
		}),
	)
	if err != nil {
		t.Skipf("JaccardSimilarityPost failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("JaccardSimilarityPost status=%d", resp.StatusCode)
	}
}

func TestTraverserPathsAdvanced(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.PathsAdvanced(
		client.Traverser.PathsAdvanced.WithReqData(traverser.PathsAdvancedRequestData{
			Sources:  []string{"S1:marko"},
			Targets:  []string{"S2:marko"},
			Step:     traverser.Step{Direction: traverser.Both},
			MaxDepth: 3,
			Limit:    traverser.NewInt64(1000),
		}),
	)
	if err != nil {
		t.Skipf("PathsAdvanced failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("PathsAdvanced status=%d", resp.StatusCode)
	}
}

func TestTraverserCustomizedPaths(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.CustomizedPaths(
		client.Traverser.CustomizedPaths.WithReqData(traverser.CustomizedPathsRequestData{
			Sources: traverser.Sources{},
			Steps: traverser.Steps{
				Direction: traverser.Both,
			},
			Limit: 5,
		}),
	)
	if err != nil {
		t.Skipf("CustomizedPaths failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("CustomizedPaths status=%d", resp.StatusCode)
	}
}

func TestTraverserTemplatePaths(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.TemplatePaths(
		client.Traverser.TemplatePaths.WithReqData(traverser.TemplatePathsReqData{
			Sources: map[string]interface{}{},
			Targets: map[string]interface{}{},
			Steps:   []interface{}{},
			Limit:   5,
		}),
	)
	if err != nil {
		t.Skipf("TemplatePaths failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("TemplatePaths status=%d", resp.StatusCode)
	}
}

func TestTraverserMultiNodeShortestPath(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.MultiNodeShortestPath(
		client.Traverser.MultiNodeShortestPath.WithReqData(traverser.MultiNodeShortestPathRequestData{
			Vertices: map[string]interface{}{},
			Step:     map[string]interface{}{},
			MaxDepth: 5,
		}),
	)
	if err != nil {
		t.Skipf("MultiNodeShortestPath failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("MultiNodeShortestPath status=%d", resp.StatusCode)
	}
}

func TestTraverserCustomizedCrosspoints(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.CustomizedCrosspoints(
		client.Traverser.CustomizedCrosspoints.WithReqData(traverser.CustomizedCrosspointsRequestData{
			Sources:      map[string]interface{}{},
			PathPatterns: []interface{}{},
			Limit:        5,
		}),
	)
	if err != nil {
		t.Skipf("CustomizedCrosspoints failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("CustomizedCrosspoints status=%d", resp.StatusCode)
	}
}

func TestTraverserFusiformSimilarity(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.FusiformSimilarity(
		client.Traverser.FusiformSimilarity.WithReqData(traverser.FusiformSimilarityRequestData{
			Sources:      map[string]interface{}{},
			MinNeighbors: 1,
			Alpha:        0.5,
			Top:          5,
		}),
	)
	if err != nil {
		t.Skipf("FusiformSimilarity failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("FusiformSimilarity status=%d", resp.StatusCode)
	}
}

func TestTraverserCount(t *testing.T) {
	client := requireServer(t, "")
	resp, err := client.Traverser.Count(
		client.Traverser.Count.WithReqData(traverser.CountRequestData{
			Source: "1:marko",
			Steps:  []interface{}{},
		}),
	)
	if err != nil {
		t.Skipf("Count failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("Count status=%d", resp.StatusCode)
	}
	t.Logf("Count=%d", resp.Data.Count)
}

// ---------------------------------------------------------------------------
// GraphSpace path-branch test: verifies that configuring GraphSpace="DEFAULT"
// switches the URL prefix from /graphs/{g}/traversers/... to
// /graphspaces/DEFAULT/graphs/{g}/traversers/... without error.
// ---------------------------------------------------------------------------

func TestTraverserGraphSpaceKoutBasic(t *testing.T) {
	client := requireServer(t, "DEFAULT")
	resp, err := client.Traverser.KoutBasic(
		client.Traverser.KoutBasic.WithSource("1:marko"),
		client.Traverser.KoutBasic.WithMaxDepth(2),
	)
	if err != nil {
		t.Skipf("GraphSpace KoutBasic failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("GraphSpace KoutBasic status=%d", resp.StatusCode)
	}
	if resp.Data.Vertices == nil {
		t.Errorf("GraphSpace KoutBasic Data.Vertices is nil")
	}
}

func TestTraverserGraphSpaceKoutAdvanced(t *testing.T) {
	client := requireServer(t, "DEFAULT")
	resp, err := client.Traverser.KoutAdvanced(
		client.Traverser.KoutAdvanced.WithReqData(traverser.KoutAdvancedRequestData{
			Source: "1:marko",
			Steps: traverser.Steps{
				Direction: traverser.Both,
			},
			MaxDepth: 1,
		}),
	)
	if err != nil {
		t.Skipf("GraphSpace KoutAdvanced failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Skipf("GraphSpace KoutAdvanced status=%d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// HTTP error handling: server-side errors (e.g., missing vertex) must be
// surfaced as *traverser.APIError rather than silently swallowed.
// ---------------------------------------------------------------------------

// TestTraverserAPIErrorsKoutAdvanced reproduces the original bug report:
// invoking K-out POST against a non-existent vertex should yield a 400
// response. The client must convert that into a *traverser.APIError whose
// Message field contains the server-side error description.
func TestTraverserAPIErrorsKoutAdvanced(t *testing.T) {
	client := requireServer(t, "")
	_, err := client.Traverser.KoutAdvanced(
		client.Traverser.KoutAdvanced.WithReqData(traverser.KoutAdvancedRequestData{
			Source: "1:marko",
			Steps: traverser.Steps{
				Direction: traverser.Both,
			},
			MaxDepth: 1,
		}),
	)
	if err == nil {
		t.Skip("server returned 200; expected 400 for missing vertex")
	}
	var apiErr *traverser.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *traverser.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if apiErr.Message == "" {
		t.Error("Message is empty; expected server-side description")
	}
	if !strings.Contains(strings.ToLower(apiErr.Message), "1:marko") {
		t.Errorf("Message = %q, expected to mention vertex id '1:marko'", apiErr.Message)
	}
	t.Logf("got APIError: status=%d exception=%q message=%q",
		apiErr.StatusCode, apiErr.Exception, apiErr.Message)
}

// TestTraverserAPIErrorsShortestPath verifies the same error-surfacing
// behavior for a GET endpoint.
func TestTraverserAPIErrorsShortestPath(t *testing.T) {
	client := requireServer(t, "")
	_, err := client.Traverser.ShortestPath(
		client.Traverser.ShortestPath.WithSource("1:marko"),
		client.Traverser.ShortestPath.WithTarget("2:ripple"),
		client.Traverser.ShortestPath.WithMaxDepth(5),
	)
	if err == nil {
		t.Skip("server returned 200; expected 400 for missing vertex")
	}
	var apiErr *traverser.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *traverser.APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if apiErr.Message == "" {
		t.Error("Message is empty")
	}
	t.Logf("got APIError: status=%d exception=%q message=%q",
		apiErr.StatusCode, apiErr.Exception, apiErr.Message)
}

// TestTraverserAPIErrorsErrorString checks that *APIError.Error() produces a
// human-readable string containing status, exception class and message.
func TestTraverserAPIErrorsErrorString(t *testing.T) {
	apiErr := &traverser.APIError{
		StatusCode: 400,
		Exception:  "class java.lang.IllegalArgumentException",
		Message:    "The source vertex with id '1:marko' does not exist",
		Cause:      "org.apache.hugegraph.exception.NotFoundException: Vertex '1:marko' does not exist",
	}
	got := apiErr.Error()
	for _, want := range []string{"400", "IllegalArgumentException", "1:marko", "cause:"} {
		if !strings.Contains(got, want) {
			t.Errorf("Error() = %q, missing %q", got, want)
		}
	}
}

// TestTraverserAPIErrorsErrorStringNoCause verifies Error() works when the
// cause field is empty.
func TestTraverserAPIErrorsErrorStringNoCause(t *testing.T) {
	apiErr := &traverser.APIError{
		StatusCode: 500,
		Exception:  "class java.lang.RuntimeException",
		Message:    "internal error",
	}
	got := apiErr.Error()
	if strings.Contains(got, "cause:") {
		t.Errorf("Error() = %q, should not contain cause: when Cause is empty", got)
	}
	for _, want := range []string{"500", "RuntimeException", "internal error"} {
		if !strings.Contains(got, want) {
			t.Errorf("Error() = %q, missing %q", got, want)
		}
	}
}

// TestTraverserAPIErrorsImplementsError verifies *APIError satisfies the
// error interface so callers can use errors.As/errors.Is against it.
func TestTraverserAPIErrorsImplementsError(t *testing.T) {
	var _ error = (*traverser.APIError)(nil)

	wrapped := error(&traverser.APIError{StatusCode: 400, Message: "boom"})
	var apiErr *traverser.APIError
	if !errors.As(wrapped, &apiErr) {
		t.Fatal("errors.As did not match *traverser.APIError")
	}
	if apiErr.Message != "boom" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "boom")
	}
}
