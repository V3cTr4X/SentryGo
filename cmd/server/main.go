package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Metric struct {
	ID          uint      `gorm:"primaryKey"`
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

type Host struct {
	Hostname string `gorm:"primaryKey"`
	Token    string
	LastSeen time.Time
}

var db *gorm.DB

func main() {
    var err error
    db, err = gorm.Open(sqlite.Open("sentrygo.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    db.AutoMigrate(&Metric{}, &Host{})

    r := gin.Default()

    // Ruta raíz: servir el archivo index.html
    r.GET("/", func(c *gin.Context) {
        c.File("./web/index.html")
    })

    // Si tienes otros archivos estáticos (CSS, JS, etc.), sírvelos desde /static
    // r.Static("/static", "./web/static") // opcional si necesitas assets

    r.POST("/api/metrics", receiveMetrics)
    r.GET("/api/hosts", getHosts)
    r.GET("/api/metrics/:hostname", getMetrics)
    r.Run(":8080")
}

func receiveMetrics(c *gin.Context) {
	var metric Metric
	if err := c.ShouldBindJSON(&metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Token check (pendiente)
	if err := db.Create(&metric).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	host := Host{Hostname: metric.Hostname, LastSeen: time.Now()}
	db.Where(Host{Hostname: metric.Hostname}).Assign(Host{LastSeen: time.Now()}).FirstOrCreate(&host)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getHosts(c *gin.Context) {
	var hosts []Host
	db.Find(&hosts)
	c.JSON(http.StatusOK, hosts)
}

func getMetrics(c *gin.Context) {
	hostname := c.Param("hostname")
	var metrics []Metric
	db.Where("hostname = ?", hostname).Order("timestamp desc").Limit(100).Find(&metrics)
	c.JSON(http.StatusOK, metrics)
}