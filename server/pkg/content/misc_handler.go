package content

import (
	"net/http"
	"strings"

	"github.com/dpb587/tsg/pkg/frontend/request"
	"github.com/dpb587/tsg/pkg/frontend/router"
	"github.com/dpb587/tsg/pkg/graph"
	"github.com/dpb587/tsg/pkg/graph/node"
	"github.com/dpb587/tsg/pkg/graph/schemaorg/schemaorgproperty"
	"github.com/dpb587/tsg/pkg/graph/schemaorg/schemaorgtype"
	"github.com/dpb587/tsg/pkg/graph/schemaorg/webpagebuilder"
	"github.com/gorilla/mux"
)

type MiscHandler struct {
	g *graph.Graph
	r *router.Router
	m *request.Manager
}

func NewMiscHandler(g *graph.Graph, r *router.Router, m *request.Manager) *MiscHandler {
	return &MiscHandler{
		g: g,
		r: r,
		m: m,
	}
}

func (h *MiscHandler) HandleError404(w http.ResponseWriter, r *http.Request) {
	if mux.CurrentRoute(r) == nil {
		targets, _, err := h.g.Search(
			node.NewAttributeValueMatcher(
				schemaorgproperty.URL,
				node.NewString(r.URL.Path),
			),
			graph.SearchOptions{},
		)
		if err != nil {
			panic(err)
		}

		if len(targets) > 0 {
			nu, found, err := h.r.FrontendLink(targets[0].(node.ObjectNode).ID(), true)
			if err != nil {
				panic(err)
			}

			if found {
				http.Redirect(w, r, nu, http.StatusTemporaryRedirect)

				return
			}
		}

		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			url := r.URL
			url.Path = strings.TrimSuffix(url.Path, "/")

			nr, err := http.NewRequest(r.Method, url.String(), nil)
			if err != nil {
				panic(err)
			}

			m := mux.RouteMatch{}

			if h.r.Frontend().Match(nr, &m) {
				http.Redirect(w, r, url.String(), http.StatusTemporaryRedirect)

				return
			}
		}
	}

	mainEntity, err := h.g.Get("https://dpb587.github.io/dpb587.me/content/internal/http-404.md")
	if err != nil {
		panic(err)
	}

	n := node.NewObjectNode(schemaorgtype.WebPage, "")
	n.Add(node.NewAttributeNode(schemaorgproperty.Error, node.NewString("http404")))
	n.Add(node.NewAttributeNode(schemaorgproperty.MainEntity, mainEntity))

	err = webpagebuilder.DefaultBuilder{}.BuildWebPage(h.g, n)
	if err != nil {
		panic(err)
	}

	req := h.m.BuildRequest(w, r)
	req.SetStatus(http.StatusNotFound)
	req.SetSubject("mainEntity", n)

	h.m.Handle(req)
}
