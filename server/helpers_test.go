package server

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/onsi/gomega/types"
)

func BeDNSRecord(domain string, dnsType uint16, ttl uint32, answer string) types.GomegaMatcher {
	return &dnsRecordMatcher{
		domain:  domain,
		dnsType: dnsType,
		TTL:     ttl,
		answer:  answer,
	}
}

type dnsRecordMatcher struct {
	domain  string
	dnsType uint16
	TTL     uint32
	answer  string
}

func (matcher *dnsRecordMatcher) matchSingle(rr dns.RR) (success bool, err error) {
	if (rr.Header().Name != matcher.domain) ||
		(rr.Header().Rrtype != matcher.dnsType) ||
		(matcher.TTL > 0 && (rr.Header().Ttl != matcher.TTL)) {
		return false, nil
	}

	switch v := rr.(type) {
	case *dns.A:
		return v.A.String() == matcher.answer, nil
	case *dns.AAAA:
		return v.AAAA.String() == matcher.answer, nil
	case *dns.PTR:
		return v.Ptr == matcher.answer, nil
	case *dns.MX:
		return v.Mx == matcher.answer, nil
	}

	return false, nil
}

// Match checks the DNS record
func (matcher *dnsRecordMatcher) Match(actual interface{}) (success bool, err error) {
	switch i := actual.(type) {
	case dns.RR:
		return matcher.matchSingle(i)
	case []dns.RR:
		return matcher.matchSingle(i[0])
	default:
		return false, fmt.Errorf("DNSRecord matcher expects an dns.RR or []dns.RR")
	}
}

// FailureMessage generates a failure messge
func (matcher *dnsRecordMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\n to contain\n\t domain '%s', ttl '%d', type '%s', answer '%s'",
		actual, matcher.domain, matcher.TTL, dns.TypeToString[matcher.dnsType], matcher.answer)
}

// NegatedFailureMessage creates negated message
func (matcher *dnsRecordMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\n not to contain\n\t domain '%s', ttl '%d', type '%s', answer '%s'",
		actual, matcher.domain, matcher.TTL, dns.TypeToString[matcher.dnsType], matcher.answer)
}
