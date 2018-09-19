package main

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	err := serve(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

type requestParam struct {
	sid   string
	rid   string
	delay time.Duration
	score int
}

type server struct {
	l      sync.Mutex
	sid    string
	params map[string]*requestParam

	topIn  *requestParam
	topOut *requestParam
}

func (s *server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	d, out := s.randomDelay()
	p := &requestParam{
		sid:   q.Get("sid"),
		rid:   q.Get("rid"),
		delay: d,
		score: s.randomScore(out),
	}
	go s.put(p)
	time.Sleep(p.delay)
	rw.WriteHeader(http.StatusOK)
	rw.Write(([]byte)(strconv.Itoa(p.score)))
}

func (s *server) put(p *requestParam) {
	s.l.Lock()
	defer s.l.Unlock()
	if s.sid != p.sid {
		log.Printf("sid changed: %q", p.sid)
		s.sid = p.sid
		s.params = make(map[string]*requestParam)
		s.topIn = nil
		s.topOut = nil
	}
	s.params[p.rid] = p
	if p.delay < (100*time.Millisecond) {
		if s.topIn == nil || p.score > s.topIn.score {
			s.topIn = p
			log.Printf("topIn changed: rid=%q score=%d", p.rid, p.score)
		}
	} else {
		if s.topOut == nil || p.score > s.topOut.score {
			s.topOut = p
			log.Printf("topOut changed: rid=%q score=%d", p.rid, p.score)
		}
	}
}

func (s *server) randomDelay() (time.Duration, bool) {
	if rand.Float64() < 0.2 {
		return time.Duration(rand.Intn(100)+100) * time.Millisecond, true
	}
	return time.Duration(rand.Intn(100)) * time.Millisecond, false
}

func (s *server) randomScore(out bool) int {
	if out {
		return rand.Intn(100) + 1000
	}
	return rand.Intn(1000)
}

func serve(addr string) error {
	var h http.Handler
	h = &server{
		params: map[string]*requestParam{},
	}
	return http.ListenAndServe(addr, h)
}
