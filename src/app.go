package main

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/ssh"
)

var (
	luaScript   []byte
	sshHost     string
	sshPort     string
	sshUser     string
	sshPassword string

	system_status = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "system_status",
		Help: "System status",
	}, []string{"hostname", "machine", "architecture", "kernel_release", "openwrt_release", "local_time", "uptime_str"})
	system_loadAverage_1m = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_load_average",
		Help: "System load average",
		ConstLabels: prometheus.Labels{
			"duration": "1m",
		},
	})
	system_loadAverage_5m = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_load_average",
		Help: "System load average",
		ConstLabels: prometheus.Labels{
			"duration": "5m",
		},
	})
	system_loadAverage_15m = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_load_average",
		Help: "System load average",
		ConstLabels: prometheus.Labels{
			"duration": "15m",
		},
	})

	pon_ploamStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_ploam_status",
		Help: "PLOAM status (Code)",
	})
	pon_cpu0Temp_celsius = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_cpu0_temp",
		Help: "CPU0 temperature",
		ConstLabels: prometheus.Labels{
			"unit": "celsius",
		},
	})
	pon_cpu0Temp_farhenheit = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_cpu0_temp",
		Help: "CPU0 temperature",
		ConstLabels: prometheus.Labels{
			"unit": "farhenheit",
		},
	})
	pon_cpu1Temp_celsius = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_cpu1_temp",
		Help: "CPU1 temperature",
		ConstLabels: prometheus.Labels{
			"unit": "celsius",
		},
	})
	pon_cpu1Temp_farhenheit = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_cpu1_temp",
		Help: "CPU1 temperature",
		ConstLabels: prometheus.Labels{
			"unit": "farhenheit",
		},
	})
	pon_opticTemp_celsius = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_optic_temp",
		Help: "Optic temperature",
		ConstLabels: prometheus.Labels{
			"unit": "celsius",
		},
	})
	pon_opticTemp_farhenheit = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_optic_temp",
		Help: "Optic temperature",
		ConstLabels: prometheus.Labels{
			"unit": "farhenheit",
		},
	})
	pon_voltage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_voltage",
		Help: "Voltage",
		ConstLabels: prometheus.Labels{
			"unit": "volts",
		},
	})
	pon_txBias = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_tx_bias",
		Help: "TX Bias",
		ConstLabels: prometheus.Labels{
			"unit": "mA",
		},
	})
	pon_txPower_mW = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_tx_power",
		Help: "TX Power",
		ConstLabels: prometheus.Labels{
			"unit": "mW",
		},
	})
	pon_txPower_dBm = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_tx_power",
		Help: "TX Power",
		ConstLabels: prometheus.Labels{
			"unit": "dBm",
		},
	})
	pon_rxPower_mW = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_rx_power",
		Help: "RX Power",
		ConstLabels: prometheus.Labels{
			"unit": "mW",
		},
	})
	pon_rxPower_dBm = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_rx_power",
		Help: "RX Power",
		ConstLabels: prometheus.Labels{
			"unit": "dBm",
		},
	})
	pon_ethSpeed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pon_eth_speed",
		Help: "Ethernet speed",
		ConstLabels: prometheus.Labels{
			"unit": "Mbps",
		},
	})
	pon_moduleInfo = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pon_stick_info",
		Help: "PON stick information",
	}, []string{"vendor_name", "vendor_pn", "vendor_rev", "pon_mode", "module_type", "active_bank"})

	memory_total_bytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "memory_total",
		Help: "Total memory",
		ConstLabels: prometheus.Labels{
			"unit": "bytes",
		},
	})
	memory_free_bytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "memory_free",
		Help: "Free memory",
		ConstLabels: prometheus.Labels{
			"unit": "bytes",
		},
	})
	memory_buffers_bytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "memory_buffers",
		Help: "Memory buffers",
		ConstLabels: prometheus.Labels{
			"unit": "bytes",
		},
	})
	memory_cached_bytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "memory_cached",
		Help: "Memory cached",
		ConstLabels: prometheus.Labels{
			"unit": "bytes",
		},
	})
	memory_used_bytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "memory_used",
		Help: "Used memory",
		ConstLabels: prometheus.Labels{
			"unit": "bytes",
		},
	})
	memory_used_percent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "memory_used",
		Help: "Used memory",
		ConstLabels: prometheus.Labels{
			"unit": "percent",
		},
	})
)

func init() {
	sshHost = os.Getenv("SSH_HOST")
	if sshHost == "" {
		sshHost = "192.168.11.1"
	}
	sshPort = os.Getenv("SSH_PORT")
	if sshPort == "" {
		sshPort = "22"
	}
	sshUser = os.Getenv("SSH_USERNAME")
	if sshUser == "" {
		sshUser = "root"
	}
	sshPassword = os.Getenv("SSH_PASSWORD")
	if sshPassword == "" {
		log.Fatal("SSH_PASSWORD environment variable is required")
	}
}

func main() {
	filePath := "gpon_status.lua"
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to read Lua script: %v", err)
	}
	luaScript = content
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/metrics", metricsHandler)
	log.Printf("listening on :9100")
	log.Fatal(http.ListenAndServe(":9100", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "8311 PON Stick Exporter")
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	scrape()
	promhttp.Handler().ServeHTTP(w, r)
}

func runLuaScript(client *ssh.Client) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	command := fmt.Sprintf("lua -e '%s'", string(luaScript))
	if err := session.Run(command); err != nil {
		return "", fmt.Errorf("failed to execute command: %w, stderr: %s", err, stderrBuf.String())
	}

	return strings.Trim(stdoutBuf.String(), "\n"), nil
}

