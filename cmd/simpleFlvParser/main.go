package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/foolishCDN/AV-spy/container/flv"
)

var (
	rootCmd = &cobra.Command{
		Use:   "simpleFlvParser ...[flags] <file path of http url>",
		Short: "SimpleFlvParser is a simple tool to parse FLV stream",
	}

	DefaultFormat           = "normal"
	DefaultTimeout          = 10
	DefaultShowPacketNumber = 20
)

var (
	showAll       bool
	showHeader    bool
	showMetaData  bool
	showPacket    bool
	showExtraData bool

	num     int
	timeout int
	format  string
)

func initFlags() {
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
	rootCmd.PersistentFlags().IntVarP(
		&num,
		"number",
		"n",
		DefaultShowPacketNumber,
		"show `n` packets",
	)
	rootCmd.PersistentFlags().IntVarP(
		&timeout,
		"timeout",
		"t",
		DefaultTimeout,
		"timeout for http request(seconds)",
	)
	rootCmd.PersistentFlags().StringVarP(
		&format,
		"format",
		"f",
		DefaultFormat,
		"output format",
	)
}

func main() {
	initFlags()

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if !(showPacket || showHeader || showExtraData || showMetaData || showAll) {
			return errors.New("please set one or more flags")
		}
		if len(args) < 1 {
			return errors.New("please specify a file path of http url")
		}
		path := args[0]
		r, err := parseFilePathOrURL(path)
		if err != nil {
			return err
		}
		defer r.Close()

		p, err := NewFlvParser(format)
		if err != nil {
			return err
		}

		demuxer := new(flv.Demuxer)
		header, err := demuxer.ReadHeader(r)
		if err != nil {
			return err
		}
		p.OnHeader(header)
		count := 0
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
		os.Exit(1)
	}
}

func parseFilePathOrURL(path string) (io.ReadCloser, error) {
	if isValidURL(path) {
		req, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("new request err: %v", err)
		}
		client := &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: time.Duration(timeout) * time.Second,
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("do request err: %v", err)
		}
		return resp.Body, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file err: %v", err)
	}
	return f, nil
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
