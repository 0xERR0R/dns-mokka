package mock_test

import (
	"github.com/0xERR0R/dns-mokka/mock"
	"github.com/mattn/anko/env"
	"github.com/mattn/anko/vm"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Functions", func() {
	Describe("Execution", func() {
		var env *env.Env
		BeforeEach(func() {
			env, _ = mock.CreateEnv()
		})

		When("NXDOMAIN() is executed", func() {
			It("should return nxdomain", func() {
				execute, err := vm.Execute(env, nil, "NXDOMAIN()")
				Expect(err).Should(Succeed())
				result := execute.(mock.Result)

				Expect(result.Err).Should(BeNil())
				Expect(result.RCode).Should(Equal(dns.RcodeNameError))
			})
		})

		When("NOERROR() is executed", func() {
			It("should return valid response", func() {
				execute, err := vm.Execute(env, nil, `NOERROR("A 1.2.3.4 123")`)
				Expect(err).Should(Succeed())
				result := execute.(mock.Result)

				Expect(result.Err).Should(BeNil())
				Expect(result.RCode).Should(Equal(dns.RcodeSuccess))
				Expect(result.RR).Should(HaveLen(1))
				Expect(result.RR[0].TTL).Should(BeNumerically("==", 123))
				Expect(result.RR[0].Address).Should(Equal("1.2.3.4"))
				Expect(result.RR[0].RType).Should(Equal("A"))
			})
		})
	})
})
