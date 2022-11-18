package server

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/0xERR0R/dns-mokka/config"
	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

var (
	sut *Server
)

const address = ":55555"

var _ = BeforeSuite(func() {
	os.Setenv("MOKKA_LISTEN_ADDRESS", address)
	os.Setenv("MOKKA_RULE_2", `A g/NOERROR("A 1.2.3.5 1")`)
	os.Setenv("MOKKA_RULE_1", `A google/NOERROR("A 1.2.3.4 123")`)
	os.Setenv("MOKKA_RULE_3", `A delay.com/delay(NOERROR("A 1.1.1.1 100"), "100ms")`)
	os.Setenv("MOKKA_RULE_4", `A unknown/NXDOMAIN()`)

	cfg, err := config.ReadConfig()
	Expect(err).Should(Succeed())

	log.Infof("%+v", cfg)

	sut, err = NewServer(cfg)

	Expect(err).Should(Succeed())

	// start server
	go func() {
		sut.Start()
	}()

	// wait for server start
	Eventually(func() error {
		msg := new(dns.Msg)
		msg.SetQuestion(dns.Fqdn("google.de."), dns.TypeA)
		_, err := requestServer(msg)

		return err
	}, "5s").Should(BeNil())
	DeferCleanup(sut.Stop)
})

var _ = Describe("Server", func() {
	BeforeEach(func() {

	})
	When("DNS request is performed", func() {
		Context("first rule", func() {
			It("should return expected result", func() {
				msg := new(dns.Msg)
				msg.SetQuestion(dns.Fqdn("google.de."), dns.TypeA)

				resp, err := requestServer(msg)
				Expect(err).Should(Succeed())
				Expect(resp.Answer).Should(BeDNSRecord("google.de.", dns.TypeA, 123, "1.2.3.4"))
			})
		})
		Context("second rule", func() {
			It("should return expected result", func() {
				msg := new(dns.Msg)
				msg.SetQuestion(dns.Fqdn("domainwithg."), dns.TypeA)

				resp, err := requestServer(msg)
				start := time.Now()
				Expect(err).Should(Succeed())
				Expect(resp.Answer).Should(BeDNSRecord("domainwithg.", dns.TypeA, 1, "1.2.3.5"))
				Expect(time.Since(start)).Should(BeNumerically("<", 1*time.Millisecond))
			})

			It("should ignore case", func() {
				msg := new(dns.Msg)
				msg.SetQuestion(dns.Fqdn("domAINwithG"), dns.TypeA)

				resp, err := requestServer(msg)
				start := time.Now()
				Expect(err).Should(Succeed())
				Expect(resp.Answer).Should(BeDNSRecord("domAINwithG.", dns.TypeA, 1, "1.2.3.5"))
				Expect(time.Since(start)).Should(BeNumerically("<", 1*time.Millisecond))
			})
		})
		Context("third rule with delay", func() {
			It("should return expected result after delay", func() {
				msg := new(dns.Msg)
				msg.SetQuestion(dns.Fqdn("delay.com"), dns.TypeA)

				start := time.Now()
				resp, err := requestServer(msg)
				Expect(err).Should(Succeed())
				Expect(resp.Answer).Should(BeDNSRecord("delay.com.", dns.TypeA, 100, "1.1.1.1"))

				Expect(time.Since(start)).Should(BeNumerically(">", 100*time.Millisecond))
			})
		})

		Context("forth rule with NXDOMAIN", func() {
			It("should return NXDOMAIN", func() {
				msg := new(dns.Msg)
				msg.SetQuestion(dns.Fqdn("unknown.com"), dns.TypeA)
				resp, err := requestServer(msg)
				Expect(err).Should(Succeed())
				Expect(resp.Rcode).Should(Equal(dns.RcodeNameError))
			})
		})
	})
})

func requestServer(request *dns.Msg) (*dns.Msg, error) {
	conn, err := net.Dial("udp", address)
	if err != nil {
		return nil, fmt.Errorf("could not connect to server: %w", err)
	}
	defer conn.Close()

	msg, err := request.Pack()
	if err != nil {
		return nil, fmt.Errorf("can't pack request: %w", err)
	}

	_, err = conn.Write(msg)
	if err != nil {
		return nil, fmt.Errorf("can't send request to server: %w", err)
	}

	out := make([]byte, 1024)

	if _, err := conn.Read(out); err == nil {
		response := new(dns.Msg)
		err := response.Unpack(out)

		if err != nil {
			return nil, fmt.Errorf("can't unpack response: %w", err)
		}

		return response, nil
	}

	return nil, fmt.Errorf("could not read from connection: %w", err)
}
