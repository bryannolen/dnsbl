// package spamhaus provides an mechanism to query the Spamhaus DNS RBL.
package spamhaus

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"slices"
)

const (
	spamhausSuffix = "zen.spamhaus.org"
	shDROP         = "127.0.0.9"
	shSBL          = "127.0.0.2"
	shCSS          = "127.0.0.3"
	shXBL          = "127.0.0.4"
)

// customLookupHost performs a DNS lookup for a hostname using a custom DNS resolver.
func customLookupHost(hostname, resolver string) ([]string, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", net.JoinHostPort(resolver, "53"))
		},
	}
	return r.LookupHost(context.Background(), hostname)
}

// reverseIP takes an IP as a string and returns the octet inverted order.
// e.g. 127.0.0.1 becomes 1.0.0.127
func reverseIP(ip string) (string, error) {
	nip, err := netip.ParseAddr(ip)
	if err != nil {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}
	aSlice := nip.AsSlice()
	slices.Reverse(aSlice) // In place operation.
	reverseIP, ok := netip.AddrFromSlice(aSlice)
	if !ok {
		return "", fmt.Errorf("invert error")
	}
	return reverseIP.String(), nil
}

// QueryIP takes an IP address to query and the ip address to use as the DNS resolver as strings, and returns a slice of results from SpamHaus RBL.
func QueryByIP(queryIP, resolver string) ([]string, error) {
	if queryIP == "" || resolver == "" {
		return nil, fmt.Errorf("queryIP or resolver cannot be empty")
	}
	addr, err := reverseIP(queryIP)
	if err != nil {
		return nil, err
	}
	addr += "." + spamhausSuffix

	ips, err := customLookupHost(addr, resolver)
	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			if dnsErr.IsNotFound {
				return []string{"NXDOMAIN"}, nil
			}
			return nil, dnsErr
		}
	}
	r := []string{}
	for _, ip := range ips {
		switch ip {
		case shDROP:
			r = append(r, "DROP")
		case shSBL:
			r = append(r, "SBL")
		case shCSS:
			r = append(r, "CSS")
		case shXBL:
			r = append(r, "XBL")
		default:
			r = append(r, "OTHER")
		}
	}
	return r, nil
}
