package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

var webserver_port = getEnv("LISTEN_PORT", "5005")

var unifi_host = getEnv("UNIFI_HOST", "")
var unifi_siteID = getEnv("UNIFI_SITE_ID", "")
var unifi_apiKey = getEnv("UNIFI_NETWORK_API_KEY", "")

var qrcode_ssid = getEnv("QRCODE_SSID", "")
var qrcode_password = getEnv("QRCODE_PASSWORD", "")
var qrcode_wifi_auth_type = getEnv("QRCODE_WIFI_AUTH_TYPE", "nopass")
var qrcode_wifi_hidden = getEnv("QRCODE_WIFI_HIDDEN", "false")

// API response structs
type VoucherResponse struct {
	Offset     int       `json:"offset"`
	Limit      int       `json:"limit"`
	Count      int       `json:"count"`
	TotalCount int       `json:"totalCount"`
	Data       []Voucher `json:"data"`
}

type Voucher struct {
	ID                   string    `json:"id"`
	CreatedAt            time.Time `json:"createdAt"`
	Name                 string    `json:"name"`
	Code                 string    `json:"code"`
	AuthorizedGuestLimit int       `json:"authorizedGuestLimit"`
	AuthorizedGuestCount int       `json:"authorizedGuestCount"`
	ActivatedAt          time.Time `json:"activatedAt"`
	ExpiresAt            time.Time `json:"expiresAt"`
	Expired              bool      `json:"expired"`
	TimeLimitMinutes     int       `json:"timeLimitMinutes"`
	DataUsageLimitMBytes int       `json:"dataUsageLimitMBytes"`
	RxRateLimitKbps      int       `json:"rxRateLimitKbps"`
	TxRateLimitKbps      int       `json:"txRateLimitKbps"`
}

type WifiInfo struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
}

// WiFi JSON Handler
func wifiHandler(w http.ResponseWriter, r *http.Request) {
	info := WifiInfo{
		SSID:     qrcode_ssid,
		Password: qrcode_password,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "JSON encode error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// Voucher API call
func fetchVouchers(apiURL, apiKey string) (*VoucherResponse, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result VoucherResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Generate WiFi QR
func GenerateWifiQR(ssid, password, authType string, hidden bool, filename string) error {

	qrContent := fmt.Sprintf(
		"WIFI:T:%s;S:%s;P:%s;H:%t;;",
		authType,
		ssid,
		password,
		hidden,
	)

	q, err := qrcode.New(qrContent, qrcode.Medium)
	if err != nil {
		return err
	}

	q.DisableBorder = true

	return qrcode.WriteColorFile(qrContent, qrcode.Medium, 256, color.RGBA{R: 0, G: 0, B: 0, A: 0}, color.White, filename)
}

// Voucher handler
func vouchersHandler(apiURL, apiKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vouchers, err := fetchVouchers(apiURL, apiKey)
		if err != nil {
			http.Error(w, "API error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Nur die Codes extrahieren
		codes := make([]string, len(vouchers.Data))
		for i, v := range vouchers.Data {
			codes[i] = v.Code
		}

		// JSON-Ausgabe
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(codes); err != nil {
			http.Error(w, "JSON encode error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// ---- ACCESS LOGGING MIDDLEWARE ----

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)

		log.Printf(
			"%s - %s %s %d (%s)",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			duration,
		)
	})
}

// ------------------------------------

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func main() {

	port, err := strconv.Atoi(webserver_port)
	if err != nil {
		port = 5005
	}

	if unifi_host == "" || unifi_siteID == "" || unifi_apiKey == "" {
		log.Fatal("Please set UNIFI_HOST, UNIFI_SITE_ID and UNIFI_NETWORK_API_KEY")
	}

	wifi_hidden, err := strconv.ParseBool(qrcode_wifi_hidden)
	if err != nil {
		wifi_hidden = false
	}

	err = GenerateWifiQR(qrcode_ssid, qrcode_password, qrcode_wifi_auth_type, wifi_hidden, "static/img/wifi.png")
	if err != nil {
		panic(err)
	}

	apiURL := fmt.Sprintf(
		"https://%s/proxy/network/integration/v1/sites/%s/hotspot/vouchers?limit=7&filter=authorizedGuestCount.eq(0)",
		unifi_host,
		unifi_siteID,
	)

	// Static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", accessLogMiddleware(fs))

	// JSON Endpoints
	http.Handle("/json/wifi_data", accessLogMiddleware(http.HandlerFunc(wifiHandler)))
	http.Handle("/json/vouchers", accessLogMiddleware(http.HandlerFunc(vouchersHandler(apiURL, unifi_apiKey))))

	addr := fmt.Sprintf(":%d", port)
	fmt.Println("Webserver listening on http://localhost" + addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
