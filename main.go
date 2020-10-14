package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/google/uuid"
)

var packageJSONLock = sync.Mutex{}

func main() {
	publicAddr := "wpt." + os.Getenv("PUBLIC_ADDR")
	publicAddrWWW := "a.wpt." + os.Getenv("PUBLIC_ADDR")
	publicAddrWWW1 := "b.wpt." + os.Getenv("PUBLIC_ADDR")
	publicAddrWWW2 := "c.wpt." + os.Getenv("PUBLIC_ADDR")

	if os.Getenv("DEV") != "" {
		publicAddr = "bs-local.com:" + os.Getenv("PORT")
		publicAddrWWW = "bs-local.com:" + os.Getenv("PORT")
		publicAddrWWW1 = "bs1-local.com:" + os.Getenv("PORT")
		publicAddrWWW2 = "bs2-local.com:" + os.Getenv("PORT")
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         ":" + os.Getenv("PORT"),
		Handler: wptHandler(
			publicAddr,
			publicAddrWWW,
			publicAddrWWW1,
			publicAddrWWW2,
		),
	}

	go func() {
		log.Println("http: Server listening")
		err := srv.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
}

func wptHandler(publicAddr string, publicAddrWWW string, publicAddrWWW1 string, publicAddrWWW2 string) http.Handler {
	wptURL, err := url.Parse("https://wpt.live/")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(wptURL)

	proxy.ModifyResponse = func(w *http.Response) error {
		return nil
	}

	proxy.Transport = &rewritingTransport{
		roundTripper:   http.DefaultTransport,
		publicAddr:     publicAddr,
		publicAddrWWW:  publicAddrWWW,
		publicAddrWWW1: publicAddrWWW1,
		publicAddrWWW2: publicAddrWWW2,
	}

	return proxy
}

type rewritingTransport struct {
	roundTripper   http.RoundTripper
	publicAddr     string
	publicAddrWWW  string
	publicAddrWWW1 string
	publicAddrWWW2 string
}

func (t *rewritingTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	ctx, cancel := context.WithTimeout(req.Context(), time.Second*60)
	defer cancel()

	if strings.HasPrefix(req.URL.Path, "/.") && !strings.HasPrefix(req.URL.Path, "/.well-known") {
		resp = &http.Response{}
		resp.StatusCode = http.StatusNotFound
		resp.Status = http.StatusText(http.StatusNotFound)
		resp.Body = ioutil.NopCloser(bytes.NewReader([]byte(fmt.Sprintf("%d %s", http.StatusNotFound, http.StatusText(http.StatusNotFound)))))
		return resp, nil
	}

	query := req.URL.Query()
	for k, vv := range query {
		for i, v := range vv {
			query[k][i] = t.rewriteStringReverse(v)
		}
	}

	req.URL.RawQuery = query.Encode()

	resp, err = t.roundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if resp.Header.Get("Location") != "" {
		resp.Header.Set("Location", t.rewriteString(resp.Header.Get("Location")))
	}

	setCookieValues := resp.Header.Values("Set-Cookie")
	resp.Header.Del("Set-Cookie")
	for _, v := range setCookieValues {
		resp.Header.Add("Set-Cookie", t.rewriteString(v))
	}

	b = t.rewriteBytes(b)

	// TODO : remove version param
	if resp.Header.Get("Content-Type") == "text/html" {
		re := regexp.MustCompile(`(?s)<\s*script[^>]*>([^<]+?)</script>`)

		submatchall := re.FindAllSubmatch(b, -1)
		for _, element := range submatchall {
			if len(element[1]) > 0 {
				b = bytes.Replace(b, element[1], t.transpileJS(ctx, element[1], filepath.Base(req.URL.Path)), -1)
			}
		}

		if bytes.Contains(b, []byte("<head>")) {
			b = bytes.Replace(b, []byte("<head>"), []byte("<head><script src=\"https://polyfill.io/v3/polyfill.min.js?features=all&version=3.89.4\"></script>"), 1)
		} else {
			b = bytes.Replace(b, []byte("<script"), []byte("<script src=\"https://polyfill.io/v3/polyfill.min.js?features=all&version=3.89.4\"></script><script"), 1)
		}

		b = append(b, []byte("<!--All contents pulled from https://wpt.live-->")...)
	}

	if resp.Header.Get("Content-Type") == "text/javascript; charset=utf-8" ||
		resp.Header.Get("Content-Type") == "text/javascript" {
		b = t.transpileJS(ctx, b, filepath.Base(req.URL.Path))
	}

	body := ioutil.NopCloser(bytes.NewReader(b))
	resp.Body = body
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))

	return resp, nil
}

func (t *rewritingTransport) rewriteBytes(b []byte) []byte {
	b = bytes.Replace(b, []byte("www.wpt.live:80"), []byte(t.publicAddrWWW), -1)
	b = bytes.Replace(b, []byte("www.wpt.live"), []byte(t.publicAddrWWW), -1)

	b = bytes.Replace(b, []byte("www1.wpt.live:80"), []byte(t.publicAddrWWW1), -1)
	b = bytes.Replace(b, []byte("www1.wpt.live"), []byte(t.publicAddrWWW1), -1)

	b = bytes.Replace(b, []byte("www2.wpt.live:80"), []byte(t.publicAddrWWW2), -1)
	b = bytes.Replace(b, []byte("www2.wpt.live"), []byte(t.publicAddrWWW2), -1)

	b = bytes.Replace(b, []byte("wpt.live:80"), []byte(t.publicAddr), -1)
	b = bytes.Replace(b, []byte("wpt.live"), []byte(t.publicAddr), -1)

	if os.Getenv("DEV") != "" {
		b = bytes.Replace(b, []byte("https://"), []byte("http://"), 1)
	}

	return b
}

