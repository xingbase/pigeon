package mail

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

const (
	ErrInvalidEmailFormat    = Error("invalid email format")
	ErrInvalidDomainFormat   = Error("invalid domain format")
	ErrInvalidEmptyOrTooLong = Error("invalid empty or too long")
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func IsValidEmail(email string) (bool, error) {
	// Basic email format check
	if email == "" || len(email) > 254 {
		return false, ErrInvalidEmptyOrTooLong
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false, ErrInvalidEmailFormat
	}

	domain := strings.TrimSpace(strings.ToLower(parts[1]))
	if !regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(domain) {
		return false, ErrInvalidDomainFormat
	}

	// Custom resolver with timeout
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(_ context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.Dial(network, address)
		},
	}

	// MX record lookup
	mxRecords, err := resolver.LookupMX(context.Background(), domain)
	if err != nil {
		return false, fmt.Errorf("failed to lookup MX records for domain %s: %v", domain, err)
	}

	if len(mxRecords) > 0 {
		return true, nil
	}

	return false, fmt.Errorf("no MX records found for domain %s", domain)
}

func IsValidDomain(domain string) (bool, error) {
	if domain == "" || len(domain) > 255 {
		return false, ErrInvalidEmptyOrTooLong
	}

	// Normalize domain (trim whitespace and convert to lowercase)
	domain = strings.TrimSpace(strings.ToLower(domain))

	// Validate domain format (basic check for valid characters and structure)
	if !regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(domain) {
		return false, ErrInvalidDomainFormat
	}

	// Create a custom resolver with a timeout to avoid long DNS delays
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(_ context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.Dial(network, address)
		},
	}

	// Retry logic for MX lookup
	maxRetries := 3
	var mxErr error
	var mxRecords []*net.MX
	for attempt := 1; attempt <= maxRetries; attempt++ {
		mxRecords, mxErr = resolver.LookupMX(context.Background(), domain)
		if mxErr == nil {
			if len(mxRecords) > 0 {
				break
			}
			return false, fmt.Errorf("no MX records found")
		}
		if dnsErr, ok := mxErr.(*net.DNSError); ok && dnsErr.Temporary() {
			time.Sleep(time.Duration(attempt*100) * time.Millisecond) // Exponential backoff
			continue
		}
		break
	}
	if mxErr != nil {
		return false, fmt.Errorf("failed to lookup MX records: %v", mxErr)
	}

	// Try connecting to SMTP port (25) for each MX record
	for _, mx := range mxRecords {
		conn, err := net.DialTimeout("tcp", mx.Host+":25", 5*time.Second)
		if err == nil {
			conn.Close()
			return true, nil
		}
	}

	return false, fmt.Errorf("no reachable SMTP server found")
}
