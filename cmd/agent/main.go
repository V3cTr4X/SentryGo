package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"gopkg.in/yaml.v3"
)

// Config representa la estructura del archivo YAML
type Config struct {
	ServerURL string `yaml:"server_url"`
	Token     string `yaml:"token"`
	Interval  int    `yaml:"interval"` // segundos
}

// Metrics igual que antes
type Metrics struct {
	Hostname    string    `json:"hostname"`
	Timestamp   time.Time `json:"timestamp"`
	CPUPercent  float64   `json:"cpu_percent"`
	MemTotal    uint64    `json:"mem_total"`
	MemUsed     uint64    `json:"mem_used"`
	MemPercent  float64   `json:"mem_percent"`
	DiskTotal   uint64    `json:"disk_total"`
	DiskUsed    uint64    `json:"disk_used"`
	DiskPercent float64   `json:"disk_percent"`
}

func main() {
	// 1. Leer configuración
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	// 2. Obtener hostname
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Error obteniendo hostname: %v", err)
	}

	// 3. Bucle principal
	for {
		metrics := collectMetrics(hostname)

		err := sendMetrics(config.ServerURL, config.Token, metrics)
		if err != nil {
			log.Printf("Error enviando métricas: %v", err)
		} else {
			log.Println("Métricas enviadas correctamente")
		}

		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}

// loadConfig lee y parsea el archivo YAML
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}

// collectMetrics obtiene todas las métricas
func collectMetrics(hostname string) Metrics {
	m := Metrics{
		Hostname:  hostname,
		Timestamp: time.Now(),
	}

	// CPU (con intervalo de 1 segundo para obtener valor real)
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		m.CPUPercent = cpuPercent[0]
	}

	// Memoria
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		m.MemTotal = memInfo.Total
		m.MemUsed = memInfo.Used
		m.MemPercent = memInfo.UsedPercent
	}

	// Disco (raíz)
	diskInfo, err := disk.Usage("/")
	if err == nil {
		m.DiskTotal = diskInfo.Total
		m.DiskUsed = diskInfo.Used
		m.DiskPercent = diskInfo.UsedPercent
	}

	return m
}

// sendMetrics envía las métricas al servidor vía HTTP POST
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