func (t *rewritingTransport) rewriteString(s string) string {
	s = strings.Replace(s, "www.wpt.live:80", t.publicAddrWWW, -1)
	s = strings.Replace(s, "www.wpt.live", t.publicAddrWWW, -1)

	s = strings.Replace(s, "www1.wpt.live:80", t.publicAddrWWW1, -1)
	s = strings.Replace(s, "www1.wpt.live", t.publicAddrWWW1, -1)

	s = strings.Replace(s, "www2.wpt.live:80", t.publicAddrWWW2, -1)
	s = strings.Replace(s, "www2.wpt.live", t.publicAddrWWW2, -1)

	s = strings.Replace(s, ".wpt.live:80", "."+t.publicAddr, -1)
	s = strings.Replace(s, ".wpt.live", "."+t.publicAddr, -1)

	s = strings.Replace(s, "wpt.live:80", t.publicAddr, -1)
	s = strings.Replace(s, "wpt.live", t.publicAddr, -1)

	if os.Getenv("DEV") != "" {
		s = strings.Replace(s, "https://", "http://", 1)
	}

	return s
}

func (t *rewritingTransport) rewriteBytesReverse(b []byte) []byte {
	b = bytes.Replace(b, []byte("www."+t.publicAddrWWW), []byte("www.wpt.live"), -1)

	b = bytes.Replace(b, []byte("www1."+t.publicAddrWWW), []byte("www1.wpt.live"), -1)
	b = bytes.Replace(b, []byte("www1."+t.publicAddrWWW1), []byte("www1.wpt.live"), -1)

	b = bytes.Replace(b, []byte("www2."+t.publicAddrWWW), []byte("www2.wpt.live"), -1)
	b = bytes.Replace(b, []byte("www2."+t.publicAddrWWW2), []byte("www2.wpt.live"), -1)

	b = bytes.Replace(b, []byte(t.publicAddr), []byte("wpt.live"), -1)
	b = bytes.Replace(b, []byte(t.publicAddrWWW), []byte("www.wpt.live"), -1)
	b = bytes.Replace(b, []byte(t.publicAddrWWW1), []byte("www1.wpt.live"), -1)
	b = bytes.Replace(b, []byte(t.publicAddrWWW2), []byte("www2.wpt.live"), -1)

	return b
}

func (t *rewritingTransport) rewriteStringReverse(s string) string {
	s = strings.Replace(s, "www."+t.publicAddrWWW, "www.wpt.live", -1)

	s = strings.Replace(s, "www1."+t.publicAddrWWW, "www1.wpt.live", -1)
	s = strings.Replace(s, "www1."+t.publicAddrWWW1, "www1.wpt.live", -1)

	s = strings.Replace(s, "www2."+t.publicAddrWWW, "www2.wpt.live", -1)
	s = strings.Replace(s, "www2."+t.publicAddrWWW2, "www2.wpt.live", -1)

	s = strings.Replace(s, t.publicAddr, "wpt.live", -1)
	s = strings.Replace(s, t.publicAddrWWW, "www.wpt.live", -1)
	s = strings.Replace(s, t.publicAddrWWW1, "www1.wpt.live", -1)
	s = strings.Replace(s, t.publicAddrWWW2, "www2.wpt.live", -1)

	return s
}

func (t *rewritingTransport) transpileJS(ctx context.Context, b []byte, fileName string) []byte {
	id := uuid.New()

	jsInFileName := fmt.Sprintf("artifacts/%s.js", id)
	jsOutFileName := fmt.Sprintf("dist/%s.js", id)

	// Source
	{
		f, err := os.Create(jsInFileName)
		if err != nil {
			log.Println(err)
			return b
		}

		defer f.Close()

		n, err := f.Write(b)
		if err != nil {
			log.Println(err)
			return b
		}

		if n != len(b) {
			log.Println(io.ErrShortWrite)
			return b
		}

		err = f.Close()
		if err != nil {
			log.Println(err)
			return b
		}
	}

	// Webpack
	{
		cmd := exec.Command(
			"yarn",
			"-s",
			"webpack",
			"--entry",
			"./"+jsInFileName,
			"-o",
			"./"+jsOutFileName,
		)

		cmdOutput, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(err)
			if len(cmdOutput) > 0 {
				log.Println(string(cmdOutput))
			}
			return b
		}

		if len(cmdOutput) > 0 {
			log.Println(string(cmdOutput))
		}

		result, err := ioutil.ReadFile(filepath.Join(jsOutFileName, "main.js"))
		if err != nil {
			log.Println(err)
			return b
		}

		return append(result, []byte("\n/* transpiled */\n")...)
	}
}

func validateJSSource(code []byte) error {
	result := api.Transform(string(code), api.TransformOptions{
		Loader: api.LoaderJS,
		Target: api.ES5,
	})

	hasRealWarningsOrErrors := false
	if len(result.Errors) > 0 || len(result.Warnings) > 0 {
		for _, err := range result.Errors {
			log.Println("err", err)
			hasRealWarningsOrErrors = true
		}

		for _, warning := range result.Warnings {
			if strings.Contains(warning.Text, "Comparison with -0") {
				continue
			}

			log.Println("warning", warning)
			hasRealWarningsOrErrors = true
		}
	}

	if hasRealWarningsOrErrors {
		return errors.New("Error parsing source code")
	}

	return nil
}
