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
	"fmt"

	"github.com/apache/hugegraph-toolchain/hugegraph-client-go/api"
)

// Traverser aggregates all traverser API entry points.
type Traverser struct {
	ShortestPath          ShortestPath
	AllShortestPaths      AllShortestPaths
	SameNeighbors         SameNeighbors
	JaccardSimilarity     JaccardSimilarity
	JaccardSimilarityPost JaccardSimilarityPost
	WeightedShortestPath  WeightedShortestPath
	SingleSourceShortest  SingleSourceShortestPath
	MultiNodeShortestPath MultiNodeShortestPath
	KoutBasic             KoutBasic
	KoutAdvanced          KoutAdvanced
	KneighborBasic        KneighborBasic
	KneighborAdvanced     KneighborAdvanced
	PathsBasic            PathsBasic
	PathsAdvanced         PathsAdvanced
	CustomizedPaths       CustomizedPaths
	TemplatePaths         TemplatePaths
	Crosspoints           Crosspoints
	CustomizedCrosspoints CustomizedCrosspoints
	Rings                 Rings
	Rays                  Rays
	FusiformSimilarity    FusiformSimilarity
	VerticesByID          VerticesByID
	VerticesShards        VerticesShards
	VerticesScan          VerticesScan
	EdgesByID             EdgesByID
	EdgesShards           EdgesShards
	EdgesScan             EdgesScan
	AdamicAdar            AdamicAdar
	ResourceAllocation    ResourceAllocation
	EdgeExistence         EdgeExistence
	Count                 Count
}

// New creates a new Traverser with all entry points bound to the transport.
func New(t api.Transport) *Traverser {
	return &Traverser{
		ShortestPath:          newShortestPathFunc(t),
		AllShortestPaths:      newAllShortestPathsFunc(t),
		SameNeighbors:         newSameNeighborsFunc(t),
		JaccardSimilarity:     newJaccardSimilarityFunc(t),
		JaccardSimilarityPost: newJaccardSimilarityPostFunc(t),
		WeightedShortestPath:  newWeightedShortestPathFunc(t),
		SingleSourceShortest:  newSingleSourceShortestPathFunc(t),
		MultiNodeShortestPath: newMultiNodeShortestPathFunc(t),
		KoutBasic:             newKoutBasicFunc(t),
		KoutAdvanced:          newKoutAdvancedFunc(t),
		KneighborBasic:        newKneighborBasicFunc(t),
		KneighborAdvanced:     newKneighborAdvancedFunc(t),
		PathsBasic:            newPathsBasicFunc(t),
		PathsAdvanced:         newPathsAdvancedFunc(t),
		CustomizedPaths:       newCustomizedPathsFunc(t),
		TemplatePaths:         newTemplatePathsFunc(t),
		Crosspoints:           newCrosspointsFunc(t),
		CustomizedCrosspoints: newCustomizedCrosspointsFunc(t),
		Rings:                 newRingsFunc(t),
		Rays:                  newRaysFunc(t),
		FusiformSimilarity:    newFusiformSimilarityFunc(t),
		VerticesByID:          newVerticesByIDFunc(t),
		VerticesShards:        newVerticesShardsFunc(t),
		VerticesScan:          newVerticesScanFunc(t),
		EdgesByID:             newEdgesByIDFunc(t),
		EdgesShards:           newEdgesShardsFunc(t),
		EdgesScan:             newEdgesScanFunc(t),
		AdamicAdar:            newAdamicAdarFunc(t),
		ResourceAllocation:    newResourceAllocationFunc(t),
		EdgeExistence:         newEdgeExistenceFunc(t),
		Count:                 newCountFunc(t),
	}
}

// getURL returns the URL for a traverser endpoint, automatically
// switching between the GraphSpace and non-GraphSpace form depending on the
// transport configuration. The suffix must not begin with a slash.
func getURL(transport api.Transport, suffix string) string {
	cfg := transport.GetConfig()
	if len(cfg.GraphSpace) > 0 {
		return fmt.Sprintf("/graphspaces/%s/graphs/%s/traversers/%s",
			cfg.GraphSpace, cfg.Graph, suffix)
	}
	return fmt.Sprintf("/graphs/%s/traversers/%s", cfg.Graph, suffix)
}
