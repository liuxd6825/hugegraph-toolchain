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

package cypher

import (
	"github.com/apache/hugegraph-toolchain/hugegraph-client-go/api"
)

type request struct {
	graph      string
	graphSpace string
	cypher     string
}
type Cypher struct {
	Get
	Post
}

func New(t api.Transport) *Cypher {
	return &Cypher{
		Get:  newGetFunc(t),
		Post: newPostFunc(t),
	}
}

type Get func(o ...func(*GetRequest)) (*GetResponse, error)
type Post func(o ...func(*PostRequest)) (*PostResponse, error)

func newGetFunc(t api.Transport) Get {
	return func(o ...func(*GetRequest)) (*GetResponse, error) {
		var r = GetRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}
func newPostFunc(t api.Transport) Post {
	return func(o ...func(*PostRequest)) (*PostResponse, error) {
		var r = PostRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

func (r request) buildDefaultAliases(transport api.Transport) map[string]string {
	cfg := transport.GetConfig()

	graphSpace := cfg.GraphSpace
	if r.graphSpace != "" {
		graphSpace = r.graphSpace
	}
	if graphSpace == "" {
		graphSpace = "DEFAULT"
	}

	graph := cfg.Graph
	if r.graph == "" {
		graph = r.graph
	}
	if graph == "" {
		graph = "hugegraph"
	}

	full := graphSpace + "-" + graph
	return map[string]string{
		"graph": full,
		"g":     "__g_" + full,
	}
}

func getString(strings ...string) string {
	for _, str := range strings {
		if str != "" {
			return str
		}
	}
	return ""
}
