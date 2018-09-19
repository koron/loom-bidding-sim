package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var (
	sid int
	num int
)

type resp struct {
	id    int
	score int
}

func main() {
	rand.Seed(time.Now().Unix())
	flag.IntVar(&sid, "sid", rand.Intn(1000), "session id")
	flag.IntVar(&num, "num", 100, "num of request")
	flag.Parse()
	r := run(strconv.Itoa(sid), num)
	if r == nil {
		log.Fatal("run returns nil")
	}
	log.Printf("id=%d score=%d", r.id, r.score)
}

func run(sid string, num int) *resp {
	log.Printf("sid=%q num=%d", sid, num)
	ch := make(chan *resp)
	for i := 0; i < num; i++ {
		go func(sid string, rid int) {
			score, err := get(sid, rid)
			if err != nil {
				log.Printf("faield sid=%s rid=%d: %s", sid, rid)
				return
			}
			ch <- &resp{id: rid, score: score}
		}(sid, i)
	}
	var result *resp
	t := time.NewTimer(100 * time.Millisecond)
	for {
		select {
		case r := <-ch:
			if result == nil || r.score > result.score {
				result = r
			}
		case <-t.C:
			//close(ch)
			return result
		}
	}
	return nil
}

func get(sid string, rid int) (int, error) {
	u := fmt.Sprintf("http://127.0.0.1:8080/?sid=%s&rid=%d", sid, rid)
	resp, err := http.Get(u)
	if err != nil {
		return 0, err
	}
	if resp.Body == nil {
		return 0, errors.New("empty body")
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(string(b))
	if err != nil {
		return 0, err
	}
	return n, nil
}
