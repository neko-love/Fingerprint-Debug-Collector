package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Record struct {
	Time    string      `json:"time"`
	IP      string      `json:"ip"`
	JA3     string      `json:"ja3"`
	JA4     string      `json:"ja4"`
	H2      string      `json:"h2"`
	UA      string      `json:"ua"`
	Headers http.Header `json:"headers"`
}

var (
	mu   sync.Mutex
	file *os.File
)

func main() {
	var err error
	file, err = os.OpenFile("/root/fingerprints.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rec := Record{
			Time:    time.Now().Format(time.RFC3339),
			IP:      firstNonEmpty(r.Header.Get("X-Forwarded-For"), r.RemoteAddr),
			JA3:     r.Header.Get("X-Ja3-Fingerprint"),
			JA4:     r.Header.Get("X-Ja4-Fingerprint"),
			H2:      r.Header.Get("X-Http2-Fingerprint"),
			UA:      r.UserAgent(),
			Headers: r.Header,
		}

		mu.Lock()
		line, _ := json.Marshal(rec)
		file.WriteString(string(line) + "\n")
		mu.Unlock()
		log.Println(string(line))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rec)
	})

	log.Println("后端启动在 :3005，日志：/root/fingerprints.log")
	log.Fatal(http.ListenAndServe(":3005", nil))
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return "N/A"
}