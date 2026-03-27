package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-ping/ping"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerURL  string `yaml:"server_url"`
	Token      string `yaml:"token"`
	Interval   int    `yaml:"interval"`
	PingTarget string `yaml:"ping_target"`
}

type Metrics struct {
	Hostname        string    `json:"hostname"`
	Timestamp       time.Time `json:"timestamp"`
	CPUPercent      float64   `json:"cpu_percent"`
	MemTotal        uint64    `json:"mem_total"`
	MemUsed         uint64    `json:"mem_used"`
	MemPercent      float64   `json:"mem_percent"`
	DiskTotal       uint64    `json:"disk_total"`
	DiskUsed        uint64    `json:"disk_used"`
	DiskPercent     float64   `json:"disk_percent"`
	PingLatencyMs   float64   `json:"ping_latency_ms"`
	ProcessCount    int32     `json:"process_count"`
	TopProcessCPU   string    `json:"top_process_cpu"`
	TopProcessMem   string    `json:"top_process_mem"`
}

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Error obteniendo hostname: %v", err)
	}

	if config.PingTarget == "" {
		config.PingTarget = "8.8.8.8"
	}

	log.Printf("Iniciando agente SentryGo - Host: %s, Servidor: %s", hostname, config.ServerURL)
	log.Printf("Ping target: %s, Intervalo: %d segundos", config.PingTarget, config.Interval)

	for {
		metrics := collectMetrics(hostname, config.PingTarget)
		
		// Mostrar resumen en consola
		log.Printf("CPU: %.1f%%, RAM: %.1f%%, Disco: %.1f%%, Ping: %.1fms, Procesos: %d",
			metrics.CPUPercent, metrics.MemPercent, metrics.DiskPercent, 
			metrics.PingLatencyMs, metrics.ProcessCount)

		err := sendMetrics(config.ServerURL, config.Token, metrics)
		if err != nil {
			log.Printf("Error enviando métricas: %v", err)
		} else {
			log.Println("✓ Métricas enviadas correctamente")
		}

		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}

func collectMetrics(hostname, pingTarget string) Metrics {
	m := Metrics{
		Hostname:  hostname,
		Timestamp: time.Now(),
	}

	// CPU
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		m.CPUPercent = cpuPercent[0]
	} else {
		log.Printf("Error obteniendo CPU: %v", err)
	}

	// Memoria
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		m.MemTotal = memInfo.Total
		m.MemUsed = memInfo.Used
		m.MemPercent = memInfo.UsedPercent
	} else {
		log.Printf("Error obteniendo memoria: %v", err)
	}

	// Disco
	diskInfo, err := disk.Usage("/")
	if err == nil {
		m.DiskTotal = diskInfo.Total
		m.DiskUsed = diskInfo.Used
		m.DiskPercent = diskInfo.UsedPercent
	} else {
		log.Printf("Error obteniendo disco: %v", err)
	}

	// Ping
	m.PingLatencyMs = getPingLatency(pingTarget)
	
	// Procesos
	m.ProcessCount, m.TopProcessCPU, m.TopProcessMem = getProcessStats()

	return m
}

func getPingLatency(target string) float64 {
	pinger, err := ping.NewPinger(target)
	if err != nil {
		log.Printf("Error creando pinger para %s: %v", target, err)
		return -1
	}
	pinger.Count = 3
	pinger.Timeout = 2 * time.Second
	pinger.SetPrivileged(true)

	err = pinger.Run()
	if err != nil {
		log.Printf("Error ejecutando ping a %s: %v", target, err)
		return -1
	}
	stats := pinger.Statistics()
	if stats.PacketLoss == 100 || stats.AvgRtt == 0 {
		return -1
	}
	return float64(stats.AvgRtt) / float64(time.Millisecond)
}

func getProcessStats() (int32, string, string) {
	processes, err := process.Processes()
	if err != nil {
		log.Printf("Error obteniendo procesos: %v", err)
		return 0, "N/A", "N/A"
	}

	count := int32(len(processes))

	var topCPU float64 = 0
	var topMem float64 = 0
	var topCPUName string = "N/A"
	var topMemName string = "N/A"

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}
		
		cpu, err := p.CPUPercent()
		if err == nil && cpu > topCPU {
			topCPU = cpu
			topCPUName = name
		}
		
		mem32, err := p.MemoryPercent()
		if err == nil {
			mem := float64(mem32)
			if mem > topMem {
				topMem = mem
				topMemName = name
			}
		}
	}

	topCPUFormatted := fmt.Sprintf("%s (%.1f%%)", topCPUName, topCPU)
	topMemFormatted := fmt.Sprintf("%s (%.1f%%)", topMemName, topMem)
	
	return count, topCPUFormatted, topMemFormatted
}

func sendMetrics(url, token string, metrics Metrics) error {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("servidor respondió con status %d", resp.StatusCode)
	}
	return nil
}