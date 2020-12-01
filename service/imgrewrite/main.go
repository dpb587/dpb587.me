package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	if err := mainErr(); err != nil {
		panic(err)
	}
}

func mainErr() error {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Encoding = "console"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := loggerConfig.Build()
	defer logger.Sync()

	bucketAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	bucketSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	bucketEndpoint := "s3.dualstack.us-east-1.amazonaws.com"
	bucketName := "dpb587-website-us-east-1"
	bucketPrefix := "asset/optimized/"
	publicDir := os.Args[1]
	publicURL := ""

	imaginaryOpts := []string{
		"-path-prefix", "/d0c48e72-6c4f-4c24-ba62-7ebefd4a51da",
		"-concurrency", "25",
		"-http-cache-ttl", "31536000",
		"-enable-url-source",
		"-allowed-origins", "http://127.0.0.1:1313,http://localhost:1313,https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/",
	}
	imaginaryEndpoint := "http://localhost:8088/d0c48e72-6c4f-4c24-ba62-7ebefd4a51da/"

	logger.Debug(fmt.Sprintf("starting imaginary service"))

	os.Unsetenv("PORT")
	cmd := exec.Command("/usr/local/bin/imaginary", imaginaryOpts...)
	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return errors.Wrap(err, "running imaginary")
	}

	defer cmd.Process.Kill()

	timeout := time.Now().Add(60 * time.Second)

	for {
		res, _ := http.DefaultClient.Get(fmt.Sprintf("%shealth", imaginaryEndpoint))
		if res != nil && res.StatusCode == 200 {
			break
		} else if time.Now().After(timeout) {
			return fmt.Errorf("exceeded timeout waiting for imaginary service")
		}

		time.Sleep(time.Second)
	}

	logger.Info(fmt.Sprintf("started imaginary service"))

	client, err := minio.New(bucketEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(bucketAccessKey, bucketSecretKey, ""),
		Secure: true,
	})
	if err != nil {
		return errors.Wrap(err, "creating s3 client")
	}

	objects := &ObjectStorage{
		logger:         logger.With(zap.String("component", "ObjectStorage")),
		client:         client,
		bucketEndpoint: fmt.Sprintf("https://%s/%s", bucketEndpoint, bucketName),
		bucketName:     bucketName,
		bucketPrefix:   bucketPrefix,
		publicURL:      publicURL,
	}

	optimizer := &Optimizer{
		logger:  logger.With(zap.String("component", "Optimizer")),
		storage: objects,
	}

	fileProcessor := &FileProcessor{
		logger:            logger.With(zap.String("component", "FileProcessor")),
		optimizer:         optimizer,
		imaginaryEndpoint: imaginaryEndpoint,
	}

	//

	err = objects.Reload()
	if err != nil {
		return errors.Wrap(err, "finding existing objects")
	}

	//

	logger.Debug("processing files")

	var filesProcessed int

	err = filepath.Walk(publicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "walking %s", path)
		} else if info.IsDir() {
			return nil
		} else if !strings.HasSuffix(path, ".html") {
			return nil
		}

		filesProcessed++

		err = fileProcessor.ProcessFile(path)
		if err != nil {
			return errors.Wrapf(err, "optimizing %s", path)
		}

		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "walking %s", publicDir)
	}

	logger.Info(fmt.Sprintf("processed files (processed: %d)", filesProcessed))

	return nil
}

type Optimizer struct {
	imaginaryEndpoint string
	logger            *zap.Logger
	storage           *ObjectStorage
}