func setLoadAverage(loadAverageValue string, gauge prometheus.Gauge) {
	value, err := strconv.ParseFloat(loadAverageValue, 64)
	if err != nil {
		log.Printf("failed to parse load average: %v", err)
		return
	}
	gauge.Set(value)
}

func setValue(stringValue string, gauge prometheus.Gauge) {
	value, err := strconv.ParseFloat(stringValue, 64)
	if err != nil {
		log.Printf("failed to parse temperature: %v", err)
		return
	}
	gauge.Set(value)
}

func toFarhenheit(celsius float64) float64 {
	return (celsius * 9 / 5) + 32
}

func setTemperature(strValue string, celsiusGauge prometheus.Gauge, farhenheitGauge prometheus.Gauge) {
	celsiusValue, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		log.Printf("failed to parse temperature: %v", err)
		return
	}
	celsiusGauge.Set(celsiusValue)
	farhenheitValue := toFarhenheit(celsiusValue)
	farhenheitGauge.Set(farhenheitValue)
}

func toDbm(ma float64) float64 {
	return 10 * math.Log10(ma)
}

func setPower(strValue string, maGauge prometheus.Gauge, dbmGauge prometheus.Gauge) {
	maValue, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		log.Printf("failed to parse power: %v", err)
		return
	}
	maGauge.Set(maValue)
	dbmValue := toDbm(maValue)
	dbmGauge.Set(dbmValue)
}

func scrape() {
	log.Printf("scraping metrics")

	system_status.Reset()
	system_loadAverage_1m.Set(0)
	system_loadAverage_5m.Set(0)
	system_loadAverage_15m.Set(0)
	pon_moduleInfo.Reset()
	pon_ploamStatus.Set(0)
	pon_cpu0Temp_celsius.Set(0)
	pon_cpu0Temp_farhenheit.Set(0)
	pon_cpu1Temp_celsius.Set(0)
	pon_cpu1Temp_farhenheit.Set(0)
	pon_opticTemp_celsius.Set(0)
	pon_opticTemp_farhenheit.Set(0)
	pon_voltage.Set(0)
	pon_txBias.Set(0)
	pon_txPower_mW.Set(0)
	pon_txPower_dBm.Set(0)
	pon_rxPower_mW.Set(0)
	pon_rxPower_dBm.Set(0)
	pon_ethSpeed.Set(0)
	memory_total_bytes.Set(0)
	memory_free_bytes.Set(0)
	memory_buffers_bytes.Set(0)
	memory_cached_bytes.Set(0)
	memory_used_bytes.Set(0)
	memory_used_percent.Set(0)

	authMethods := []ssh.AuthMethod{ssh.Password(sshPassword)}
	config := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", sshHost+":"+sshPort, config)
	if err != nil {
		log.Printf("failed to dial SSH: %v", err)
		return
	}
	defer client.Close()

	output, err := runLuaScript(client)
	if err != nil {
		log.Printf("failed to run Lua script: %v", err)
		return
	}
	lines := strings.Split(output, "\n")

	systemValues := strings.Split(lines[0], "\t")
	if len(systemValues) < 7 {
		log.Printf("unexpected number of systemValues from Lua script: %d", len(systemValues))
		return
	}
	ponValues := strings.Split(lines[1], "\t")
	if len(ponValues) < 15 {
		log.Printf("unexpected number of ponValues from Lua script: %d", len(ponValues))
		return
	}

	memoryValues := strings.Split(lines[2], "\t")
	if len(memoryValues) < 6 {
		log.Printf("unexpected number of memoryValues from Lua script: %d", len(memoryValues))
		return
	}

	system_status.WithLabelValues(systemValues[0], systemValues[1], systemValues[2], systemValues[3], systemValues[4], systemValues[5], systemValues[6]).Set(1)
	loadAverageValues := strings.Split(systemValues[7], " ")
	setLoadAverage(loadAverageValues[0], system_loadAverage_1m)
	setLoadAverage(loadAverageValues[1], system_loadAverage_5m)
	setLoadAverage(loadAverageValues[2], system_loadAverage_15m)

	setValue(ponValues[0], pon_ploamStatus)
	setTemperature(ponValues[1], pon_cpu0Temp_celsius, pon_cpu0Temp_farhenheit)
	setTemperature(ponValues[2], pon_cpu1Temp_celsius, pon_cpu1Temp_farhenheit)
	setTemperature(ponValues[3], pon_opticTemp_celsius, pon_opticTemp_farhenheit)
	setValue(ponValues[4], pon_voltage)
	setValue(ponValues[5], pon_txBias)
	setPower(ponValues[6], pon_txPower_mW, pon_txPower_dBm)
	setPower(ponValues[7], pon_rxPower_mW, pon_rxPower_dBm)
	setValue(ponValues[8], pon_ethSpeed)
	pon_moduleInfo.WithLabelValues(ponValues[9], ponValues[10], ponValues[11], ponValues[12], ponValues[13], ponValues[14]).Set(1)

	setValue(memoryValues[0], memory_total_bytes)
	setValue(memoryValues[1], memory_free_bytes)
	setValue(memoryValues[2], memory_buffers_bytes)
	setValue(memoryValues[3], memory_cached_bytes)
	setValue(memoryValues[4], memory_used_bytes)
	setValue(memoryValues[5], memory_used_percent)
}
