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
	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/api"
)

// Traverser is the aggregator of all traverser APIs exposed by HugeGraph.
//
// It mirrors the structure of org.apache.hugegraph.api.traverser in the Java
// client, where each subfield corresponds to one RESTful endpoint documented
// at https://hugegraph.apache.org/docs/clients/restful-api/traverser/.
//
// Two-verb endpoints (e.g. Kout.Get + Kout.Post) are exposed as a nested
// pair type so callers can disambiguate by HTTP method.
type Traverser struct {
	ShortestPath             ShortestPath
	AllShortestPaths         AllShortestPaths
	MultiNodeShortestPath    MultiNodeShortestPath
	SingleSourceShortestPath SingleSourceShortestPath
	WeightedShortestPath     WeightedShortestPath
	Kout                     KoutPair
	Kneighbor                KneighborPair
	Crosspoints              Crosspoints
	CustomizedCrosspoints    CustomizedCrosspoints
	Paths                    PathsPair
	CustomizedPaths          CustomizedPaths
	TemplatePaths            TemplatePaths
	Rings                    Rings
	Rays                     Rays
	JaccardSimilarity        JaccardSimilarityPair
	FusiformSimilarity       FusiformSimilarity
	SameNeighbors            SameNeighbors
	Vertices                 VerticesPair
	Edges                    EdgesPair
}

// New creates a Traverser bound to the given api.Transport. Every subfield is
// initialised via the corresponding newXxxFunc helper.
func New(t api.Transport) *Traverser {
	return &Traverser{
		ShortestPath:             newShortestPathFunc(t),
		AllShortestPaths:         newAllShortestPathsFunc(t),
		MultiNodeShortestPath:    newMultiNodeShortestPathFunc(t),
		SingleSourceShortestPath: newSingleSourceShortestPathFunc(t),
		WeightedShortestPath:     newWeightedShortestPathFunc(t),
		Kout:                     newKoutPair(t),
		Kneighbor:                newKneighborPair(t),
		Crosspoints:              newCrosspointsFunc(t),
		CustomizedCrosspoints:    newCustomizedCrosspointsFunc(t),
		Paths:                    newPathsPair(t),
		CustomizedPaths:          newCustomizedPathsFunc(t),
		TemplatePaths:            newTemplatePathsFunc(t),
		Rings:                    newRingsFunc(t),
		Rays:                     newRaysFunc(t),
		JaccardSimilarity:        newJaccardSimilarityPair(t),
		FusiformSimilarity:       newFusiformSimilarityFunc(t),
		SameNeighbors:            newSameNeighborsFunc(t),
		Vertices:                 newVerticesPair(t),
		Edges:                    newEdgesPair(t),
	}
}
