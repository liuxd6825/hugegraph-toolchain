package cypher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	url2 "net/url"

	"github.com/apache/hugegraph-toolchain/hugegraph-client-go/api"
)

type GetRequest struct {
	request
	ctx context.Context
}
type GetResponse struct {
	StatusCode int               `json:"-"`
	Header     http.Header       `json:"-"`
	Body       io.ReadCloser     `json:"-"`
	Data       *PostResponseData `json:"data"`
}

type GetResponseData struct {
	RequestID string `json:"requestId,omitempty"`
	Status    struct {
		Message    string `json:"message"`
		Code       int    `json:"code"`
		Attributes struct {
		} `json:"attributes"`
	} `json:"status"`
	Result struct {
		Data interface{} `json:"data"`
		Meta interface{} `json:"meta"`
	} `json:"result,omitempty"`
	Exception string   `json:"exception,omitempty"`
	Message   string   `json:"message,omitempty"`
	Cause     string   `json:"cause,omitempty"`
	Trace     []string `json:"trace,omitempty"`
}

func (g Get) WithCypher(cypher string) func(request *GetRequest) {
	return func(r *GetRequest) {
		r.cypher = cypher
	}
}
func (g Get) WithGraph(graph string) func(request *GetRequest) {
	return func(r *GetRequest) {
		r.graph = graph
	}
}

func (g Get) WithGraphSpace(space string) func(request *GetRequest) {
	return func(r *GetRequest) {
		r.graphSpace = space
	}
}

func (g GetRequest) Do(ctx context.Context, transport api.Transport) (*GetResponse, error) {

	params := &url2.Values{}
	if len(g.cypher) <= 0 {
		return nil, errors.New("please set gremlin")
	} else {
		params.Add("cypher", g.cypher)
	}

	graphSpace := getString(g.graphSpace, transport.GetConfig().GraphSpace, "DEFAULT")
	graph := getString(g.graph, transport.GetConfig().Graph, "hugegraph")

	url := fmt.Sprintf("/graphspaces/%s/graphs/%s/cypher", graphSpace, graph)

	req, err := api.NewRequest("GET", url, params, nil)
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	res, err := transport.Perform(req)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(bytes))

	gremlinGetResponse := &GetResponse{}
	gremlinGetResponse.StatusCode = res.StatusCode
	return gremlinGetResponse, nil
}
