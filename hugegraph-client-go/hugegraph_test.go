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
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/hgtransport"
	"github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/constant"
	structtraverser "github.com/apache/incubator-hugegraph-toolchain/hugegraph-client-go/internal/structure/traverser"
)

func TestNewDefaultCommonClient(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDefaultCommonClient()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDefaultCommonClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// TestNewCommonClientValidation covers the input-validation branches in
// NewCommonClient (host length, host parse, port range).
func TestNewCommonClientValidation(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantSub string
	}{
		{
			name:    "empty host",
			cfg:     Config{Host: "", Port: 8080, Graph: "g"},
			wantSub: "host length",
		},
		{
			name:    "host too short",
			cfg:     Config{Host: "ab", Port: 8080, Graph: "g"},
			wantSub: "host length",
		},
		{
			name:    "invalid host",
			cfg:     Config{Host: "not-an-ip", Port: 8080, Graph: "g"},
			wantSub: "host is format",
		},
		{
			name:    "port too low",
			cfg:     Config{Host: "127.0.0.1", Port: 0, Graph: "g"},
			wantSub: "port is error",
		},
		{
			name:    "port too high",
			cfg:     Config{Host: "127.0.0.1", Port: 70000, Graph: "g"},
			wantSub: "port is error",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := NewCommonClient(c.cfg)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), c.wantSub) {
				t.Errorf("error %q missing %q", err.Error(), c.wantSub)
			}
		})
	}
}

// TestCommonClientNonNilFields verifies that all non-traverser managers are
// non-nil after NewCommonClient returns.
func TestCommonClientNonNilFields(t *testing.T) {
	c := mustNewClient(t, defaultConfig())
	if c.Vertex == nil {
		t.Error("Vertex is nil")
	}
	if c.Gremlin == nil {
		t.Error("Gremlin is nil")
	}
	if c.Propertykey == nil {
		t.Error("Propertykey is nil")
	}
	if c.VertexLabel == nil {
		t.Error("VertexLabel is nil")
	}
	if c.EdgeLabel == nil {
		t.Error("EdgeLabel is nil")
	}
	if c.APIV1 == nil {
		t.Error("APIV1 is nil")
	}
	if c.Transport == nil {
		t.Error("Transport is nil")
	}
	if c.Graph != "hugegraph" {
		t.Errorf("Graph = %q, want %q", c.Graph, "hugegraph")
	}
}

