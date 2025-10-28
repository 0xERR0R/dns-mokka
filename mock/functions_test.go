package mock_test

import (
	"time"

	"github.com/0xERR0R/dns-mokka/mock"
	"github.com/mattn/anko/env"
	"github.com/mattn/anko/vm"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Functions", func() {
	Describe("Execution", func() {
		var e *env.Env
		BeforeEach(func() {
			e, _ = mock.CreateEnv()
		})

		When("NXDOMAIN() is executed", func() {
			It("should return nxdomain", func() {
				execute, err := vm.Execute(e, nil, "NXDOMAIN()")
				Expect(err).Should(Succeed())
				result := execute.(mock.Result)

				Expect(result.Err).Should(BeNil())
				Expect(result.RCode).Should(Equal(dns.RcodeNameError))
			})
		})

		When("NOERROR() is executed", func() {
			It("should return valid response", func() {
				execute, err := vm.Execute(e, nil, `NOERROR("A 1.2.3.4 123")`)
				Expect(err).Should(Succeed())
				result := execute.(mock.Result)

				Expect(result.Err).Should(BeNil())
				Expect(result.RCode).Should(Equal(dns.RcodeSuccess))
				Expect(result.RR).Should(HaveLen(1))
				Expect(result.RR[0].TTL).Should(BeNumerically("==", 123))
				Expect(result.RR[0].Address).Should(Equal("1.2.3.4"))
				Expect(result.RR[0].RType).Should(Equal("A"))
			})

			It("should return valid response with multiple records", func() {
				execute, err := vm.Execute(e, nil, `NOERROR("A 1.2.3.4 123", "AAAA ::1 321")`)
				Expect(err).Should(Succeed())
				result := execute.(mock.Result)

				Expect(result.Err).Should(BeNil())
				Expect(result.RCode).Should(Equal(dns.RcodeSuccess))
				Expect(result.RR).Should(HaveLen(2))
				Expect(result.RR[0].TTL).Should(BeNumerically("==", 123))
				Expect(result.RR[0].Address).Should(Equal("1.2.3.4"))
				Expect(result.RR[0].RType).Should(Equal("A"))
				Expect(result.RR[1].TTL).Should(BeNumerically("==", 321))
				Expect(result.RR[1].Address).Should(Equal("::1"))
				Expect(result.RR[1].RType).Should(Equal("AAAA"))
			})

			It("should return error on invalid record", func() {
				execute, err := vm.Execute(e, nil, `NOERROR("A 1.2.3.4")`)
				Expect(err).Should(Succeed())
				result := execute.(mock.Result)

				Expect(result.Err).Should(HaveOccurred())
				Expect(result.Err.Error()).Should(ContainSubstring("record should be in format"))
			})

			It("should return valid response with RRSIG record", func() {
				execute, err := vm.Execute(e, nil,
					`NOERROR("A 1.2.3.4 300", "RRSIG A 13 3 300 20991231235959 20230101000000 12345 invalid.dnssec.test. fakesignaturedata== 300")`)
				Expect(err).Should(Succeed())
				result := execute.(mock.Result)

				Expect(result.Err).Should(BeNil())
				Expect(result.RCode).Should(Equal(dns.RcodeSuccess))
				Expect(result.RR).Should(HaveLen(2))

				// First record is A record
				Expect(result.RR[0].TTL).Should(BeNumerically("==", 300))
				Expect(result.RR[0].Address).Should(Equal("1.2.3.4"))
				Expect(result.RR[0].RType).Should(Equal("A"))

				// Second record is RRSIG
				Expect(result.RR[1].TTL).Should(BeNumerically("==", 300))
				Expect(result.RR[1].RType).Should(Equal("RRSIG"))
				Expect(result.RR[1].Address).Should(ContainSubstring("A 13 3 300"))
			})
		})

		When("delay() is executed", func() {
			It("should delay the response", func() {
				start := time.Now()
				execute, err := vm.Execute(e, nil, `delay(NXDOMAIN(), "100ms")`)
				duration := time.Since(start)

				Expect(err).Should(Succeed())
				result := execute.(mock.Result)

				Expect(result.Err).Should(BeNil())
				Expect(result.RCode).Should(Equal(dns.RcodeNameError))
				Expect(duration).Should(BeNumerically(">=", 100*time.Millisecond))
			})

			It("should return error on wrong duration", func() {
				execute, err := vm.Execute(e, nil, `delay(NXDOMAIN(), "100qwertz")`)
				Expect(err).Should(Succeed())
				result := execute.(mock.Result)

				Expect(result.Err).Should(HaveOccurred())
				Expect(result.Err.Error()).Should(ContainSubstring("can't parse duration"))
			})
		})
	})
})
