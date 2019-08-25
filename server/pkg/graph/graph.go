package graph

import (
	"strings"

	"github.com/dpb587/tsg/pkg/graph"
	"github.com/dpb587/tsg/pkg/graph/dataset/filesystem"
	"github.com/dpb587/tsg/pkg/graph/dataset/multiple"
	"github.com/dpb587/tsg/pkg/graph/dataset/transformer"
	"github.com/dpb587/tsg/pkg/graph/index/inmemory"
	"github.com/dpb587/tsg/pkg/graph/node"
	"github.com/dpb587/tsg/pkg/graph/schemaorg/schemaorgproperty"
)

func NewGraph() *graph.Graph {
	ds := multiple.NewDataset()

	transformers := transformer.NewTransformers(
		transformer.TransformerFunc(func(n node.ObjectNode) (node.ObjectNode, error) {
			if strings.Contains(n.ID(), "/posts/") && !strings.Contains(n.ID(), "index.json") {
				n.Add(node.NewAttributeNode(schemaorgproperty.Author, node.NewNodeRefScalarNode("https://dpb587.github.io/dpb587.me/content/id.md")))
			}

			return n, nil
		}),
	)

	tds := transformer.NewDataset(ds, transformers)
	g := graph.NewGraph(tds)

	transformers.AddTransformer(transformer.NewAutoIndexTransformer(g))

	idx := inmemory.NewIndex(ds)

	ds.AddDataset(filesystem.NewDataset(idx, "https://dpb587.github.io/dpb587.me/content", "content"))

	return g
}
