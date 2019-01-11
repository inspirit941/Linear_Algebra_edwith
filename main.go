package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
)

// https://youtu.be/nvijc5J-JAQ?si=xEprwjHq8K6PlsUh
// SSE: one-way communication from your server to your client
// useful when you need real-time update, 아니면 return message가 필요 없는 경우라던가
func main() {
	http.HandleFunc("/events", sseHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("unable to start server: %s", err.Error())
	}
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	// set headers
	// 타입 설정.
	w.Header().Set("Content-Type", "text/event-stream")
	// 이벤트 받는 쪽에서 캐시하지 않도록 설정
	w.Header().Set("Cache-Control", "no-cache")
	// connection 유지
	w.Header().Set("Connection", "keep-alive")
	// 로컬서버에서 cors 안 나게 설정
	w.Header().Set("Access-Control-Allow-Origin", "*")

	memTick := time.NewTicker(time.Second)
	defer memTick.Stop()

	cpuTick := time.NewTicker(time.Second)
	defer cpuTick.Stop()

	clientGone := r.Context().Done()

	// response controller
	rc := http.NewResponseController(w)
	for {
		select {
		case <-memTick.C:
			m, err := mem.VirtualMemory()
			if err != nil {
				log.Printf("unable to get memory: %s", err.Error())
			}
			if _, err := fmt.Fprintf(w, "event:mem\ndata:Total: %d, Used:%d Percenage: %.2f%%\n\n", m.Total, m.Used, m.UsedPercent); err != nil {
				log.Printf("unable to write event: %s", err.Error())
				return
			}
			rc.Flush() // 이벤트를 외부로 전달
		case <-cpuTick.C:
			c, err := cpu.Times(false) // total cpu 사용량
			if err != nil {
				log.Printf("unable to get CPU info: %s", err.Error())
			}
			if _, err := fmt.Fprintf(w, "event:cpu\ndata:User: %.2f, System:%.2f Idle: %.2f\n\n", c[0].User, c[0].System, c[0].Idle); err != nil {
				log.Printf("unable to write event: %s", err.Error())
				return
			}
			rc.Flush() // 이벤트를 외부로 전달
		case <-clientGone:
			fmt.Println("client has disconnected")
		}
	}
}
