package export

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	TextContentAPI    = "https://api.namedgraph.com/toolkit.v0/textContent.export"
	StructuredDataAPI = "https://api.namedgraph.com/toolkit.v0/structuredData.export"
	SourceURIPrefix   = "https://github.com/dpb587/dpb587.me/blob/main/content/"
)

// FileResolver resolves a URI path (e.g. "/post/foo") to an absolute filesystem path.
type FileResolver interface {
	Resolve(uriPath string) (string, error)
}

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
)

// Handler serves the /~/export/* endpoints.
// Set CacheDir to enable file-based caching of upstream API responses.
// Set ContentDir to enable /~/export/source.
type Handler struct {
	PublicResolver FileResolver // resolves request URI paths to absolute HTML file paths
	ContentDir     string       // absolute path to content directory (for source export)
	CacheDir       string       // if non-empty, successful API responses are cached here
}

// HandleTextContent handles /~/export/text-content requests.
func (h *Handler) HandleTextContent(w http.ResponseWriter, r *http.Request) {
	h.ngExport(w, r, TextContentAPI, r.URL.Query().Get("format"), nil)
}

// HandleStructuredData handles /~/export/structured-data requests.
func (h *Handler) HandleStructuredData(w http.ResponseWriter, r *http.Request) {
	h.ngExport(w, r, StructuredDataAPI, r.URL.Query().Get("format"), map[string]string{
		"experimental": "ontology.schema.mapper=xsd",
	})
}

// HandleSource handles /~/export/source requests.
func (h *Handler) HandleSource(w http.ResponseWriter, r *http.Request) {
	if h.ContentDir == "" {
		http.NotFound(w, r)
		return
	}

	uriParam := r.URL.Query().Get("uri")
	if uriParam == "" {
		http.Error(w, "Missing required query parameter: uri", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(uriParam, SourceURIPrefix) {
		http.Error(w, "uri must start with "+SourceURIPrefix, http.StatusBadRequest)
		return
	}

	relativePath := strings.TrimPrefix(uriParam, SourceURIPrefix)
	localPath, ok := safeLocalPath(h.ContentDir, relativePath)
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Link", fmt.Sprintf(`<%s>; rel=canonical`, uriParam))
	http.ServeFile(w, r, localPath)
}

func (h *Handler) ngExport(w http.ResponseWriter, r *http.Request, apiURL, format string, extraFields map[string]string) {
	uriParam := r.URL.Query().Get("uri")
	if uriParam == "" {
		http.Error(w, "Missing required query parameter: uri", http.StatusBadRequest)
		return
	}

	localPath, err := h.PublicResolver.Resolve(uriParam)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	h.serveNgExport(w, r, apiURL, format, extraFields, uriParam, localPath, nil)
}

// ServeTextContent serves the text-content export for uriPath directly,
// without requiring a ?uri query parameter. Intended for content negotiation
// where the path is already known from the request URL.
// Returns false if uriPath could not be resolved (no response written),
// true once a response has been written.
func (h *Handler) ServeTextContent(w http.ResponseWriter, r *http.Request, uriPath string) bool {
	localPath, err := h.PublicResolver.Resolve(uriPath)
	if err != nil {
		return false
	}
	okHeaders := http.Header{
		"Content-Location": {"/~/export/text-content?uri=" + url.QueryEscape(uriPath) + "&format=markdown"},
	}
	h.serveNgExport(w, r, TextContentAPI, "markdown", nil, uriPath, localPath, okHeaders)
	return true
}

func (h *Handler) serveNgExport(w http.ResponseWriter, r *http.Request, apiURL, format string, extraFields map[string]string, uriParam, localPath string, okHeaders http.Header) {
	fileData, err := os.ReadFile(localPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		slog.Error("export: failed to read resolved file", "path", localPath, "error", err)
		return
	}

	if h.CacheDir != "" {
		key := computeCacheKey(apiURL, format, uriParam, fileData)
		pipeHeaders := http.Header{
			"Link": {fmt.Sprintf(`<%s>; rel=canonical`, uriParam)},
		}
		for k, vs := range okHeaders {
			pipeHeaders[k] = vs
		}
		if hit, err := pipeCache(w, h.CacheDir, key, pipeHeaders); hit {
			if err != nil {
				slog.Error("export: failed to pipe cache to client", "error", err)
			}
			return
		}
	}

	respBody, statusCode, respHeaders, err := callAPI(apiURL, format, uriParam, filepath.Base(localPath), fileData, extraFields)
	if err != nil {
		http.Error(w, "Failed to reach upstream", http.StatusBadGateway)
		slog.Error("export: upstream API call failed", "api", apiURL, "error", err)
		return
	}

	if h.CacheDir != "" && statusCode == http.StatusOK {
		key := computeCacheKey(apiURL, format, uriParam, fileData)
		if cacheErr := writeCache(h.CacheDir, key, respHeaders, respBody); cacheErr != nil {
			slog.Error("export: failed to write cache", "error", cacheErr)
		}
	}

	for k, vs := range respHeaders {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	if statusCode == http.StatusOK {
		w.Header().Set("Link", fmt.Sprintf(`<%s>; rel=canonical`, uriParam))
		for k, vs := range okHeaders {
			w.Header()[k] = vs
		}
	}
	w.WriteHeader(statusCode)
	w.Write(respBody)
}

func callAPI(apiURL, format, uriParam, filename string, fileData []byte, extraFields map[string]string) (body []byte, statusCode int, headers http.Header, err error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormFile("sourceFile", filename)
	if err != nil {
		return nil, 0, nil, err
	}
	if _, err = fw.Write(fileData); err != nil {
		return nil, 0, nil, err
	}
	if err = mw.WriteField("format", format); err != nil {
		return nil, 0, nil, err
	}
	if err = mw.WriteField("baseUrl", "https://dpb587.me"+uriParam); err != nil {
		return nil, 0, nil, err
	}
	for k, v := range extraFields {
		if err = mw.WriteField(k, v); err != nil {
			return nil, 0, nil, err
		}
	}
	mw.Close()

	req, err := http.NewRequest(http.MethodPost, apiURL, &buf)
	if err != nil {
		return nil, 0, nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if token := os.Getenv("NG_API_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, nil, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, nil, err
	}

	return body, resp.StatusCode, resp.Header, nil
}

func computeCacheKey(apiURL, format, uriParam string, fileData []byte) string {
	h := sha256.New()
	h.Write([]byte(apiURL))
	h.Write([]byte{0})
	h.Write([]byte(format))
	h.Write([]byte{0})
	h.Write([]byte(uriParam))
	h.Write([]byte{0})
	h.Write(fileData)
	return hex.EncodeToString(h.Sum(nil))
}

func pipeCache(w http.ResponseWriter, cacheDir, key string, okHeaders http.Header) (bool, error) {
	data, err := os.ReadFile(filepath.Join(cacheDir, key))
	if err != nil {
		return false, nil
	}
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(data)), nil)
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()

	wh := w.Header()
	for k, vs := range resp.Header {
		wh[k] = vs
	}
	for k, vs := range okHeaders {
		wh[k] = vs
	}
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, resp.Body)
	return true, err
}

func writeCache(cacheDir, key string, headers http.Header, body []byte) error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	keptHeaders := http.Header{}
	for _, name := range []string{"Content-Type", "Content-Disposition"} {
		if vs := headers.Values(name); len(vs) > 0 {
			keptHeaders[name] = vs
		}
	}
	resp := &http.Response{
		StatusCode:    http.StatusOK,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        keptHeaders,
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
	}
	data, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cacheDir, key), data, 0644)
}
