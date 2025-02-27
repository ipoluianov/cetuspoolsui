package system

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	cetuspoolssystem "github.com/ipoluianov/cetuspools/system"
	"github.com/ipoluianov/cetuspoolsui/httpserver"
	"github.com/ipoluianov/cetuspoolsui/repo"
	"github.com/ipoluianov/gomisc/logger"
)

type System struct {
	httpServer *httpserver.HttpServer
	lastData   cetuspoolssystem.CetusStatsPools
}

func NewSystem() *System {
	var c System
	return &c
}

func (c *System) Start() {
	c.httpServer = httpserver.NewHttpServer()
	c.httpServer.Start()

	go c.ThWork()
}

func (c *System) Stop() {
	c.httpServer.Stop()
}

func (c *System) CreateZipWithJSON(jsonData []byte) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	fileWriter, err := zipWriter.Create("data.json")
	if err != nil {
		return nil, err
	}

	_, err = fileWriter.Write(jsonData)
	if err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *System) extractFileFromZipImage(zipImage []byte) ([]byte, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipImage), int64(len(zipImage)))
	if err != nil {
		return nil, err
	}

	for _, f := range zipReader.File {
		if f.Name == "data.json" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			return io.ReadAll(rc)
		}
	}

	return nil, nil
}

func (c *System) fetchZipFileFromServer(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (c *System) ThWork() {
	for {
		time.Sleep(1 * time.Second)
		r, err := http.Get("https://u00.io:8488/")
		if err != nil {
			logger.Println("http.Get error:", err)
			continue
		}
		body, err := io.ReadAll(r.Body)
		r.Body.Close()

		if err != nil {
			logger.Println("io.ReadAll error:", err)
			continue
		}

		var items []string
		err = json.Unmarshal(body, &items)
		if err != nil {
			logger.Println("json.Unmarshal error:", err)
			continue
		}

		sort.Slice(items, func(i, j int) bool {
			return items[i] < items[j]
		})

		lastItem := items[len(items)-1]

		zipFileImage, err := c.fetchZipFileFromServer("https://u00.io:8488/" + lastItem)
		if err != nil {
			logger.Println("fetchZipFileFromServer error:", err)
			continue
		}

		jsonData, err := c.extractFileFromZipImage(zipFileImage)
		if err != nil {
			logger.Println("extractFileFromZipImage error:", err)
			continue
		}

		err = json.Unmarshal(jsonData, &c.lastData)
		if err != nil {
			logger.Println("json.Unmarshal error:", err)
			continue
		}

		res := "<html><body><table>"
		res += "<tr>"
		res += "<td>Logo</td>"
		res += "<td>Coin A</td>"
		res += "<td>Logo</td>"
		res += "<td>Coin B</td>"
		res += "<td>Price</td>"
		res += "<td>Price Rev</td>"
		res += "<td>Pure TVL</td>"
		res += "<td>Vol 24H</td>"
		res += "</tr>"

		for _, pool := range c.lastData.Data.LpList {
			//res += pool.Name + " " + pool.CoinA.Symbol + " " + pool.CoinB.Symbol + " " + pool.Price + "\n"

			var price float64
			price, err = strconv.ParseFloat(pool.Price, 64)
			if err != nil {
				price = 0
			}
			priceStr := strconv.FormatFloat(price, 'f', 6, 64)

			priceRev := 0.0
			if price != 0 {
				priceRev = 1.0 / price
			}

			priceRevStr := strconv.FormatFloat(priceRev, 'f', 6, 64)

			res += "<tr>"
			res += "<td><image style='width: 32px; height: 32px' src='" + pool.CoinA.LogoUrl + "'></td>"
			res += "<td>" + pool.CoinA.Symbol + "</td>"
			res += "<td><image style='width: 32px; height: 32px' src='" + pool.CoinB.LogoUrl + "'></td>"
			res += "<td>" + pool.CoinB.Symbol + "</td>"
			res += "<td>" + priceStr + "</td>"
			res += "<td>" + priceRevStr + "</td>"
			res += "<td>" + pool.PureTvlInUsd + "</td>"
			res += "<td>" + pool.VolInUsd24H + "</td>"
			res += "</tr>"
		}

		res += "</table></body></html>"

		repo.Get().Add("items", res)

		//logger.Println("items:", c.lastData)
	}
}
