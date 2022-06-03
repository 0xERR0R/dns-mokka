package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/mattn/anko/env"

	"github.com/0xERR0R/dns-mokka/mock"
	"github.com/mattn/anko/vm"
	"github.com/sirupsen/logrus"

	"github.com/mattn/anko/parser"
	"github.com/miekg/dns"
)

const (
	prefix           = "MOKKA_"
	envLogLevel      = prefix + "LOG_LEVEL"
	envListenAddress = prefix + "LISTEN_ADDRESS"
	envRule          = prefix + "RULE_"
	tupleSize        = 2
)

type RegexRule struct {
	Regex *regexp.Regexp
	Rule  string
}

type Config struct {
	LogLevel      logrus.Level
	ListenAddress string
	Rules         map[dns.Type][]RegexRule
}

func ReadConfig() (*Config, error) {
	c := &Config{
		LogLevel:      logrus.InfoLevel,
		ListenAddress: ":53",
	}

	err := readEnv(c)

	return c, err
}

func readEnv(c *Config) error {
	var err error
	c.LogLevel, err = retrieveLogLevelFromEnv()

	if err != nil {
		return err
	}

	if addr, found := os.LookupEnv(envListenAddress); found {
		c.ListenAddress = addr
	}

	env, err := mock.CreateEnv()
	if err != nil {
		return fmt.Errorf("can't create env: %w", err)
	}

	c.Rules, err = retrieveRules(env)

	return err
}

func retrieveLogLevelFromEnv() (level logrus.Level, err error) {
	if l, found := os.LookupEnv(envLogLevel); found {
		level, err = logrus.ParseLevel(l)

		if err != nil {
			return logrus.FatalLevel, fmt.Errorf("unknown log level: %w", err)
		}

		return level, err
	}

	return logrus.InfoLevel, nil
}

func retrieveRules(env *env.Env) (map[dns.Type][]RegexRule, error) {
	rules := make(map[dns.Type][]RegexRule, 0)

	ruleNames := retrieveRulesFromEnv()

	sort.Strings(ruleNames)

	for _, r := range ruleNames {
		// example: A google.com/NOERROR(1.2.3.4)
		rule := os.Getenv(r)
		typeAddressRulePair := strings.SplitN(rule, "/", tupleSize)

		if len(typeAddressRulePair) != tupleSize {
			return nil, errors.New("rule should contain '/'")
		}

		typeAddressPair := strings.SplitN(typeAddressRulePair[0], " ", tupleSize)

		rType, found := dns.StringToType[typeAddressPair[0]]
		if !found {
			return nil, fmt.Errorf("unknown type '%s'", typeAddressPair[0])
		}

		regex, err := regexp.Compile(strings.ToLower(typeAddressPair[1]))
		if err != nil {
			return nil, fmt.Errorf("can't parse Regex '%s': %w", rule, err)
		}

		fn := typeAddressRulePair[1]
		_, err = parser.ParseSrc(fn)

		if err != nil {
			return nil, fmt.Errorf("can't parse Rule '%s': %w", rule, err)
		}

		res, err := vm.Execute(env, nil, fn)
		if err != nil {
			return nil, fmt.Errorf("can't execute function: %w", err)
		}

		result := res.(mock.Result)
		if result.Err != nil {
			return nil, fmt.Errorf("can't execute function: %w", result.Err)
		}

		if rules[dns.Type(rType)] == nil {
			rules[dns.Type(rType)] = make([]RegexRule, 0)
		}

		rules[dns.Type(rType)] = append(rules[dns.Type(rType)], RegexRule{
			Regex: regex,
			Rule:  fn,
		})
	}

	return rules, nil
}

func retrieveRulesFromEnv() (ruleNames []string) {
	for _, r := range os.Environ() {
		if strings.HasPrefix(r, envRule) {
			pair := strings.SplitN(r, "=", tupleSize)
			ruleNames = append(ruleNames, pair[0])
		}
	}

	sort.Strings(ruleNames)

	return
}
