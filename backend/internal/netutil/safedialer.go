// Package netutil provides network utilities with SSRF protection and safe
// HTTP transports for outbound requests.
package netutil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var disallowedCIDRs []*net.IPNet

func init() {
	cidrs := []string{
		"0.0.0.0/8",          // Current network (RFC 791)
		"10.0.0.0/8",         // Private network (RFC 1918)
		"100.64.0.0/10",      // Carrier-grade NAT (RFC 6598)
		"127.0.0.0/8",        // Loopback (RFC 1122)
		"169.254.0.0/16",     // Link-Local / Cloud Metadata (RFC 3927)
		"172.16.0.0/12",      // Private network (RFC 1918)
		"192.0.0.0/24",       // IETF Protocol Assignments (RFC 6890)
		"192.0.2.0/24",       // Documentation TEST-NET-1 (RFC 5737)
		"192.168.0.0/16",     // Private network (RFC 1918)
		"198.18.0.0/15",      // Benchmarking (RFC 2544)
		"198.51.100.0/24",    // Documentation TEST-NET-2 (RFC 5737)
		"203.0.113.0/24",     // Documentation TEST-NET-3 (RFC 5737)
		"224.0.0.0/4",        // Multicast (RFC 5771)
		"240.0.0.0/4",        // Reserved (RFC 1112)
		"255.255.255.255/32", // Limited Broadcast (RFC 919)
		"::/128",             // Unspecified IPv6
		"::1/128",            // Loopback IPv6
		"100::/64",           // Discard-Only
		"2001:db8::/32",      // Documentation IPv6
		"fc00::/7",           // Unique Local Address (ULA) IPv6
		"fe80::/10",          // Link-Local IPv6
		"ff00::/8",           // Multicast IPv6
	}
	for _, c := range cidrs {
		_, netBlock, err := net.ParseCIDR(c)
		if err == nil {
			disallowedCIDRs = append(disallowedCIDRs, netBlock)
		}
	}
}

// IsDisallowedIP reports whether an IP address is private, loopback, link-local,
// multicast, cloud metadata, or reserved for non-public routing.
func IsDisallowedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	// Unmap IPv4-in-IPv6 if applicable (e.g. ::ffff:127.0.0.1)
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}
	for _, block := range disallowedCIDRs {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// ValidateURL checks that rawURL has a safe format (http/https), does not
// contain embedded user credentials, has a valid host, and does not point
// directly to private/internal IP literals or known local hostnames.
func ValidateURL(rawURL string) error {
	if len(rawURL) > 2048 {
		return errors.New("url exceeds maximum length of 2048 characters")
	}
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("malformed URL: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q (only http and https are allowed)", u.Scheme)
	}
	if u.User != nil {
		return errors.New("embedded credentials in URL are prohibited")
	}
	hostname := u.Hostname()
	if hostname == "" {
		return errors.New("URL hostname is missing")
	}
	lowerHost := strings.ToLower(hostname)
	if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost") ||
		strings.HasSuffix(lowerHost, ".local") || strings.HasSuffix(lowerHost, ".internal") ||
		strings.HasSuffix(lowerHost, ".lan") || strings.HasSuffix(lowerHost, ".home.arpa") {
		return fmt.Errorf("access to internal/local domain %q is prohibited", hostname)
	}
	if ip := net.ParseIP(hostname); ip != nil {
		if IsDisallowedIP(ip) {
			return fmt.Errorf("access to restricted IP address (%s) is prohibited", ip.String())
		}
	}
	return nil
}

// SafeDialContext creates a dial function that resolves hostnames and verifies
// each candidate IP against disallowed private/internal address ranges before
// dialing, preventing SSRF and DNS rebinding attacks.
func SafeDialContext(dialer *net.Dialer) func(ctx context.Context, network, addr string) (net.Conn, error) {
	if dialer == nil {
		dialer = &net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid address %q: %w", addr, err)
		}

		// If host is already an IP literal
		if ip := net.ParseIP(host); ip != nil {
			if IsDisallowedIP(ip) {
				return nil, fmt.Errorf("connection to restricted IP address (%s) is prohibited", ip.String())
			}
			return dialer.DialContext(ctx, network, addr)
		}

		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("resolve host %q: %w", host, err)
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("no IP addresses found for host %q", host)
		}

		var lastErr error
		for _, ipAddr := range ips {
			if IsDisallowedIP(ipAddr.IP) {
				return nil, fmt.Errorf("connection to restricted IP address (%s) is prohibited", ipAddr.IP.String())
			}
			target := net.JoinHostPort(ipAddr.IP.String(), port)
			conn, err := dialer.DialContext(ctx, network, target)
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		return nil, lastErr
	}
}

// SafeTransport returns an http.Transport configured with SafeDialContext.
func SafeTransport(timeout time.Duration) *http.Transport {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}
	return &http.Transport{
		DialContext:           SafeDialContext(dialer),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: timeout,
	}
}

// SafeHTTPClient creates an *http.Client backed by SafeTransport and equipped
// with redirect validation to prevent SSRF via HTTP 301/302 redirects.
func SafeHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: SafeTransport(timeout),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			if err := ValidateURL(req.URL.String()); err != nil {
				return fmt.Errorf("redirect blocked: %w", err)
			}
			return nil
		},
	}
}
