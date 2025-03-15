package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/foolishCDN/AV-spy/container/flv"
)

var (
	rootCmd = &cobra.Command{
		Use:           "simpleFlvParser ...[flags] <file path of http url> ...[flags]",
		Short:         "SimpleFlvParser is a simple tool to parse FLV stream",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	DefaultFormat           = "normal"
	DefaultTimeout          = 10
	DefaultShowPacketNumber = 0 // no limit

	seiFormatByte   = "byte"
	seiFormatString = "string"
	seiFormatHex    = "hex"
)

var (
	verbose bool

	// packet options
	showAll       bool
	showHeader    bool
	showMetaData  bool
	showPacket    bool
	showExtraData bool
	showSEI       bool
	seiFormat     string // default: hex
	num           int
	format        string

	// http options
	timeout    int
	header     []string
	follow302  bool
	insecure   bool
	serverName string

	// compute cache options
	diffThreshold int
)

func initFlags() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	// packet flags
	rootCmd.PersistentFlags().BoolVar(
		&showAll,
		"show",
		false,
		"will show all message",
	)
	rootCmd.PersistentFlags().BoolVar(
		&showHeader,
		"show_header",
		false,
		"will show flv file header",
	)
	rootCmd.PersistentFlags().BoolVar(
		&showMetaData,
		"show_metadata",
		false,
		"will show meta data",
	)
	rootCmd.PersistentFlags().BoolVar(
		&showPacket,
		"show_packets",
		false,
		"will show packets info",
	)
	rootCmd.PersistentFlags().BoolVar(
		&showExtraData,
		"show_extradata",
		false,
		"will show codec extradata(sequence header)",
	)
	rootCmd.PersistentFlags().BoolVar(
		&showSEI,
		"show_sei",
		false,
		"will show SEI(Supplemental Enhancement Information) and the packet info that carryed SEI",
	)
	rootCmd.PersistentFlags().StringVar(
		&seiFormat,
		"sei_format",
		seiFormatHex,
		"how to show SEI",
	)
	rootCmd.PersistentFlags().IntVarP(
		&num,
		"number",
		"n",
		DefaultShowPacketNumber,
		"show `n` packets (no limit if n<=0)",
	)
	rootCmd.PersistentFlags().StringVarP(
		&format,
		"format",
		"f",
		DefaultFormat,
		"output format",
	)
	// http flags
	rootCmd.PersistentFlags().IntVarP(
		&timeout,
		"timeout",
		"t",
		DefaultTimeout,
		"timeout for http request(seconds)",
	)
	rootCmd.PersistentFlags().StringSliceVarP(
		&header,
		"header",
		"H",
		[]string{},
		"http request header",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&follow302,
		"location",
		"L",
		false,
		"follow 302")
	rootCmd.PersistentFlags().BoolVarP(
		&insecure,
		"insecure_tls",
		"k",
		false,
		"insecure TLS connection")
	rootCmd.PersistentFlags().StringVar(
		&serverName,
		"server_name",
		"",
		"server name for TLS handshake",
	)
	rootCmd.PersistentFlags().IntVar(
		&diffThreshold,
		"diff_threshold",
		5,
		"when the diff between the real fps(using time) and the fps(using timestamp) is less than this threshold(percent), it is considered that all cache have been received",
	)
}

func main() {
	initFlags()
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if !(showPacket || showHeader || showExtraData || showMetaData || showAll || showSEI) {
			cmd.Usage()
			return errors.New("please set one or more flags to show")
		}
		if showAll {
			showPacket = true
			showHeader = true
			showExtraData = true
			showMetaData = true
			showSEI = true
		}
		if len(args) < 1 {
			cmd.Usage()
			return errors.New("please specify a file path of http url")
		}
		path := args[0]
		r, err := parseFilePathOrURL(path)
		if err != nil {
			return err
		}
		defer func() {
			_ = r.Close()
		}()

		p, err := NewFlvParser(format)
		if err != nil {
			return err
		}
		p.videoCounter.DiffThreshold = diffThreshold
		p.audioCounter.DiffThreshold = diffThreshold

		demuxer := new(flv.Demuxer)
		header, err := demuxer.ReadHeader(r)
		if err != nil {
			return err
		}
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			p.Summary()
			os.Exit(1)
		}()
		p.OnHeader(header)
		count := 0
		defer func() {
			p.Summary()
		}()
		for {
			tag, err := demuxer.ReadTag(r)
			if err != nil {
				if err == io.EOF {
					return nil
				} else {
					return err
				}
			}
			count++
			if err := p.OnPacket(tag); err != nil {
				return err
			}
			if num > 0 && count > num {
				break
			}
		}
		return nil
	}
	if err := rootCmd.Execute(); err != nil {
		logrus.Errorf("simpleFlvParser: %v", err)
		os.Exit(1)
	}
}

func parseFilePathOrURL(path string) (io.ReadCloser, error) {
	if isValidURL(path) {
		return doRequest(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file err: %v", err)
	}
	return f, nil
}

func doRequest(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request err: %v", err)
	}
	req.Header.Set("User-Agent", "SimpleFlvParser")
	for _, h := range header {
		kv := strings.Split(h, ":")
		name := textproto.CanonicalMIMEHeaderKey(strings.TrimSpace(kv[0]))
		value := strings.TrimSpace(kv[1])
		if name == "Host" {
			req.Host = value
		}
		req.Header.Set(name, value)
	}
	dumpRequest(req)
	client := makeClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request err: %v", err)
	}
	dumpResponse(resp)
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		if isRedirect(resp.StatusCode) {
			return nil, fmt.Errorf("do request err: status code=%d, use -L to follow location: %s", resp.StatusCode, resp.Header.Get("Location"))
		}
		return nil, fmt.Errorf("do request err: status code=%d", resp.StatusCode)
	}
	return resp.Body, nil
}

func isValidURL(path string) bool {
	u, err := url.ParseRequestURI(path)
	if err != nil {
		return false
	}
	if u.Scheme == "" || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
func makeClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: time.Duration(timeout) * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
				ServerName:         serverName,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if follow302 {
				logrus.Debugf("redirect to %s", req.URL.String())
				dumpRequest(req)
				return nil
			}
			return http.ErrUseLastResponse
		},
	}
	return client
}

func isRedirect(statusCode int) bool {
	return statusCode == http.StatusMovedPermanently ||
		statusCode == http.StatusFound ||
		statusCode == http.StatusSeeOther ||
		statusCode == http.StatusTemporaryRedirect ||
		statusCode == http.StatusPermanentRedirect
}

func dumpRequest(req *http.Request) {
	reqDump, _ := httputil.DumpRequest(req, false)
	for _, line := range strings.Split(string(reqDump), "\r\n") {
		if len(line) == 0 {
			continue
		}
		logrus.Debug(line)
	}
}

func dumpResponse(resp *http.Response) {
	respDump, _ := httputil.DumpResponse(resp, false)
	for _, line := range strings.Split(string(respDump), "\r\n") {
		if len(line) == 0 {
			continue
		}
		logrus.Debug(line)
	}
}
