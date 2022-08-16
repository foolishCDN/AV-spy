package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http/httptrace"
	"time"

	"github.com/fatih/color"
)

const httpTemplate = "                   %s\n" +
	"DNS Lookup:     %s %s\n" +
	"Connection:     %s %s\n" +
	"WriteRequest:   %s %s\n" +
	"ServerResponse: %s %s\n\n"

const httpsTemplate = "                   %s\n" +
	"DNS Lookup:     %s %s\n" +
	"Connection:     %s %s\n" +
	"TLSHandshake:   %s %s\n" +
	"WriteRequest:   %s %s\n" +
	"ServerResponse: %s %s\n\n"

type Trace struct {
	DNSLookup        time.Duration
	Connection       time.Duration
	TLSHandshake     time.Duration
	WriteRequest     time.Duration
	ServerProcessing time.Duration
	contentTransfer  time.Duration

	GotConn time.Duration

	NameLookup    time.Duration
	Connect       time.Duration
	PreTransfer   time.Duration
	WroteRequest  time.Duration
	StartTransfer time.Duration

	AvailableIPs []net.IPAddr
	RemoteAddr   net.Addr
	LocalAddr    net.Addr

	getConn      time.Time
	gotConn      time.Time
	dnsStart     time.Time
	connectStart time.Time
	tlsStart     time.Time
	tlsDone      time.Time
	wroteReq     time.Time
	gotFirstByte time.Time

	isTLS        bool
	isReused     bool
	wasConnIdle  bool
	connIdleTime time.Duration
}

// start returns the request start time
func (t *Trace) start() time.Time {
	if !t.dnsStart.IsZero() {
		return t.dnsStart
	}
	return t.getConn
}

// end returns the get conn time
func (t *Trace) end() time.Time {
	if !t.tlsDone.IsZero() {
		return t.tlsDone
	}
	return t.gotConn
}

func (t *Trace) Pretty() string {
	var buf bytes.Buffer
	if len(t.AvailableIPs) > 0 {
		_, _ = fmt.Fprintf(&buf, "Avaliable IPs:     %v\n", t.AvailableIPs)
	}
	_, _ = fmt.Fprintf(&buf, "Connect to         %s\n", t.RemoteAddr)
	if t.isTLS {
		_, _ = fmt.Fprintf(&buf, httpsTemplate,
			fmtBlueStr("total"),
			fmtCyan(t.DNSLookup), fmtBlue(t.NameLookup),
			fmtCyan(t.Connection), fmtBlue(t.Connect),
			fmtCyan(t.TLSHandshake), fmtBlue(t.PreTransfer),
			fmtCyan(t.WriteRequest), fmtBlue(t.WroteRequest),
			fmtCyan(t.ServerProcessing), fmtBlue(t.StartTransfer),
		)
	} else {
		_, _ = fmt.Fprintf(&buf, httpTemplate,
			fmtBlueStr("total"),
			fmtCyan(t.DNSLookup), fmtBlue(t.NameLookup),
			fmtCyan(t.Connection), fmtBlue(t.Connect),
			fmtCyan(t.WriteRequest), fmtBlue(t.WroteRequest),
			fmtCyan(t.ServerProcessing), fmtBlue(t.StartTransfer),
		)
	}

	return buf.String()
}

func WithTrace(ctx context.Context, t *Trace) context.Context {
	return httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			now := time.Now()
			t.getConn = now
		},
		GotConn: func(info httptrace.GotConnInfo) {
			now := time.Now()
			t.gotConn = now
			t.GotConn = now.Sub(t.getConn)
			if info.Reused {
				t.isReused = true
			}
			if info.WasIdle {
				t.wasConnIdle = true
				t.connIdleTime = info.IdleTime
			}
			t.RemoteAddr = info.Conn.RemoteAddr()
			t.LocalAddr = info.Conn.LocalAddr()
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			t.dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			t.AvailableIPs = info.Addrs

			now := time.Now()
			t.DNSLookup = now.Sub(t.dnsStart)
			t.NameLookup = now.Sub(t.start())
		},
		ConnectStart: func(network, addr string) {
			now := time.Now()
			t.connectStart = now
		},
		ConnectDone: func(_, _ string, _ error) {
			now := time.Now()
			t.Connection = now.Sub(t.connectStart)
			t.Connect = now.Sub(t.start())
		},
		TLSHandshakeStart: func() {
			t.isTLS = true
			t.tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			now := time.Now()
			t.tlsDone = now
			t.TLSHandshake = now.Sub(t.tlsStart)
			t.PreTransfer = now.Sub(t.start())
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			now := time.Now()
			t.wroteReq = now
			t.WriteRequest = now.Sub(t.end())
			t.WroteRequest = now.Sub(t.start())
		},
		GotFirstResponseByte: func() {
			now := time.Now()
			t.gotFirstByte = now

			t.ServerProcessing = now.Sub(t.wroteReq)
			t.StartTransfer = now.Sub(t.start())
		},
	})
}

func fmtCyan(d time.Duration) string {
	return color.CyanString("%7dms", d.Milliseconds())
}

func fmtBlue(d time.Duration) string {
	return color.BlueString("%7dms", d.Milliseconds())
}

func fmtBlueStr(str string) string {
	return color.BlueString("%16s", str)
}
