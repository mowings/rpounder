package main

import (
	"errors"
	"flag"
	"github.com/miekg/dns"
	"github.com/montanaflynn/stats"
	"log"
	"math/rand"
	"net"
	"regexp"
	"strings"
	"time"
)

type Msg struct {
	Times    []float64
	ErrCount int
}

func main() {
	var passes int
	var concurrency int
	var resolver string
	var name string
	log.Printf("Resolver benchmark starting...")
	flag.IntVar(&passes, "p", 100, "Number of passes")
	flag.IntVar(&concurrency, "c", 10, "Concurrent processes")
	flag.StringVar(&resolver, "r", "localhost:53", "resolver ip")
	flag.StringVar(&name, "n", "localhost", "Hostname(s) to look up")
	flag.Parse()
	passes_per_goproc := passes / concurrency
	if passes_per_goproc == 0 {
		log.Fatalf("concurrency needs to be greater than passes")
	}
	mod := passes % concurrency
	c := make(chan Msg)

	for i := 0; i < concurrency; i++ {
		go func() {
			num := passes_per_goproc
			if mod > 0 {
				num += 1
				mod--
			}
			pound(num, resolver, name, c)
		}()
	}

	i := 0
	times := make([]float64, 0, passes)
	errs := 0
	for msg := range c {
		errs += msg.ErrCount
		times = append(times, msg.Times...)
		i++
		if i >= concurrency {
			break
		}
	}
	// Print out percentiles
	for i = 10; i <= 100; i += 10 {
		rank := float64(i)
		if rank == 100 {
			rank = 99.9999999
		}
		v, _ := stats.Percentile(times, rank)
		log.Printf("%dth  => %s", i, time.Duration(v))
	}
	min, _ := stats.Min(times)
	max, _ := stats.Max(times)
	log.Printf("Fastest: %s  -- Slowest: %s", time.Duration(min), time.Duration(max))
	log.Printf("Total errors: %d", errs)
}

func pound(passes int, res, hostnames string, c chan Msg) {
	sprex := regexp.MustCompile("[\\s\\,]+")
	lookup_hosts := sprex.Split(hostnames, -1)
	msg := Msg{make([]float64, passes), 0}
	resolvers := sprex.Split(res, -1)
	resolver := NewResolver(resolvers)
	var err error
	hi := 0
	for i := 0; i < passes; i++ {
		start := time.Now()
		hostname := lookup_hosts[hi]
		hi++
		if hi >= len(lookup_hosts) {
			hi = 0
		}
		_, err = resolver.LookupHost(hostname)
		if err != nil {
			msg.ErrCount++
		}
		msg.Times[i] = float64(time.Since(start))
	}
	c <- msg
}

// DnsResolver represents a dns resolver
type DnsResolver struct {
	Servers []string
	r       *rand.Rand
}

// New initializes DnsResolver.
func NewResolver(servers []string) *DnsResolver {
	for i := range servers {
		if len(strings.Split(servers[i], ":")) < 2 {
			servers[i] = net.JoinHostPort(servers[i], "53")
		}
	}
	return &DnsResolver{servers, rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// LookupHost returns IP addresses of provied host.
func (r *DnsResolver) LookupHost(host string) ([]net.IP, error) {
	return r.lookupHost(host)
}

func (r *DnsResolver) lookupHost(host string) ([]net.IP, error) {
	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = true
	m1.Question = make([]dns.Question, 1)
	m1.Question[0] = dns.Question{dns.Fqdn(host), dns.TypeA, dns.ClassINET}
	in, err := dns.Exchange(m1, r.Servers[r.r.Intn(len(r.Servers))])

	result := []net.IP{}

	if err != nil {
		return result, err
	}

	if in != nil && in.Rcode != dns.RcodeSuccess {
		return result, errors.New(dns.RcodeToString[in.Rcode])
	}

	for _, record := range in.Answer {
		if t, ok := record.(*dns.A); ok {
			result = append(result, t.A)
		}
	}
	return result, err
}
