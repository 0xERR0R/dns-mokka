package config

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Config", func() {

	Describe("parse config", func() {
		When("config is valid", func() {
			BeforeEach(func() {
				os.Setenv(envLogLevel, "warn")
				os.Setenv(envRule+"2", `A test/delay(NOERROR("A 1.2.3.4 20"),"30ms")`)
				os.Setenv(envRule+"1", `AAAA ./NOERROR("A 1.2.3.4 20")`)
				DeferCleanup(os.Clearenv)
			})
			It("should create valid config", func() {
				cfg, err := ReadConfig()
				Expect(err).Should(Succeed())
				Expect(cfg.LogLevel).Should(Equal(logrus.WarnLevel))
				Expect(cfg.Rules).Should(HaveLen(2))
			})
		})

		When("loglevel is invalid", func() {
			BeforeEach(func() {
				os.Setenv(envLogLevel, "invalid")
				DeferCleanup(os.Clearenv)
			})
			It("should fail", func() {
				_, err := ReadConfig()
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("unknown log level"))
			})
		})

		When("function is unknown", func() {
			BeforeEach(func() {
				os.Setenv(envRule+"1", `AAAA ./unknown_fun("A 1.2.3.4 20")`)
				DeferCleanup(os.Clearenv)
			})
			It("should fail", func() {
				_, err := ReadConfig()
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("undefined symbol 'unknown_fun'"))
			})
		})

		When("syntax error", func() {
			BeforeEach(func() {
				os.Setenv(envRule+"1", `AAAA ./NOERROR("A 1.2.3.4 20"`)
				DeferCleanup(os.Clearenv)
			})
			It("should fail", func() {
				_, err := ReadConfig()
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("can't parse Rule 'AAAA ./NOERROR(\"A 1.2.3.4 20\"': syntax error"))
			})
		})

		When("wrong arguments record", func() {
			BeforeEach(func() {
				os.Setenv(envRule+"1", `AAAA ./NOERROR("A 1.2.3.4")`)
				DeferCleanup(os.Clearenv)
			})
			It("should fail", func() {
				_, err := ReadConfig()
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("record should be in format"))
			})
		})

		When("wrong TTL record", func() {
			BeforeEach(func() {
				os.Setenv(envRule+"1", `AAAA ./NOERROR("A 1.2.3.4 20s")`)
				DeferCleanup(os.Clearenv)
			})
			It("should fail", func() {
				_, err := ReadConfig()
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("TTL can't be parsed"))
			})
		})

		When("wrong arguments delay", func() {
			BeforeEach(func() {
				os.Setenv(envRule+"1", `AAAA ./delay(NOERROR("A 1.2.3.4 20"),"100wrongdelay")`)
				DeferCleanup(os.Clearenv)
			})
			It("should fail", func() {
				_, err := ReadConfig()
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("can't parse duration"))
			})
		})

		When("query type is unknown", func() {
			BeforeEach(func() {
				os.Setenv(envRule+"1", `Unknown ./NOERROR("A 1.2.3.4 20")`)
				DeferCleanup(os.Clearenv)
			})
			It("should fail", func() {
				_, err := ReadConfig()
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("unknown type 'Unknown'"))
			})
		})

		When("Regex is invalid", func() {
			BeforeEach(func() {
				os.Setenv(envRule+"1", `A .[/NOERROR("A 1.2.3.4 20")`)
				DeferCleanup(os.Clearenv)
			})
			It("should fail", func() {
				_, err := ReadConfig()
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("can't parse Regex 'A .[/NOERROR(\"A 1.2.3.4 20\")'"))
			})
		})

		When("Rule in wrong format", func() {
			BeforeEach(func() {
				os.Setenv(envRule+"1", `A NOERROR("A 1.2.3.4 20")`)
				DeferCleanup(os.Clearenv)
			})
			It("should fail", func() {
				_, err := ReadConfig()
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("rule should contain '/'"))
			})
		})
	})

})
