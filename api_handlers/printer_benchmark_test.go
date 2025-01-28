package api_handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// PrintJobTest สำหรับทดสอบ
type PrintJobTest struct {
	ID        int    `json:"id"`
	Content   string `json:"content"`
	PrinterID string `json:"printer_id"`
	Status    string `json:"status"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// สร้างข้อมูลทดสอบ
func generateTestPrintJobs(count int) []PrintJobTest {
	jobs := make([]PrintJobTest, count)
	mockContent := make([]byte, 5*1024) // ลดขนาดลงเพื่อการทดสอบ
	for i := range mockContent {
		mockContent[i] = byte(i % 256)
	}
	encodedContent := base64.StdEncoding.EncodeToString(mockContent)

	for i := 0; i < count; i++ {
		jobs[i] = PrintJobTest{
			ID:        i + 1,
			Content:   encodedContent,
			PrinterID: fmt.Sprintf("PRINTER_%d", i%5),
			Status:    "pending",
		}
	}
	return jobs
}

// ทดสอบ HTTP Polling
func BenchmarkHTTPPolling(b *testing.B) {
	jobs := generateTestPrintJobs(100000)
	var jobIndex atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentIndex := int(jobIndex.Load())
		if currentIndex < len(jobs) {
			batchSize := 20
			endIndex := currentIndex + batchSize
			if endIndex > len(jobs) {
				endIndex = len(jobs)
			}

			response := jobs[currentIndex:endIndex]
			jobIndex.Add(int32(batchSize))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]PrintJobTest{})
		}
	}))
	defer server.Close()

	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	b.ResetTimer()
	start := time.Now()

	receivedJobs := 0
	for i := 0; i < b.N && receivedJobs < len(jobs); i++ {
		resp, err := http.Get(server.URL)
		if err != nil {
			b.Fatal(err)
		}

		var batch []PrintJobTest
		if err := json.NewDecoder(resp.Body).Decode(&batch); err != nil {
			resp.Body.Close()
			b.Fatal(err)
		}
		resp.Body.Close()

		receivedJobs += len(batch)
		if len(batch) == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}

	duration := time.Since(start)

	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	b.ReportMetric(float64(duration.Milliseconds())/float64(receivedJobs), "ms/job")
	b.ReportMetric(float64(memStatsAfter.Alloc-memStatsBefore.Alloc)/1024/1024, "MB")
	b.ReportMetric(float64(receivedJobs)/duration.Seconds(), "jobs/sec")
}

// ทดสอบ WebSocket
func BenchmarkWebSocket(b *testing.B) {
	jobs := generateTestPrintJobs(100)
	var jobIndex atomic.Int32
	var wg sync.WaitGroup

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			currentIndex := int(jobIndex.Load())
			if currentIndex >= len(jobs) {
				break
			}

			idx := jobIndex.Add(1) - 1
			if idx < int32(len(jobs)) {
				if err := conn.WriteJSON(jobs[idx]); err != nil {
					break
				}
			} else {
				break
			}
		}
	}))
	defer server.Close()

	wsURL := fmt.Sprintf("ws%s", server.URL[4:])

	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	b.ResetTimer()
	start := time.Now()

	wg.Add(1)
	receivedJobs := 0
	go func() {
		defer wg.Done()
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			b.Error(err)
			return
		}
		defer conn.Close()

		for receivedJobs < len(jobs) {
			var job PrintJobTest
			if err := conn.ReadJSON(&job); err != nil {
				break
			}
			receivedJobs++
		}
	}()

	wg.Wait()
	duration := time.Since(start)

	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	b.ReportMetric(float64(duration.Milliseconds())/float64(receivedJobs), "ms/job")
	b.ReportMetric(float64(memStatsAfter.Alloc-memStatsBefore.Alloc)/1024/1024, "MB")
	b.ReportMetric(float64(receivedJobs)/duration.Seconds(), "jobs/sec")
}

// ทดสอบ Concurrent Connections
func BenchmarkConcurrentOperations(b *testing.B) {
	jobs := generateTestPrintJobs(100)

	b.Run("HTTP-Concurrent", func(b *testing.B) {
		var jobIndex atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idx := int(jobIndex.Add(1) - 1)
			if idx < len(jobs) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode([]PrintJobTest{jobs[idx]})
			} else {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode([]PrintJobTest{})
			}
		}))
		defer server.Close()

		b.ResetTimer()
		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := http.Get(server.URL)
				if err == nil {
					resp.Body.Close()
				}
			}()
		}
		wg.Wait()
	})

	b.Run("WebSocket-Concurrent", func(b *testing.B) {
		var jobIndex atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()

			idx := int(jobIndex.Add(1) - 1)
			if idx < len(jobs) {
				conn.WriteJSON(jobs[idx])
			}
		}))
		defer server.Close()

		wsURL := fmt.Sprintf("ws%s", server.URL[4:])
		b.ResetTimer()

		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
				if err == nil {
					conn.Close()
				}
			}()
		}
		wg.Wait()
	})
}
