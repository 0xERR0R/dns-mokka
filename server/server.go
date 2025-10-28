package server

import (
	"fmt"
	"strings"

	"github.com/0xERR0R/dns-mokka/config"
	"github.com/0xERR0R/dns-mokka/mock"
	"github.com/mattn/anko/env"
	"github.com/mattn/anko/vm"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	dnsServers []*dns.Server
	cfg        *config.Config
	env        *env.Env
}

func NewServer(cfg *config.Config) (*Server, error) {
	dnsServers := []*dns.Server{
		createUDPServer(cfg.ListenAddress),
		createTCPServer(cfg.ListenAddress),
	}

	env, err := mock.CreateEnv()

	if err != nil {
		return nil, fmt.Errorf("can't create env: %w", err)
	}

	s := &Server{
		dnsServers: dnsServers,
		cfg:        cfg,
		env:        env,
	}

	for _, server := range s.dnsServers {
		handler := server.Handler.(*dns.ServeMux)
		handler.HandleFunc(".", s.OnRequest)
	}

	return s, nil
}

func createUDPServer(address string) *dns.Server {
	const maxUDPSize = 65535

	return &dns.Server{
		Addr:    address,
		Net:     "udp",
		Handler: dns.NewServeMux(),
		NotifyStartedFunc: func() {
			log.Infof("UDP server is up and running on: '%s'", address)
		},
		UDPSize: maxUDPSize,
	}
}

func createTCPServer(address string) *dns.Server {
	return &dns.Server{
		Addr:    address,
		Net:     "tcp",
		Handler: dns.NewServeMux(),
		NotifyStartedFunc: func() {
			log.Infof("TCP server is up and running on: '%s'", address)
		},
	}
}

// returns EDNS upd size or if not present, 512 for UDP and 64K for TCP
func getMaxResponseSize(network string, request *dns.Msg) int {
	edns := request.IsEdns0()
	if edns != nil && edns.UDPSize() > 0 {
		return int(edns.UDPSize())
	}

	if network == "tcp" {
		return dns.MaxMsgSize
	}

	return dns.MinMsgSize
}

func (s *Server) OnRequest(rw dns.ResponseWriter, request *dns.Msg) {
	rulesForType := s.cfg.Rules[dns.Type(request.Question[0].Qtype)]

	rCode := dns.RcodeNameError

	var answers []dns.RR

	if rulesForType != nil {
		answers, rCode = s.processRules(rulesForType, request.Question[0].Name)
	}

	response := new(dns.Msg)
	response.SetRcode(request, rCode)
	response.Answer = answers

	response.MsgHdr.RecursionAvailable = request.MsgHdr.RecursionDesired

	// truncate if necessary
	response.Truncate(getMaxResponseSize(rw.LocalAddr().Network(), request))

	// enable compression
	response.Compress = true

	if err := rw.WriteMsg(response); err != nil {
		log.Error("can't write response: ", err)
	}
}

func (s *Server) processRules(rulesForType []config.RegexRule, name string) (answers []dns.RR, rCode int) {
	matched := false
	for _, rr := range rulesForType {
		if rr.Regex.MatchString(strings.ToLower(name)) {
			matched = true
			res, err := vm.Execute(s.env, nil, rr.Rule)
			if err != nil {
				log.Fatalf("can't execute rule '%s': %v", rr.Rule, err)
			}

			result := res.(mock.Result)

			rCode = result.RCode

			for _, rr := range result.RR {
				answer, err := dns.NewRR(fmt.Sprintf("%s %d %s %s %s",
					name, rr.TTL, "IN", rr.RType, rr.Address))
				if err != nil {
					log.Fatal("can't create answer", err)
				}

				answers = append(answers, answer)
			}

			break
		}
	}

	if !matched {
		rCode = dns.RcodeNameError
	}

	return
}

// Start starts the server
func (s *Server) Start() {
	log.Info("Starting server")

	for _, srv := range s.dnsServers {
		srv := srv

		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.Fatalf("start %s listener failed: %v", srv.Net, err)
			}
		}()
	}
}

// Stop stops the server
func (s *Server) Stop() {
	log.Info("Stopping server")

	for _, server := range s.dnsServers {
		if err := server.Shutdown(); err != nil {
			log.Fatalf("stop %s listener failed: %v", server.Net, err)
		}
	}
}
