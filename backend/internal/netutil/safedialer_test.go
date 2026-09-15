package netutil

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestIsDisallowedIP(t *testing.T) {
	disallowed := []string{
		"127.0.0.1",
		"127.0.1.1",
		"10.0.0.1",
		"10.254.0.1",
		"172.16.0.1",
		"172.31.255.255",
		"192.168.1.1",
		"192.168.254.254",
		"169.254.169.254", // AWS/GCP metadata
		"169.254.1.1",
		"100.64.0.1",   // CGNAT
		"100.127.0.1",  // CGNAT
		"192.0.2.1",    // TEST-NET-1
		"198.51.100.1", // TEST-NET-2
		"203.0.113.1",  // TEST-NET-3
		"0.0.0.0",
		"224.0.0.1",
		"240.0.0.1",
		"::1",
		"::",
		"fe80::1",
		"fc00::1",
		"fd00:ec2::254",
		"ff02::1",
		"::ffff:127.0.0.1",
		"::ffff:10.0.0.1",
		"::ffff:192.168.1.1",
	}

	for _, ipStr := range disallowed {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			t.Fatalf("failed to parse IP %q", ipStr)
		}
		if !IsDisallowedIP(ip) {
			t.Errorf("expected IsDisallowedIP(%q) to be true, got false", ipStr)
		}
	}

	allowed := []string{
		"8.8.8.8",
		"1.1.1.1",
		"93.184.216.34",
		"142.250.190.46",
		"2606:4700:4700::1111",
		"2001:4860:4860::8888",
	}

	for _, ipStr := range allowed {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			t.Fatalf("failed to parse IP %q", ipStr)
		}
		if IsDisallowedIP(ip) {
			t.Errorf("expected IsDisallowedIP(%q) to be false, got true", ipStr)
		}
	}
}

func TestValidateURL(t *testing.T) {
	invalidURLs := []string{
		"",
		"ftp://example.com/feed.xml",
		"file:///etc/passwd",
		"gopher://example.com",
		"http://",
		"http://user:pass@example.com/feed.xml",
		"http://127.0.0.1/feed.xml",
		"http://10.0.0.1/feed.xml",
		"http://192.168.1.1/feed.xml",
		"http://169.254.169.254/latest/meta-data/",
		"http://localhost:8080/feed",
		"http://app.localhost/feed",
		"http://server.local/feed",
		"http://service.internal/feed",
		"http://[::1]/feed",
		"http://[fe80::1]/feed",
	}

	for _, raw := range invalidURLs {
		if err := ValidateURL(raw); err == nil {
			t.Errorf("expected ValidateURL(%q) to fail, but it succeeded", raw)
		}
	}

	validURLs := []string{
		"https://techcrunch.com/feed/",
		"http://news.ycombinator.com/rss",
		"https://blog.google/technology/ai/rss/",
		"https://feeds.arstechnica.com/arstechnica/index",
	}

	for _, raw := range validURLs {
		if err := ValidateURL(raw); err != nil {
			t.Errorf("expected ValidateURL(%q) to succeed, but got: %v", raw, err)
		}
	}
}

func TestSafeDialContextBlocksDisallowed(t *testing.T) {
	dialFunc := SafeDialContext(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Direct loopback dial attempt should fail with prohibited error
	_, err := dialFunc(ctx, "tcp", "127.0.0.1:80")
	if err == nil {
		t.Fatal("expected dial to 127.0.0.1:80 to fail, but got nil")
	}

	// Direct metadata dial attempt should fail
	_, err = dialFunc(ctx, "tcp", "169.254.169.254:80")
	if err == nil {
		t.Fatal("expected dial to 169.254.169.254:80 to fail, but got nil")
	}
}
