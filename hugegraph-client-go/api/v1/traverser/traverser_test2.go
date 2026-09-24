/*
 * Licensed to the Apache Software Apache (ASF) under one or more
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
	"testing"
)

// TestTraverserStub is a placeholder test that documents how each traverser
// endpoint can be invoked through the Go client. The actual calls require a
// running HugeGraph server and are left commented for reference.
//
// See https://hugegraph.apache.org/cn/docs/clients/restful-api/traverser/
// for the full REST API reference.
func TestTraverserStub(t *testing.T) {

	// -------------------------------------------------------------------------
	// 3.2.1 / 3.2.2  K-out (GET basic, POST advanced)
	// -------------------------------------------------------------------------
	// client.Traverser.KoutBasic.WithSource("1:marko"), WithMaxDepth(2)
	// client.Traverser.KoutAdvanced.WithReqData(KoutAdvancedRequestData{
	//     Source:   "1:marko",
	//     Steps:    ...,
	//     MaxDepth: 1,
	// })

	// -------------------------------------------------------------------------
	// 3.2.3 / 3.2.4  K-neighbor (GET basic, POST advanced)
	// -------------------------------------------------------------------------
	// client.Traverser.KneighborBasic.WithSource("1:marko"), WithMaxDepth(2)
	// client.Traverser.KneighborAdvanced.WithReqData(KneighborAdvancedRequestData{
	//     Source:   "1:marko",
	//     Steps:    ...,
	//     MaxDepth: 3,
	// })

	// -------------------------------------------------------------------------
	// 3.2.5  Same Neighbors (GET)
	// -------------------------------------------------------------------------
	// client.Traverser.SameNeighbors.WithVertex("1:marko"), WithOther("1:josh")

	// -------------------------------------------------------------------------
	// 3.2.6 / 3.2.7  Jaccard Similarity (GET, POST)
	// -------------------------------------------------------------------------
	// client.Traverser.JaccardSimilarity.WithVertex("1:marko"), WithOther("1:josh")
	// client.Traverser.JaccardSimilarityPost.WithReqData(JaccardSimilarityPostRequestData{
	//     Vertex: "1:marko",
	//     Step:   ...,
	//     Top:    3,
	// })

	// -------------------------------------------------------------------------
	// 3.2.8 / 3.2.9  Shortest Path (GET) / All Shortest Paths (GET)
	// -------------------------------------------------------------------------
	// client.Traverser.ShortestPath.WithSource("1:marko"), WithTarget("2:ripple"), WithMaxDepth(3)
	// client.Traverser.AllShortestPaths.WithSource("1:marko"), WithTarget("2:ripple"), WithMaxDepth(3)

	// -------------------------------------------------------------------------
	// 3.2.10 / 3.2.11  Weighted Shortest Path / Single Source Shortest Path (GET)
	// -------------------------------------------------------------------------
	// client.Traverser.WeightedShortestPath.WithSource("1:marko"), WithTarget("2:lop"), WithWeight("weight")
	// client.Traverser.SingleSourceShortest.WithSource("1:marko")

	// -------------------------------------------------------------------------
	// 3.2.12  Multi Node Shortest Path (POST)
	// -------------------------------------------------------------------------
	// client.Traverser.MultiNodeShortestPath.WithReqData(MultiNodeShortestPathRequestData{
	//     Vertices: ...,
	//     Step:     ...,
	//     MaxDepth: 5,
	// })

	// -------------------------------------------------------------------------
	// 3.2.13 / 3.2.14  Paths (GET basic, POST advanced)
	// -------------------------------------------------------------------------
	// client.Traverser.PathsBasic.WithSource("1:marko"), WithTarget("2:lop"), WithMaxDepth(3)
	// client.Traverser.PathsAdvanced.WithReqData(PathsAdvancedRequestData{
	//     Sources:  ...,
	//     Targets:  ...,
	//     Step:     ...,
	//     MaxDepth: 3,
	// })

	// -------------------------------------------------------------------------
	// 3.2.15 / 3.2.16  Customized Paths / Template Paths (POST)
	// -------------------------------------------------------------------------
	// client.Traverser.CustomizedPaths.WithReqData(CustomizedPathsRequestData{
	//     Sources: ...,
	//     Steps:   ...,
	// })
	// client.Traverser.TemplatePaths.WithReqData(TemplatePathsReqData{
	//     Sources: ...,
	//     Targets: ...,
	//     Steps:   ...,
	// })

	// -------------------------------------------------------------------------
	// 3.2.17 / 3.2.18  Crosspoints / Customized Crosspoints (GET / POST)
	// -------------------------------------------------------------------------
	// client.Traverser.Crosspoints.WithSource("1:marko"), WithTarget("2:lop"), WithMaxDepth(3)
	// client.Traverser.CustomizedCrosspoints.WithReqData(CustomizedCrosspointsRequestData{
	//     Sources:      ...,
	//     PathPatterns: ...,
	// })

	// -------------------------------------------------------------------------
	// 3.2.19 / 3.2.20  Rings / Rays (GET)
	// -------------------------------------------------------------------------
	// client.Traverser.Rings.WithSource("1:marko"), WithMaxDepth(3)
	// client.Traverser.Rays.WithSource("1:marko"), WithMaxDepth(3)

	// -------------------------------------------------------------------------
	// 3.2.21  Fusiform Similarity (POST)
	// -------------------------------------------------------------------------
	// client.Traverser.FusiformSimilarity.WithReqData(FusiformSimilarityRequestData{
	//     Sources:      ...,
	//     MinNeighbors: 1,
	//     Alpha:        0.8,
	//     Top:          5,
	// })

	// -------------------------------------------------------------------------
	// 3.2.22  Vertices APIs (GET x3)
	// -------------------------------------------------------------------------
	// client.Traverser.VerticesByID.WithIDs([]string{"1:marko", "1:josh"})
	// client.Traverser.VerticesShards.WithSplitSize(1024 * 1024)
	// client.Traverser.VerticesScan.WithStart(""), WithEnd("z")

	// -------------------------------------------------------------------------
	// 3.2.23  Edges APIs (GET x3)
	// -------------------------------------------------------------------------
	// client.Traverser.EdgesByID.WithIDs([]string{"S1:marko>1>20130220>S1:josh"})
	// client.Traverser.EdgesShards.WithSplitSize(1024 * 1024)
	// client.Traverser.EdgesScan.WithStart(""), WithEnd("z")

	// -------------------------------------------------------------------------
	// 3.2.24 / 3.2.25  Adamic-Adar / Resource Allocation (GET)
	// -------------------------------------------------------------------------
	// client.Traverser.AdamicAdar.WithVertex("1:marko"), WithOther("1:josh")
	// client.Traverser.ResourceAllocation.WithVertex("1:marko"), WithOther("1:josh")

	// -------------------------------------------------------------------------
	// 3.2.26  Edge Existence (GET)
	// -------------------------------------------------------------------------
	// client.Traverser.EdgeExistence.WithSource("1:marko"), WithTarget("2:lop"), WithLabel("created")

	// -------------------------------------------------------------------------
	// 3.2.27  Count (POST)
	// -------------------------------------------------------------------------
	// client.Traverser.Count.WithReqData(CountRequestData{
	//     Source: "1:marko",
	//     Steps:  ...,
	// })
}