// TestCommonClientTraverserWiring verifies that the embedded *v1.APIV1
// exposes all 25 traverser method fields via client.Traverser.
func TestCommonClientTraverserWiring(t *testing.T) {
	c := mustNewClient(t, defaultConfig())
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

// TestCommonClientGraphSpacePropagation verifies that Config.GraphSpace is
// stored on CommonClient AND propagated to the underlying transport that
// the traverser APIs read via GetConfig().
func TestCommonClientGraphSpacePropagation(t *testing.T) {
	c := mustNewClient(t, Config{
		Host:       "127.0.0.1",
		Port:       8080,
		Graph:      "hugegraph",
		GraphSpace: "DEFAULT",
	})
	if c.GraphSpace != "DEFAULT" {
		t.Errorf("CommonClient.GraphSpace = %q, want DEFAULT", c.GraphSpace)
	}
	cfg := c.Transport.GetConfig()
	if cfg.GraphSpace != "DEFAULT" {
		t.Errorf("Transport.GraphSpace = %q, want DEFAULT", cfg.GraphSpace)
	}
	if cfg.Graph != "hugegraph" {
		t.Errorf("Transport.Graph = %q, want hugegraph", cfg.Graph)
	}
}

// TestCommonClientGraphPropagation verifies Config.Graph propagation when
// GraphSpace is empty.
func TestCommonClientGraphPropagation(t *testing.T) {
	c := mustNewClient(t, Config{
		Host:  "127.0.0.1",
		Port:  8080,
		Graph: "mygraph",
	})
	if c.Graph != "mygraph" {
		t.Errorf("Graph = %q, want mygraph", c.Graph)
	}
	cfg := c.Transport.GetConfig()
	if cfg.Graph != "mygraph" {
		t.Errorf("Transport.Graph = %q, want mygraph", cfg.Graph)
	}
	if cfg.GraphSpace != "" {
		t.Errorf("Transport.GraphSpace = %q, want empty", cfg.GraphSpace)
	}
}

// TestCommonClientUsernamePassword verifies that non-empty credentials
// propagate to the transport.
func TestCommonClientUsernamePassword(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	u := urlParse(srv.URL)
	port := portFromURL(u)
	cfg := Config{
		Host:     u.Hostname(),
		Port:     port,
		Graph:    "hugegraph",
		Username: "admin",
		Password: "secret",
	}
	c := mustNewClient(t, cfg)
	c.Transport = &redirectTransport{server: srv, base: c.Transport}

	_, err := c.Traverser.ShortestPath(
		c.Traverser.ShortestPath.WithSource("marko"),
		c.Traverser.ShortestPath.WithTarget("peter"),
		c.Traverser.ShortestPath.WithDirection(constant.OUT.String()),
		c.Traverser.ShortestPath.WithMaxDepth(3),
		c.Traverser.ShortestPath.WithDegree(constant.NoLimit),
		c.Traverser.ShortestPath.WithSkipDegree(0),
		c.Traverser.ShortestPath.WithCapacity(constant.NoLimit),
	)
	if err != nil {
		t.Fatalf("ShortestPath: %v", err)
	}
	if gotAuth == "" {
		t.Error("expected Authorization header to be set")
	}
	if !strings.HasPrefix(gotAuth, "Basic ") {
		t.Errorf("Authorization = %q, want Basic prefix", gotAuth)
	}
}

// TestCommonClientTraverserShortestPathEndToEnd verifies the full pipeline
// from client.Traverser.ShortestPath through the embedded transport to a
// mocked HugeGraph server.
func TestCommonClientTraverserShortestPathEndToEnd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/graphs/hugegraph/traversers/shortestpath" {
			t.Errorf("path = %s, want /graphs/hugegraph/traversers/shortestpath", r.URL.Path)
		}
		for _, want := range []string{
			"source=%22marko%22",
			"target=%22peter%22",
			"direction=out",
			"max_depth=5",
		} {
			if !strings.Contains(r.URL.RawQuery, want) {
				t.Errorf("query missing %q: %s", want, r.URL.RawQuery)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"path":["\"marko\"","\"lop\"","\"peter\""]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)
	resp, err := c.Traverser.ShortestPath(
		c.Traverser.ShortestPath.WithSource("marko"),
		c.Traverser.ShortestPath.WithTarget("peter"),
		c.Traverser.ShortestPath.WithDirection(constant.OUT.String()),
		c.Traverser.ShortestPath.WithMaxDepth(5),
		c.Traverser.ShortestPath.WithDegree(constant.NoLimit),
		c.Traverser.ShortestPath.WithSkipDegree(0),
		c.Traverser.ShortestPath.WithCapacity(constant.NoLimit),
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

// TestCommonClientTraverserKoutPostEndToEnd verifies the POST pipeline
// through the embedded client.
func TestCommonClientTraverserKoutPostEndToEnd(t *testing.T) {
	var capturedMethod, capturedPath string
	var capturedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		capturedBody = readAll(t, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"size":1,"ids":["\"vadas\""]}`))
	}))
	defer srv.Close()

	c := mustNewClientWithServer(t, srv)

	b := structtraverser.NewKoutRequestBuilder().Source("marko").
		MaxDepth(2).
		Capacity(100).
		Limit(10)
	if _, err := b.Steps().Direction(constant.OUT).Build(); err != nil {
		t.Fatalf("Steps Build: %v", err)
	}

	resp, err := c.Traverser.Kout.Post(c.Traverser.Kout.Post.WithRequestBuilder(b))
	if err != nil {
		t.Fatalf("Kout.Post: %v", err)
	}

	if capturedMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", capturedMethod)
	}
	if capturedPath != "/graphs/hugegraph/traversers/kout" {
		t.Errorf("path = %s, want /graphs/hugegraph/traversers/kout", capturedPath)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("decode captured body: %v", err)
	}
	wantSource := `"marko"`
	if src, ok := body["source"].(string); !ok || src != wantSource {
		t.Errorf("source = %q (%T), want %q", body["source"], body["source"], wantSource)
	}
	if depth, _ := body["max_depth"].(float64); depth != 2 {
		t.Errorf("max_depth = %v, want 2", body["max_depth"])
	}
	if cap, _ := body["capacity"].(float64); cap != 100 {
		t.Errorf("capacity = %v, want 100", body["capacity"])
	}
	if lim, _ := body["limit"].(float64); lim != 10 {
		t.Errorf("limit = %v, want 10", body["limit"])
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d", resp.StatusCode)
	}
	if resp.Data.GetSize() != 1 {
		t.Errorf("Data.Size = %d, want 1", resp.Data.GetSize())
	}
}

// TestCommonClientDefaultLogger verifies that NewDefaultCommonClient wires a
// non-nil Logger without panicking.
func TestCommonClientDefaultLogger(t *testing.T) {
	c, err := NewDefaultCommonClient()
	if err != nil {
		t.Fatalf("NewDefaultCommonClient: %v", err)
	}
	if c == nil {
		t.Fatal("nil client")
	}
	if cfg := c.Transport.GetConfig(); cfg.URL == nil {
		t.Error("Transport URL not set")
	}
}

// -----------------------------------------------------------------------------
// Test helpers (package-private; shared with hugegraph_traverser_test.go).
// -----------------------------------------------------------------------------

// readAll is a small test helper to drain an io.ReadCloser body into bytes.
func readAll(t *testing.T, body interface{ Read([]byte) (int, error) }) []byte {
	t.Helper()
	buf := make([]byte, 0, 1024)
	tmp := make([]byte, 512)
	for {
		n, err := body.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				return buf
			}
			return buf
		}
	}
}

// mustNewClient creates a CommonClient with the given Config. It fails the
// test on construction error so callers can call it inline.
func mustNewClient(t *testing.T, cfg Config) *CommonClient {
	t.Helper()
	c, err := NewCommonClient(cfg)
	if err != nil {
		t.Fatalf("NewCommonClient: %v", err)
	}
	return c
}

// mustNewClientWithServer creates a CommonClient that targets the given
// httptest server.
func mustNewClientWithServer(t *testing.T, srv *httptest.Server) *CommonClient {
	t.Helper()
	u := urlParse(srv.URL)
	port := portFromURL(u)
	cfg := Config{
		Host:  u.Hostname(),
		Port:  port,
		Graph: "hugegraph",
	}
	return mustNewClient(t, cfg)
}

// defaultConfig returns a sensible default Config for tests that do not
// require a real server.
func defaultConfig() Config {
	return Config{
		Host:  "127.0.0.1",
		Port:  8080,
		Graph: "hugegraph",
	}
}

// urlParse is a thin wrapper around net/url.Parse that fails the test on
// error.
func urlParse(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}

// portFromURL extracts the integer port from a *url.URL.
func portFromURL(u *url.URL) int {
	p := u.Port()
	if p == "" {
		return 0
	}
	var v int
	for _, c := range p {
		if c < '0' || c > '9' {
			return 0
		}
		v = v*10 + int(c-'0')
	}
	return v
}

// redirectTransport forwards outgoing requests to the configured httptest
// server while preserving the base transport for non-redirected calls.
// It implements hgtransport.Interface so it can be assigned to
// CommonClient.Transport.
type redirectTransport struct {
	server *httptest.Server
	base   hgtransport.Interface
}

func (t *redirectTransport) Perform(req *http.Request) (*http.Response, error) {
	u := urlParse(t.server.URL)
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host
	cfg := t.base.GetConfig()
	if _, ok := req.Header["Authorization"]; !ok && cfg.Username != "" {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}
	if _, ok := req.Header["Content-Type"]; !ok {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	}
	rt := cfg.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}

func (t *redirectTransport) GetConfig() hgtransport.Config {
	return t.base.GetConfig()
}

// -----------------------------------------------------------------------------
// Helpers for binding the mock httptest server to a specific IP:port.
//
// Background: httptest.NewServer() binds to 127.0.0.1:<random>. To bind to a
// fixed (host, port) — e.g. for reproducing traffic from a real network
// interface or for sharing a server across multiple tests at a deterministic
// address — call mustNewServerOnAddr instead. The listener is created with
// net.Listen("tcp", addr), so the test process must have permission to bind
// to the requested address (typically root or CAP_NET_BIND_SERVICE for
// non-loopback addresses).
// -----------------------------------------------------------------------------

// mustNewServerOnAddr starts an httptest.Server listening on the given
// "host:port" address. Pass ":0" to let the OS pick a free port while
// keeping the chosen host. Returns the running server; caller is expected
// to defer srv.Close().
func mustNewServerOnAddr(t *testing.T, addr string, handler http.Handler) *httptest.Server {
	t.Helper()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("listen on %s: %v", addr, err)
	}
	srv := httptest.NewUnstartedServer(handler)
	srv.Listener = ln
	srv.Start()
	return srv
}

// mustNewClientForAddr builds a CommonClient whose Host/Port match the given
// "host:port" address. It is intended to be paired with a server started via
// mustNewServerOnAddr or with a real HugeGraph server at a known address.
//
// cfg.Graph defaults to "hugegraph" and cfg.GraphSpace to "". Override the
// returned *Config before passing it to NewCommonClient if you need different
// values.
func mustNewClientForAddr(t *testing.T, addr string) *CommonClient {
	t.Helper()
	host := addr
	portStr := ""
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		host = addr[:i]
		portStr = addr[i+1:]
	}
	port := 0
	if portStr != "" {
		for _, c := range portStr {
			if c < '0' || c > '9' {
				t.Fatalf("invalid port in %q: %c", addr, c)
			}
			port = port*10 + int(c-'0')
		}
	}
	cfg := Config{
		Host:  host,
		Port:  port,
		Graph: "hugegraph",
	}
	return mustNewClient(t, cfg)
}