func (o *Optimizer) Optimize(rawURL string) (string, error) {
	def, err := o.calculateOptimizedObject(rawURL)
	if err != nil {
		return "", errors.Wrap(err, "calculating optimized object")
	}

	ref := def.OptimizedObjectRef

	existing, found := o.storage.Get(ref)
	if found {
		return existing, nil
	}

	o.logger.Debug(fmt.Sprintf("creating optimization %s", ref.KeyV1()))

	// imaginary

	o.logger.Debug(fmt.Sprintf("getting image %s", rawURL))

	req, err := http.DefaultClient.Get(def.URL)
	if err != nil {
		return "", errors.Wrapf(err, "getting image %s", rawURL)
	} else if v := req.StatusCode; v != 200 {
		return "", fmt.Errorf("unexpected status code from %s: status code %d", rawURL, v)
	}

	fhImaginary, err := ioutil.TempFile("", "imgrewrite-imaginary-*")
	if err != nil {
		return "", errors.Wrap(err, "creating temp imaginary file")
	}

	defer fhImaginary.Close()

	_, err = io.Copy(fhImaginary, req.Body)
	if err != nil {
		return "", errors.Wrapf(err, "downloading %s", rawURL)
	}

	// guetzli

	o.logger.Debug(fmt.Sprintf("compressing image %s", rawURL))

	fhGuetzli, err := ioutil.TempFile("", "imgrewrite-guetzli-*")
	if err != nil {
		return "", errors.Wrap(err, "creating temp guetzli file")
	}

	defer fhGuetzli.Close()

	guetzli := exec.Command("guetzli", "--quality", "90", fhImaginary.Name(), fhGuetzli.Name())
	err = guetzli.Run()
	if err != nil {
		return "", errors.Wrap(err, "running guetzli")
	}

	// upload

	err = o.storage.Put(*def, fhGuetzli)
	if err != nil {
		return "", errors.Wrap(err, "putting")
	}

	o.logger.Info(fmt.Sprintf("created optimization %s", ref.KeyV1()))

	return rawURL, nil
}

func (o *Optimizer) calculateOptimizedObject(rawURL string) (*OptimizedObjectDef, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.Wrap(err, "parsing URL")
	}

	var opOrigin string
	pathSplit := strings.Split(parsedURL.Path, "/")
	var opName = pathSplit[len(pathSplit)-1]
	var opArgs []string
	var opType string

	for key, values := range parsedURL.Query() {
		for _, value := range values {
			opArgs = append(opArgs, fmt.Sprintf("%s=%s", url.QueryEscape(key), url.QueryEscape(value)))
		}

		if key == "url" {
			if v := len(values); v != 1 {
				return nil, fmt.Errorf("expected a single value for url argument: found %d values", v)
			}

			opOrigin = values[0]
		} else if key == "type" {
			if v := len(values); v != 1 {
				return nil, fmt.Errorf("expected a single value for type argument: found %d values", v)
			}

			opType = values[0]
		}
	}

	if len(opOrigin) == 0 {
		return nil, errors.New("expected a single value for url argument: value missing")
	}

	if len(opType) == 0 {
		opType = "jpeg"
		opArgs = append(opArgs, fmt.Sprintf("type=%s", opType))
	}

	sort.Strings(opArgs)

	opFull := fmt.Sprintf("%s?%s", opName, strings.Join(opArgs, "&"))

	urlFull, err := parsedURL.Parse(opFull)
	if err != nil {
		return nil, errors.Wrap(err, "parsing optimization url")
	}

	fileName := path.Base(opOrigin)

	if strings.HasSuffix(fileName, fmt.Sprintf(".%s", opType)) {
		// nop
	} else if opType == "jpeg" && strings.HasSuffix(fileName, ".jpg") {
		// nop
	} else {
		fileName = fmt.Sprintf("%s.%s", fileName, opType)
	}

	ref := NewOptimizedObjectRef(opOrigin, opFull, opType)

	def := OptimizedObjectDef{
		OptimizedObjectRef: ref,
		URL:                urlFull.String(),
		FileName:           fileName,
	}

	o.logger.Debug(opFull)

	return &def, nil
}

type OptimizedObjectDef struct {
	OptimizedObjectRef
	URL      string
	FileName string
}

type FileProcessor struct {
	logger            *zap.Logger
	optimizer         *Optimizer
	imaginaryEndpoint string
}

func (fp *FileProcessor) ProcessFile(path string) error {
	var reFileMatcher = regexp.MustCompile(fmt.Sprintf(`[">](%s([^"<]+))`, regexp.QuoteMeta(fp.imaginaryEndpoint)))

	fp.logger.Debug(fmt.Sprintf("processing %s", path))

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "reading %s", path)
	}

	var originalData = string(buf)
	optimizationMap := map[string]string{}

	for _, match := range reFileMatcher.FindAllStringSubmatch(originalData, -1) {
		matchedURL := match[1]
		unescapedURL := strings.ReplaceAll(html.UnescapeString(matchedURL), "\\u0026", "&")

		optimizedURL, err := fp.optimizer.Optimize(unescapedURL)
		if err != nil {
			return errors.Wrapf(err, "replacing %s", matchedURL)
		}

		optimizationMap[matchedURL] = optimizedURL
	}

	//

	var updatedData = string(buf)

	for matchURL, optimizedURL := range optimizationMap {
		updatedData = strings.ReplaceAll(updatedData, fmt.Sprintf(`"%s"`, matchURL), fmt.Sprintf(`"%s"`, optimizedURL))
		updatedData = strings.ReplaceAll(updatedData, fmt.Sprintf(`>%s<`, matchURL), fmt.Sprintf(`>%s<`, optimizedURL))
	}

	bufUpdated := []byte(updatedData)

	if bytes.Compare(buf, bufUpdated) == 0 {
		fp.logger.Debug(fmt.Sprintf("processed %s (unchanged)", path))

		return nil
	}

	err = ioutil.WriteFile(path, bufUpdated, 0) // empty perm (must already exist)
	if err != nil {
		return errors.Wrapf(err, "writing %s", path)
	}

	fp.logger.Info(fmt.Sprintf("processed %s (optimizations: %d)", path, len(optimizationMap)))

	return nil
}

