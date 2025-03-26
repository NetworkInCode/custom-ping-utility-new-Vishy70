package helpers

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// AddrOptions are flags received from the command line
// related to using Internet Protocol version 4 or 6, to be used by pinger
// pinger will *exit* if both are set true.
// pinger will use IPv4 if none are set
type AddrOptions struct {
	V4 bool // use IPv4
	V6 bool // use IPv6
}

// UnMarshalledAddr is the type returned after preprocessing the user - given
// hostname or address.
type UnMarshalledAddr struct {
	Addr   string // address for pinger to use
	IsIPv6 bool   // protocol used: default is IPv4!
}

// UnMarshalledAddr setter function.
func (setAddr *UnMarshalledAddr) set(addr string, isIPv6 bool) {
	setAddr.Addr = addr
	setAddr.IsIPv6 = isIPv6
}

// Validate an IPv6 address
//
// valIPv6 validates a colon - delimited string to be IPv6 or not.
// It's a simple wrapper, using package net's ParseIP() and To16() methods.
func valIPv6(addr string) (bool, error) {
	ip := net.ParseIP(addr)

	if ip == nil {
		return false, fmt.Errorf("%v is not a valid IP address", addr)
	}

	if ip.To16() != nil {
		return true, nil
	}

	return false, nil
}

// Validate an IPv4 address
//
// valIPv4 validates a dotted - decimal string to be IPv4 or not.
// It's a simple wrapper, using package net's ParseIP() and To4() methods.
func valIPv4(addr string) (bool, error) {
	ip := net.ParseIP(addr)

	if ip == nil {
		return false, fmt.Errorf("%v is not a valid IP address", addr)
	}

	if ip.To4() != nil {
		return true, nil
	}

	return false, nil
}

//TODO: reduce code duplication in above 2 functions

// Check if hostname follows RFC 1123-ish format
//
// validateHostname sees if a given *host* string is a plausible domain name.
// Uses simple regex mapping; however, the hostname need not be fully qualified.
func validateHostname(host string) bool {
	const hostnameRegexRaw = `^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([a-zA-Z]{1,})$`

	host = strings.TrimSpace(host)
	re := regexp.MustCompile(hostnameRegexRaw)

	// Ensure no consecutive dots
	if strings.Contains(host, "..") {
		return false
	}

	// Compare the regex and the given hostname... does not check if hostname exists!
	return re.MatchString(host)

}

// Resolve a hostname using the local DNS resolver
//
// HostToAddr utilizes package net's LookupHost() method, to
// to determine an IP address for the given *host*.
// *options* specify whether to choose an IPv4 or IPv6 address.
// The default is Ipv4, different from normal ping.
func HostToAddr(host string, options AddrOptions) (UnMarshalledAddr, error) {
	var addr UnMarshalledAddr

	//TODO: Convert to error? Currently exiting...
	if options.V4 && options.V6 {
		addr.set("", false)
		return addr, fmt.Errorf("only one -4 or -6 option may be specified")
	}

	listOfIPs, err := net.LookupHost(host)

	if err != nil {
		return UnMarshalledAddr{}, err
	}

	if !options.V4 && !options.V6 && len(listOfIPs) != 0 {
		ip := listOfIPs[0]
		isIPv6, _ := valIPv6(ip)
		addr.set(ip, isIPv6)
		return addr, nil
	}

	for _, ip := range listOfIPs {
		if isIPv4, _ := valIPv4(ip); options.V4 && isIPv4 {
			addr.set(ip, false)
			return addr, nil
		}
		if isIPv6, _ := valIPv6(ip); options.V6 && isIPv6 {
			addr.set(ip, true)
			return addr, nil
		}
	}

	// Empty list caught here
	return UnMarshalledAddr{}, fmt.Errorf("could not resolve hostname %v. Please ensure a valid hostname is used", host)
}

// Resolve a *host* string to an appropriate Internet Protocol Address
//
// Determine if *host* string is a domain name or IP address
// If it's a domain, resolve to IPv4/6 using *options*
// Else, it's an IP address... ensure it is not malformed, determine v4/6 and return
func AddrResolution(host string, options AddrOptions) (UnMarshalledAddr, error) {
	var addr UnMarshalledAddr

	if validateHostname(host) {
		finalAddr, err := HostToAddr(host, options)

		return finalAddr, err
	}

	isIPv6, err := valIPv6(host)
	if err != nil {
		//addr.set("", false)
		return addr, err
	}

	if options.V6 && isIPv6 {
		addr.set(host, true)
		return addr, nil
	}

	//For now, default is resolve to IPv4
	isIPv4, err := valIPv4(host)
	if err != nil {
		//addr.set("", false)
		return addr, err
	}

	if isIPv4 {
		addr.set(host, false)
		return addr, nil
	}

	if options.V6 != isIPv6 {
		return addr, fmt.Errorf("option -6 specified does not match given IP: %v", host)
	} else if options.V4 != isIPv4 {
		return addr, fmt.Errorf("option -4 specified does not match given IP: %v", host)
	} else {
		return addr, fmt.Errorf("warning: an unexpected error occurred")
	}

}
