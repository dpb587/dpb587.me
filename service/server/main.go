package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	appCommit = "unknown"
	appBuilt  = "unknown"
)

func main() {
	serverTag := fmt.Sprintf("dpb587.me (commit %s; built %s; %s)", appCommit, appBuilt, runtime.Version())

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Logger())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			c.Response().Header().Set("Server", serverTag)

			return next(c)
		}
	})

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}

		if code >= http.StatusInternalServerError {
			c.Logger().Error(err)
		}

		c.Response().WriteHeader(code)

		err = c.File(fmt.Sprintf("docroot/internal/http-%d/index.html", code))
		if err == nil {
			return
		}

		c.String(code, fmt.Sprintf("HTTP %d: %s\n", code, http.StatusText(code)))
	}

	{
		upstream, err := url.Parse("https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/")
		if err != nil {
			panic(err)
		}

		rp := httputil.NewSingleHostReverseProxy(upstream)
		rpd := rp.Director
		rp.Director = func(r *http.Request) {
			rpd(r)
			r.Host = upstream.Host
		}

		rp.ModifyResponse = func(w *http.Response) error {
			if statusCode := w.StatusCode; statusCode >= 400 {
				if statusCode == http.StatusForbidden {
					statusCode = http.StatusNotFound
				}

				rbuf, err := ioutil.ReadAll(w.Body)
				if err != nil {
					e.Logger.Warnf("error reading upstream error body: %s", err)
				}

				buf, err := ioutil.ReadFile(fmt.Sprintf("docroot/internal/http-%d/index.html", statusCode))
				if err == nil {
					w.Body = ioutil.NopCloser(io.MultiReader(bytes.NewBuffer(buf), bytes.NewBufferString(fmt.Sprintf("\n<!-- upstream response\n%s\n-->", rbuf))))
					w.Header = http.Header{}
					w.Header.Set("Content-Type", "text/html")

					return nil
				}
			}

			w.Header.Set("Server", serverTag)

			for key := range w.Header {
				if strings.HasPrefix(strings.ToLower(key), "x-amz-") {
					w.Header.Del(key)
				}
			}

			return nil
		}

		e.GET("/asset/*", echo.WrapHandler(rp))
	}

	e.Use(middleware.Static(os.Args[1]))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("listening on :%s (commit %s; built %s; %s)\n", port, appCommit, appBuilt, runtime.Version())

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
