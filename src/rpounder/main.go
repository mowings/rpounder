package main

import (
	"errors"
	"flag"
	"github.com/miekg/dns"
	"github.com/montanaflynn/stats"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type Msg struct {
	Times    []float64
	ErrCount int
}

const VERSION = "1.0.0"

func main() {
	var passes int
	var concurrency int
	var resolver string
	var name string
	log.Printf("rpounder resolver benchmark %s starting", VERSION)
	flag.IntVar(&passes, "p", 100, "Number of passes")
	flag.IntVar(&concurrency, "c", 10, "Concurrent processes")
	flag.StringVar(&resolver, "r", "localhost:53", "resolver(s) ip/port. Separate multiple resolvers with spaces or commas")
	flag.StringVar(&name, "n", "example.com", "Hostname(s) to look up. Separate multiple hostnames with spaces or commas")
	flag.Parse()
	passes_per_goproc := passes / concurrency
	if passes_per_goproc == 0 {
		log.Fatalf("concurrency needs to be greater than passes")
	}
	mod := passes % concurrency
	c := make(chan Msg)
	q := make(chan struct{})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("=== INTERRUPT: sig = %v", sig)
		close(q)
	}()

	for i := 0; i < concurrency; i++ {
		go func() {
			num := passes_per_goproc
			if mod > 0 {
				num += 1
				mod--
			}
			pound(num, resolver, name, c, q)
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
	log.Printf("Percentage of the requests served within a certain time:")
	pctls := []int{50, 66, 75, 80, 90, 95, 99}
	for _, rank := range pctls {
		v, _ := stats.Percentile(times, float64(rank))
		log.Printf("%d%%  => %s", rank, time.Duration(v))
	}
	min, _ := stats.Min(times)
	max, _ := stats.Max(times)
	log.Printf("Fastest: %s  -- Slowest: %s", time.Duration(min), time.Duration(max))
	log.Printf("Total passes: %d. Total errors: %d", len(times), errs)
}

func pound(passes int, res, hostnames string, c chan Msg, q chan struct{}) {
	sprex := regexp.MustCompile("[\\s\\,]+")
	lookup_hosts := sprex.Split(hostnames, -1)
	msg := Msg{make([]float64, 0, passes), 0}
	defer func() {
		c <- msg
	}()
	resolvers := sprex.Split(res, -1)
	resolver := NewResolver(resolvers)
	var err error
	hi := 0
	for i := 0; i < passes; i++ {
		select {
		case <-q:
			return
		default:
		}
		hostname := lookup_hosts[hi]
		hi++
		if hi >= len(lookup_hosts) {
			hi = 0
		}
		start := time.Now()
		_, err = resolver.LookupHost(hostname)
		end := float64(time.Since(start))
		if err != nil {
			msg.ErrCount++
		}
		msg.Times = append(msg.Times, end)
	}
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