var reOptimizedObjectKey = regexp.MustCompile(`v1/([a-z0-9]{2})/([a-zA-Z0-9_\-]+)/([a-zA-Z0-9_\-]+)\.([a-z0-9]+)`)

type OptimizedObjectRef struct {
	Origin    string // base64
	Operation string // base64
	Ext       string
}

func (oor OptimizedObjectRef) KeyV1() string {
	originRaw, err := base64.RawURLEncoding.DecodeString(oor.Origin)
	if err != nil {
		panic(err) // should never happen
	}

	return fmt.Sprintf(
		"v1/%s/%s/%s.%s",
		fmt.Sprintf("%x", originRaw)[0:2],
		oor.Origin,
		oor.Operation,
		oor.Ext,
	)
}

func NewOptimizedObjectRef(origin, operation, ext string) OptimizedObjectRef {
	originHash := sha1.New()
	originHash.Write([]byte(origin))

	operationHash := sha1.New()
	operationHash.Write([]byte(operation))

	return OptimizedObjectRef{
		Origin:    base64.RawURLEncoding.EncodeToString(originHash.Sum(nil)),
		Operation: base64.RawURLEncoding.EncodeToString(operationHash.Sum(nil)),
		Ext:       ext,
	}
}

type OptimizedObjectURLs map[OptimizedObjectRef]string

type ObjectStorage struct {
	logger         *zap.Logger
	client         *minio.Client
	bucketEndpoint string
	bucketName     string
	bucketPrefix   string
	publicURL      string
	existing       OptimizedObjectURLs
}

func (oc *ObjectStorage) Reload() error {
	ctx := context.Background()

	oc.logger.Debug("reloading objects")

	objects := oc.client.ListObjects(
		ctx,
		oc.bucketName,
		minio.ListObjectsOptions{
			Prefix:    oc.bucketPrefix,
			Recursive: true,
		},
	)

	existing := OptimizedObjectURLs{}

	for object := range objects {
		if object.Err != nil {
			return errors.Wrap(object.Err, "listing objects")
		}

		match := reOptimizedObjectKey.FindStringSubmatch(strings.TrimPrefix(object.Key, oc.bucketPrefix))
		if len(match) == 0 {
			return fmt.Errorf("unable to match: %s", object.Key)
		}

		ref := OptimizedObjectRef{
			Origin:    match[2],
			Operation: match[3],
			Ext:       match[4],
		}

		existing[ref] = fmt.Sprintf("%s/%s", oc.publicURL, object.Key)
	}

	oc.existing = existing

	oc.logger.Info(fmt.Sprintf("reloaded optimizations (found: %d)", len(oc.existing)))

	return nil
}

func (oc *ObjectStorage) Get(ref OptimizedObjectRef) (string, bool) {
	v, found := oc.existing[ref]

	return v, found
}

func (oc *ObjectStorage) Put(def OptimizedObjectDef, file io.Reader) error {
	ctx := context.Background()

	objectKey := fmt.Sprintf("%s%s", oc.bucketPrefix, def.KeyV1())

	_, err := oc.client.PutObject(
		ctx,
		oc.bucketName,
		objectKey,
		file,
		-1,
		minio.PutObjectOptions{
			ContentDisposition: fmt.Sprintf("inline; filename=%q", def.FileName),
			CacheControl:       "public, max-age=2592000",
			ContentType:        fmt.Sprintf("image/%s", strings.TrimPrefix(filepath.Ext(objectKey), ".")),
		},
	)
	if err != nil {
		return errors.Wrap(err, "putting object")
	}

	oc.existing[def.OptimizedObjectRef] = fmt.Sprintf("%s/%s", oc.publicURL, objectKey)

	return nil
}
