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

	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/hgtransport"
)

// basePath constructs the graph/graphspace URL prefix for traverser APIs.
//
// When config.GraphSpace is empty, returns "/graphs/{graph}".
// Otherwise returns "/graphspaces/{graphspace}/graphs/{graph}".
func basePath(cfg hgtransport.Config) string {
	if cfg.GraphSpace != "" {
		return fmt.Sprintf("/graphspaces/%s/graphs/%s", cfg.GraphSpace, cfg.Graph)
	}
	return fmt.Sprintf("/graphs/%s", cfg.Graph)
}
