package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// sync with package-public.sh
var compressedExts = map[string]struct{}{
	".css":  {},
	".html": {},
	".js":   {},
	".svg":  {},
	".xml":  {},
}

type fileHandler struct {
	root         fs.StatFS
	rootPath     string
	redirectsMap map[string]string
}

type fileRequest struct {
	FilePath string

	UserPath          string
	CanonicalUserPath string

	CompressionExt      string
	CompressionEncoding string

	ExtraHeaders http.Header
}

func (h *fileHandler) serveFile(w http.ResponseWriter, r *http.Request, fr fileRequest) bool {
	frp := strings.TrimPrefix(fr.FilePath, "/")

	info, err := h.root.Stat(frp)
	if err != nil || info.IsDir() {
		return false
	}

	if len(fr.CanonicalUserPath) > 0 && fr.CanonicalUserPath != r.URL.Path {
		http.Redirect(w, r, fr.CanonicalUserPath, http.StatusFound)

		return true
	}

	f, err := h.root.Open(frp)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		slog.Error("failed to open file after stat", "path", fr.FilePath, "error", err)

		return true
	}

	defer f.Close()

	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))

	if fr.CompressionEncoding != "" {
		w.Header().Set("Content-Encoding", fr.CompressionEncoding)
	}

	if len(fr.ExtraHeaders) > 0 {
		wh := w.Header()

		for key, values := range fr.ExtraHeaders {
			for _, value := range values {
				wh.Add(key, value)
			}
		}
	}

	http.ServeContent(w, r, strings.TrimSuffix(filepath.Base(fr.FilePath), fr.CompressionExt), info.ModTime(), f.(io.ReadSeeker))

	return true
}

func (h *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var compressionExt, compressionEncoding string

	if acceptEncoding := r.Header.Get("Accept-Encoding"); len(acceptEncoding) > 0 {
		if strings.Contains(acceptEncoding, "br") {
			compressionExt = ".br"
			compressionEncoding = "br"
		} else if strings.Contains(acceptEncoding, "zstd") {
			compressionExt = ".zst"
			compressionEncoding = "zstd"
		} else if strings.Contains(acceptEncoding, "gzip") {
			compressionExt = ".gz"
			compressionEncoding = "gzip"
		}
	}

	if strings.HasSuffix(r.URL.Path, "/index.html") {
		if h.serveFile(w, r, fileRequest{
			FilePath:            r.URL.Path + compressionExt,
			CompressionExt:      compressionExt,
			CompressionEncoding: compressionEncoding,
			UserPath:            r.URL.Path,
			CanonicalUserPath:   strings.TrimSuffix(r.URL.Path, "/index.html"),
		}) {
			return
		}
	}

	if strings.HasSuffix(r.URL.Path, "/") || !strings.Contains(r.URL.Path, ".") {
		if h.serveFile(w, r, fileRequest{
			FilePath:            filepath.Join(r.URL.Path, "index.html"+compressionExt),
			CompressionExt:      compressionExt,
			CompressionEncoding: compressionEncoding,
			UserPath:            r.URL.Path,
			CanonicalUserPath:   strings.TrimSuffix(r.URL.Path, "/"),
		}) {
			return
		}

		if h.serveFile(w, r, fileRequest{
			FilePath:          filepath.Join(r.URL.Path, "index.html"),
			UserPath:          r.URL.Path,
			CanonicalUserPath: strings.TrimSuffix(r.URL.Path, "/"),
		}) {
			return
		}
	}

	//

	var extraHeaders http.Header

	if strings.HasPrefix(r.URL.Path, "/assets/") {
		extraHeaders = http.Header{
			"Cache-Control": {"public, max-age=604800"},
		}
	}

	if _, ok := compressedExts[filepath.Ext(r.URL.Path)]; ok {
		if h.serveFile(w, r, fileRequest{
			FilePath:            r.URL.Path + compressionExt,
			CompressionExt:      compressionExt,
			CompressionEncoding: compressionEncoding,
			UserPath:            r.URL.Path,
			ExtraHeaders:        extraHeaders,
		}) {
			return
		}
	}

	if h.serveFile(w, r, fileRequest{
		FilePath:     r.URL.Path,
		ExtraHeaders: extraHeaders,
	}) {
		return
	}

	if dest, ok := h.redirectsMap[r.URL.Path]; ok {
		http.Redirect(w, r, dest, http.StatusMovedPermanently)

		return
	}

	// 7. Otherwise, 404
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, "Not Found")
}

func main() {
	publicDir := os.Args[1]
	redirectsDir := os.Args[2]

	type redirectMap map[string]string

	redirects, err := (func() (redirectMap, error) {
		redirects := redirectMap{}

		for _, redirectFilePath := range []string{
			"v2-generated.csv",
			"v3-content-generated.csv",
			"v3-content-manual.csv",
			"v4-manual.csv",
		} {
			fh, err := os.OpenFile(filepath.Join(redirectsDir, redirectFilePath), os.O_RDONLY, 0)
			if err != nil {
				return redirectMap{}, fmt.Errorf("open %s: %v", redirectFilePath, err)
			}

			defer fh.Close()

			r := csv.NewReader(fh)

			for {
				record, err := r.Read()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}

					return redirectMap{}, fmt.Errorf("read %s: %v", redirectFilePath, err)
				}

				redirects[record[0]] = record[1]
			}
		}

		return redirects, nil
	})()
	if err != nil {
		panic(fmt.Errorf("load redirects: %v", err))
	}

	slog.Info("Loaded redirects", "count", len(redirects))

	mux := http.NewServeMux()

	{
		upstream, err := url.Parse("https://storage.googleapis.com/dpb587-www-tilde-us-central1/")
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

				defer w.Body.Close()

				rbuf, err := io.ReadAll(w.Body)
				if err != nil {
					slog.Error("error reading upstream error body", "error", err)
				}

				w.StatusCode = statusCode
				w.Status = http.StatusText(statusCode)
				w.Body = io.NopCloser(bytes.NewBufferString(fmt.Sprintf("HTTP %d: %s\n", statusCode, http.StatusText(statusCode))))

				if statusCode == http.StatusNotFound {
					buf, err := os.ReadFile(filepath.Join(publicDir, "404.html"))
					if err == nil {
						w.Body = io.NopCloser(io.MultiReader(
							bytes.NewBuffer(buf),
							bytes.NewBufferString(fmt.Sprintf("\n<!-- upstream %s -->", base64.RawStdEncoding.EncodeToString(rbuf))),
						))
						w.Header = http.Header{}
						w.Header.Set("Content-Type", "text/html")

						return nil
					}
				}
			}

			for key := range w.Header {
				if strings.HasPrefix(strings.ToLower(key), "x-amz-") {
					w.Header.Del(key)
				}
			}

			return nil
		}

		mux.Handle("/~/blob-geojson/*", rp)
		mux.Handle("/~/blob-iiif-image-v3/*", rp)
	}

	mux.Handle("/", &fileHandler{
		root:         os.DirFS(publicDir).(fs.StatFS),
		rootPath:     publicDir,
		redirectsMap: redirects,
	})

	slog.Info("Serving files", "directory", publicDir)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Starting server", "url", "http://localhost:"+port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("could not start server", "error", err)
		os.Exit(1)
	}
}
