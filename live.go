package livelygo

import (
	"context"
	_ "embed"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

//go:embed lively.js
var javascript string

func Javascript(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/javascript")
	fmt.Fprintf(w, javascript)
}

type message struct {
	Kind string `json:"kind"`
}

type initMsg struct {
	message
	Bindings []string `json:"bindings"`
	Pushers  []string `json:"pushers"`
}

type pushMsg struct {
	message
	Name  string      `json:"name"`
	Event interface{} `json:"event"`
}

func NewPage(f http.HandlerFunc) http.HandlerFunc {
	// scope the context request to a page
	page := &page{
		sessions: make(map[string]*session),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") != "websocket" {
			// render the user content page
			s := page.NewSession()
			http.SetCookie(w, &http.Cookie{
				Name:  "livelygo-session",
				Value: s.Key,
				Path:  r.URL.Path,
			})
			ctx := context.WithValue(r.Context(), "page", s)
			f.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		cookie, err := r.Cookie("livelygo-session")
		sess := page.ActivateSession(cookie.Value)

		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			log.Printf("err: %v", err)
			return
		}
		sess.active <- 1
		for {
			// one way bindings for now server -> client
			cmd := <-sess.bindings
			err = wsjson.Write(r.Context(), c, cmd)
			if err != nil {
				page.DeleteSession(sess.Key)
			}
		}
	}
}

type page struct {
	sessionsMu sync.Mutex
	sessions   map[string]*session
}

func (p *page) NewSession() *session {
	p.sessionsMu.Lock()
	key := make([]byte, 8)
	rand.Read(key)
	digest := hex.EncodeToString(key)
	session := &session{
		Key:      digest,
		active:   make(chan int),
		bindings: make(chan setCmd),
	}
	p.sessions[digest] = session
	p.sessionsMu.Unlock()
	log.Printf("new session: %s\n", digest)
	return session
}

func (p *page) DeleteSession(key string) {
	p.sessionsMu.Lock()
	session := p.sessions[key]
	close(session.bindings)
	close(session.active)
	delete(p.sessions, key)
	p.sessionsMu.Unlock()
	log.Printf("session over: %s\n", key)
}

func (p *page) ActivateSession(key string) *session {
	p.sessionsMu.Lock()
	sess := p.sessions[key]
	p.sessionsMu.Unlock()
	log.Printf("activated session: %s\n", key)
	return sess
}

type setCmd struct {
	Kind  string `json:"kind"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type session struct {
	Key      string
	active   chan int
	bindings chan setCmd
}

func (s *session) SetVar(key, value string) {
	s.bindings <- setCmd{"setCmd", key, value}
}

func WaitForPage(r *http.Request) *session {
	s := r.Context().Value("page").(*session)
	<-s.active // wait for active session
	return s
}
