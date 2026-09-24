package cypher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/apache/hugegraph-toolchain/hugegraph-client-go/api"
)

type PostRequest struct {
	request
	ctx      context.Context
	body     io.ReadCloser
	bindings map[string]string
	language string
	aliases  map[string]string
}
type PostRequestData struct {
	Cypher   string            `json:"cypher"`
	Bindings map[string]string `json:"bindings,omitempty"`
	Language string            `json:"language,omitempty"`
	Aliases  map[string]string `json:"aliases,omitempty"`
}
type PostResponse struct {
	StatusCode int           `json:"-"`
	Header     http.Header   `json:"-"`
	Body       io.ReadCloser `json:"-"`
	Data       *ResponseData `json:"data"`
}

func (g Post) WithCypher(cypher string) func(request *PostRequest) {
	return func(r *PostRequest) {
		r.cypher = cypher
	}
}

func (g Post) WithGraph(graph string) func(request *PostRequest) {
	return func(r *PostRequest) {
		r.graph = graph
	}
}

func (g Post) WithGraphSpace(space string) func(request *PostRequest) {
	return func(r *PostRequest) {
		r.graphSpace = space
	}
}

func (g PostRequest) Do(ctx context.Context, transport api.Transport) (*PostResponse, error) {
	if len(g.cypher) < 1 {
		return nil, errors.New("PostRequest param error , gremlin is empty")
	}

	graphSpace := getString(g.graphSpace, transport.GetConfig().GraphSpace, "DEFAULT")
	graph := getString(g.graph, transport.GetConfig().Graph, "hugegraph")

	url := fmt.Sprintf("/graphspaces/%s/graphs/%s/cypher", graphSpace, graph)

	reader := strings.NewReader(g.cypher)
	req, _ := api.NewRequest("POST", url, nil, reader)

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	res, err := transport.Perform(req)
	if err != nil {
		return nil, err
	}

	gremlinPostResp := &PostResponse{}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	respData := &ResponseData{}
	err = json.Unmarshal(bytes, respData)
	if err != nil {
		return nil, err
	}
	gremlinPostResp.StatusCode = res.StatusCode
	gremlinPostResp.Header = res.Header
	gremlinPostResp.Body = res.Body
	gremlinPostResp.Data = respData
	return gremlinPostResp, nil
}
