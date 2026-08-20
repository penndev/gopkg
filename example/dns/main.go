package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const resolveAPI = "https://doh.pub/resolve"

func main() {
	addr := flag.String("addr", "127.0.0.1:53", "UDP 监听地址")
	flag.Parse()

	udpAddr, err := net.ResolveUDPAddr("udp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	httpClient := &http.Client{Timeout: 8 * time.Second}
	log.Printf("UDP %s  ->  %s?name=&type=", conn.LocalAddr(), resolveAPI)

	buf := make([]byte, 4096)
	for {
		n, client, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("read: %v", err)
			continue
		}
		q := append([]byte(nil), buf[:n]...)
		go serve(conn, client, q, httpClient)
	}
}

func serve(conn *net.UDPConn, client *net.UDPAddr, raw []byte, httpClient *http.Client) {
	req := new(dns.Msg)
	if err := req.Unpack(raw); err != nil {
		log.Printf("unpack %s: %v", client, err)
		return
	}
	if len(req.Question) == 0 {
		return
	}

	resp, err := queryResolve(httpClient, req)
	if err != nil {
		log.Printf("resolve %s: %v", client, err)
		fail := new(dns.Msg)
		fail.SetRcode(req, dns.RcodeServerFailure)
		resp = fail
	}

	out, err := resp.Pack()
	if err != nil {
		log.Printf("pack: %v", err)
		return
	}
	if _, err := conn.WriteToUDP(out, client); err != nil {
		log.Printf("write %s: %v", client, err)
	}
}

func queryResolve(httpClient *http.Client, req *dns.Msg) (*dns.Msg, error) {
	q := req.Question[0]
	name := strings.TrimSuffix(q.Name, ".")
	typ := dns.TypeToString[q.Qtype]
	if typ == "" {
		typ = fmt.Sprintf("%d", q.Qtype)
	}

	u := resolveAPI + "?name=" + url.QueryEscape(name) + "&type=" + url.QueryEscape(typ)
	httpReq, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "application/dns-json, application/json")

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(httpResp.Body, 65535))
	if err != nil {
		return nil, err
	}

	var jr jsonMsg
	if err := json.Unmarshal(body, &jr); err != nil {
		return nil, err
	}
	return jr.toMsg(req)
}

// Google / doh.pub JSON API：https://developers.google.com/speed/public-dns/docs/doh/json
type jsonMsg struct {
	Status int      `json:"Status"`
	TC     bool     `json:"TC"`
	RA     bool     `json:"RA"`
	AD     bool     `json:"AD"`
	CD     bool     `json:"CD"`
	Answer []jsonRR `json:"Answer"`
}

type jsonRR struct {
	Name string `json:"name"`
	Type uint16 `json:"type"`
	TTL  uint32 `json:"TTL"`
	Data string `json:"data"`
}

func (j jsonMsg) toMsg(req *dns.Msg) (*dns.Msg, error) {
	m := new(dns.Msg)
	m.SetReply(req)
	m.RecursionAvailable = j.RA
	m.AuthenticatedData = j.AD
	m.CheckingDisabled = j.CD
	m.Truncated = j.TC
	m.Rcode = j.Status
	for _, a := range j.Answer {
		typ := dns.TypeToString[a.Type]
		if typ == "" {
			continue
		}
		rr, err := dns.NewRR(fmt.Sprintf("%s %d IN %s %s", a.Name, a.TTL, typ, a.Data))
		if err != nil {
			log.Printf("parse rr %s: %v", a.Data, err)
			continue
		}
		m.Answer = append(m.Answer, rr)
	}
	return m, nil
}
