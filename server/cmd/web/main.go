package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/dpb587/dpb587.me/server/pkg/content"
	internalgraph "github.com/dpb587/dpb587.me/server/pkg/graph"
	"github.com/dpb587/tsg/pkg/frontend/handler"
	"github.com/dpb587/tsg/pkg/frontend/request"
	"github.com/dpb587/tsg/pkg/frontend/router"
	"github.com/dpb587/tsg/pkg/util/httputil"
	"github.com/dpb587/tsg/pkg/util/urlutil"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	var listen = "0.0.0.0:8080"
	var externalHost = "localhost:8080"
	if v := os.Getenv("EXTERNAL_HOST"); v != "" {
		externalHost = v
	}

	g := internalgraph.NewGraph()
	r := router.NewRouter(
		mux.NewRouter().Host(externalHost).Subrouter(),
		mux.NewRouter().Schemes("https").Host("dpb587.github.io").PathPrefix("/dpb587.me/content/").Subrouter(),
	)

	themeDir := "theme"
	m := request.NewManager(g, r, path.Join(themeDir, "layouts"))

	r.NewGraphRoute(func(f, g *mux.Route) {
		g.Path(`/galleries/_index.md`)
		f.Path("/galleries").Methods("GET").HandlerFunc(m.HandleDefault)
	})

	r.NewGraphRoute(func(f, g *mux.Route) {
		g.Path(`/galleries/{gallery}/_index.md`)
		f.Path("/galleries/{gallery}").Methods("GET").HandlerFunc(m.HandleDefault)
	})

	r.NewGraphRoute(func(f, g *mux.Route) {
		g.Path(`/galleries/{gallery}/{photo}.md`)
		f.Path("/galleries/{gallery}/{photo}").Methods("GET").HandlerFunc(m.HandleDefault)
	})

	r.NewGraphRoute(func(f, g *mux.Route) {
		g.Path(`/posts/{year:[^\-]+}-{month:[^\-]+}-{day:[^\-]+}-{slug}.md`)
		f.Path("/posts/{year}/{month}/{day}/{slug}").Methods("GET").HandlerFunc(m.HandleDefault)
	})

	r.NewGraphRoute(func(f, g *mux.Route) {
		g.Path(`/posts/_index.md`)
		f.Path("/posts").Methods("GET").HandlerFunc(m.HandleDefault)
	})

	r.NewGraphRoute(func(f, g *mux.Route) {
		g.Path(`/pages/home.md`)
		f.Path("/").Methods("GET").HandlerFunc(m.HandleDefault)
	})

	r.NewGraphRoute(func(f, g *mux.Route) {
		g.Path(`/pages/about.md`)
		f.Path("/about").Methods("GET").HandlerFunc(m.HandleDefault)
	})

	r.NewGraphRoute(func(f, g *mux.Route) {
		g.Path(`/pages/postsfeed.md`)
		f.Path("/index.xml").Methods("GET").HandlerFunc(m.HandleDefault)
	})

	miscHandler := content.NewMiscHandler(g, r, m)

	{
		frontendRouter := r.Frontend()

		frontendRouter.PathPrefix("/asset/").Handler(
			httputil.NewHeaderSingleHostReverseProxy(
				urlutil.MustParse("https://dpb587-website-us-east-1.s3.amazonaws.com/asset/"),
			),
		)

		{
			staticHandler := http.FileServer(http.Dir(path.Join(themeDir, "static")))

			r.Frontend().PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticHandler))
			r.Frontend().Path("/favicon.ico").Handler(staticHandler)
		}

		handler.NewGraphHandler(g).Install(frontendRouter.PathPrefix("/api/v1/graph").Subrouter())

		frontendRouter.NotFoundHandler = http.HandlerFunc(miscHandler.HandleError404)
		// frontendRouter.Use(handlers.CompressHandler)
		frontendRouter.Use(handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)))
	}

	srv := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, r),
		Addr:         listen,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	fmt.Printf("listening on http://%s/\n", listen)

	log.Fatal(srv.ListenAndServe())
}
