package httpserver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	cetuspoolssystem "github.com/ipoluianov/cetuspools/system"
	"github.com/ipoluianov/cetuspoolsui/repo"
	"github.com/ipoluianov/gomisc/logger"
)

type Host struct {
	Name string
}

type HttpServer struct {
	port   int
	srvTLS *http.Server
	rTLS   *mux.Router
}

func CurrentExePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func NewHttpServer() *HttpServer {
	var c HttpServer
	c.port = 8501
	return &c
}

func (c *HttpServer) Start() {
	logger.Println("HttpServer start")
	go c.thListenTLS()
}

func (c *HttpServer) thListenTLS() {
	tlsConfig := &tls.Config{}
	tlsConfig.Certificates = make([]tls.Certificate, 0)

	cert, err := tls.LoadX509KeyPair(CurrentExePath()+"/bundle.crt", CurrentExePath()+"/private.key")
	if err == nil {
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	} else {
		logger.Println("loading certificates error:", err.Error())
	}

	c.srvTLS = &http.Server{
		Addr:      ":" + fmt.Sprint(c.port),
		TLSConfig: tlsConfig,
	}

	c.rTLS = mux.NewRouter()

	c.rTLS.HandleFunc("/data/{id}", c.processData)
	c.rTLS.HandleFunc("/pool/{id}", c.processPool)

	c.rTLS.NotFoundHandler = http.HandlerFunc(c.processFile)
	c.srvTLS.Handler = c

	logger.Println("HttpServerTLS thListen begin")
	listener, err := tls.Listen("tcp", ":"+fmt.Sprint(c.port), tlsConfig)
	if err != nil {
		logger.Println("TLS Listener error:", err)
		return
	}

	err = c.srvTLS.Serve(listener)
	if err != nil {
		logger.Println("HttpServerTLS thListen error: ", err)
	}
	logger.Println("HttpServerTLS thListen end")
}

func (s *HttpServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}
	// Lets Gorilla work
	s.rTLS.ServeHTTP(rw, req)
}

func (c *HttpServer) Stop() error {
	var err error

	{
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err = c.srvTLS.Shutdown(ctx); err != nil {
			logger.Println(err)
		}
	}
	return err
}

func SplitRequest(path string) []string {
	return strings.FieldsFunc(path, func(r rune) bool {
		return r == '/'
	})
}

func (c *HttpServer) processData(w http.ResponseWriter, r *http.Request) {
	realIP := getRealAddr(r)
	logger.Println("processFile", realIP, r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Request-Method", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}
	parts := strings.FieldsFunc(r.URL.Path, func(r rune) bool {
		return r == '/'
	})
	if len(parts) < 2 {
		logger.Println("processData", "Invalid path")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	v := repo.Get().Get(parts[1])
	w.Write([]byte(v))
}

func (c *HttpServer) processPool(w http.ResponseWriter, r *http.Request) {
	realIP := getRealAddr(r)
	logger.Println("processFile", realIP, r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Request-Method", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}
	parts := strings.FieldsFunc(r.URL.Path, func(r rune) bool {
		return r == '/'
	})
	if len(parts) < 2 {
		logger.Println("processData", "Invalid path")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	poolSymbol := parts[1]

	v := repo.Get().Get("lastData")

	var lastData cetuspoolssystem.CetusStatsPools
	err := json.Unmarshal([]byte(v), &lastData)
	if err != nil {
		logger.Println("processPool", "json.Unmarshal error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return

	}

	type Res struct {
		TVL      string `json:"tvl"`
		Volume   string `json:"volume"`
		TotalApr string `json:"totalApr"`
		Price    string `json:"price"`
		PriceRev string `json:"price_rev"`
	}

	for _, pool := range lastData.Data.LpList {
		if pool.Symbol == poolSymbol {
			var res Res

			totalApr, _ := strconv.ParseFloat(pool.TotalApr, 64)
			totalApr *= 100
			res.TotalApr = fmt.Sprintf("%.0f", totalApr)

			tvl, _ := strconv.ParseFloat(pool.PureTvlInUsd, 64)
			tvlStr := fmt.Sprintf("%.0f", tvl)
			volume, _ := strconv.ParseFloat(pool.VolInUsd24H, 64)
			volumeStr := fmt.Sprintf("%.0f", volume)

			res.TVL = tvlStr
			res.Volume = volumeStr

			price, _ := strconv.ParseFloat(pool.Price, 64)
			res.Price = fmt.Sprintf("%.6f", price)
			if price > 0 {
				priceRev := 1 / price
				res.PriceRev = fmt.Sprintf("%.6f", priceRev)
			} else {
				res.PriceRev = "0"
			}

			v, err := json.Marshal(res)
			if err != nil {
				logger.Println("processPool", "json.Marshal error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(v)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (c *HttpServer) processFile(w http.ResponseWriter, r *http.Request) {
	realIP := getRealAddr(r)
	logger.Println("processFile", realIP, r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Request-Method", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	if strings.Contains(r.URL.Path, "..") {
		logger.Println("processFile", "Path contains '..'")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pathOfDataDirectory := CurrentExePath() + "/www"

	pathToFile := pathOfDataDirectory
	if r.URL.Path == "/" {
		pathToFile += "/index.html"
	} else {
		pathToFile += r.URL.Path
	}

	fileContent, err := os.ReadFile(pathToFile)
	if err != nil {
		logger.Println("processFile", "os.ReadFile Error", err)
		return
	}

	w.Write(fileContent)
}

func getRealAddr(r *http.Request) string {
	remoteIP := ""
	// the default is the originating ip. but we try to find better options because this is almost
	// never the right IP
	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		remoteIP = parts[0]
	}
	// If we have a forwarded-for header, take the address from there
	if xff := strings.Trim(r.Header.Get("X-Forwarded-For"), ","); len(xff) > 0 {
		addrs := strings.Split(xff, ",")
		lastFwd := addrs[len(addrs)-1]
		if ip := net.ParseIP(lastFwd); ip != nil {
			remoteIP = ip.String()
		}
		// parse X-Real-Ip header
	} else if xri := r.Header.Get("X-Real-Ip"); len(xri) > 0 {
		if ip := net.ParseIP(xri); ip != nil {
			remoteIP = ip.String()
		}
	}

	return remoteIP
}
