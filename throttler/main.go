package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// Throttler struct holds most of our logic,
// the limiter is ulule's package
// the callback will be called if the user has not hit their limit
type Throttler struct {
	limiter  *limiter.Limiter
	callback func(http.ResponseWriter, *http.Request)
}

// ServeHTTP looks familiar? this will be ran on every request
// this is where the throttling logic will live
func (t Throttler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// see below
	ip, err := resolveIP(r)

	if err != nil {
		log.Printf("error obtaining ip address: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// fetch the users usage by their IP. If they have any
	limiterCtx, err := t.limiter.Get(r.Context(), ip)
	if err != nil {
		log.Printf("IPRateLimit - ipRateLimiter.Get - err: %v, %s on %s", err, ip, r.URL)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// set some headers to inform the user of their current usage
	h := w.Header()
	h.Set("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
	h.Set("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
	h.Set("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))

	// check whether the client has reached their limit,
	// if they have throw a 429 and a helpful message
	if limiterCtx.Reached {
		log.SetOutput(os.Stderr)
		log.Printf("Too Many Requests from %s on %s", ip, r.URL)
		w.WriteHeader(429)
		w.Write([]byte("{\"msg\":\"too many requests\"}"))
		return
	}

	// all should be good let us continue
	t.callback(w, r)
}

func main() {
	// set up a remote endpoint
	remote, err := url.Parse("http://my-app")
	if err != nil {
		panic(err)
	}
	// set up the proxy to connect to this endpoint
	proxy := httputil.NewSingleHostReverseProxy(remote)

	// init the throttler from above
	t := Throttler{
		limiter: limiter.New(
			memory.NewStore(),
			limiter.Rate{
				Period: 1 * time.Hour,
				Limit:  4, // TODO this could be added as an env_var and drastically increased
			}),
		callback: handler(proxy),
	}

	// begin to listen on port 8079
	err = http.ListenAndServe(":8079", t)
	if err != nil {
		panic(err)
	}
}


// handler is ran for users who have a surplus of requests left
func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.SetOutput(os.Stderr)
		log.Println(r.URL)
		w.Header().Set("X-Ben", "Rad")
		p.ServeHTTP(w, r)
	}
}

// resolveIP a helper function that finds the IP of the current user
// tries the IP, and then the for X-FORWARDED-FOR in case of proxy
// not fool proof, but good enough
func resolveIP(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		return "", err
	}

	if forIP := r.Header.Get("X-FORWARDED-FOR"); forIP != "" {
		ip = forIP
	}
	return ip, nil
}
