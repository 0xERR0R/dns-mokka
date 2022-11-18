package mock

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/mattn/anko/env"
)

type ResolveFn func(request *dns.Msg) *dns.Msg

type Result struct {
	RCode int
	RR    []Record
	Err   error
}

type Record struct {
	TTL     int
	RType   string
	Address string
}

func nxdomain() Result {
	return Result{
		RCode: dns.RcodeNameError,
	}
}

func noerror(in ...string) Result {
	var rr = make([]Record, len(in))

	for ix, i := range in {
		record, err := parseRecord(i)
		if err != nil {
			return Result{Err: err}
		}

		rr[ix] = record
	}

	return Result{
		RCode: dns.RcodeSuccess,
		RR:    rr,
	}
}

func delay(fn Result, duration ...string) Result {
	d := time.Second

	if len(duration) != 0 {
		t, err := time.ParseDuration(duration[0])
		if err != nil {
			return Result{Err: fmt.Errorf("can't parse duration :%w", err)}
		}

		d = t
	}

	time.Sleep(d)

	return fn
}

func parseRecord(in string) (Record, error) {
	parts := strings.Split(in, " ")

	const recordTokens = 3

	if len(parts) != recordTokens {
		return Record{}, fmt.Errorf("record should be in format 'TYPE ANSWER TTL', for example 'A 1.2.3.4 20'")
	}

	ttl, err := strconv.Atoi(parts[2])
	if err == nil {
		return Record{
			RType:   parts[0],
			Address: parts[1],
			TTL:     ttl,
		}, nil
	}

	return Record{}, fmt.Errorf("TTL can't be parsed: %w", err)
}

func CreateEnv() (*env.Env, error) {
	e := env.NewEnv()

	if err := e.Define("NXDOMAIN", nxdomain); err != nil {
		return nil, err
	}

	if err := e.Define("NOERROR", noerror); err != nil {
		return nil, err
	}

	if err := e.Define("delay", delay); err != nil {
		return nil, err
	}

	return e, nil
}
