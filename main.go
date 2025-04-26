package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/go-ping/ping"
	"github.com/pelletier/go-toml/v2"
)

type MonitorConfig struct {
	WireguardInterface   string `toml:"wireguard_interface"`
	PingTarget           string `toml:"ping_target"`
	InitialIntervalSecs  int    `toml:"initial_interval_seconds"`
	MaxIntervalSecs      int    `toml:"max_interval_seconds"`
}

type Config struct {
	Monitor MonitorConfig `toml:"monitor"`
}

var (
	config          Config
	currentInterval time.Duration
)

func loadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Fehler beim Laden der Config: %v", err)
	}

	err = toml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Fehler beim Parsen der Config: %v", err)
	}

	currentInterval = time.Duration(config.Monitor.InitialIntervalSecs) * time.Second
}

func isVPNAlive() bool {
	pinger, err := ping.NewPinger(config.Monitor.PingTarget)
	if err != nil {
		log.Printf("Fehler beim Erstellen des Pingers: %v", err)
		return false
	}
	pinger.Count = 1
	pinger.Timeout = 2 * time.Second

	err = pinger.Run()
	if err != nil {
		log.Printf("Ping-Fehler: %v", err)
		return false
	}

	stats := pinger.Statistics()
	return stats.PacketsRecv > 0
}

func restartWireGuard() {
	log.Printf("VPN scheint down. Versuche Neustart von %s...", config.Monitor.WireguardInterface)
	cmd := exec.Command("systemctl", "restart", "wg-quick@"+config.Monitor.WireguardInterface)
	err := cmd.Run()
	if err != nil {
		log.Printf("Fehler beim Neustart: %v", err)
	} else {
		log.Println("WireGuard Service erfolgreich neu gestartet.")
	}
}

func main() {
	log.Println("Starte WireGuard Monitor (Go Version mit Config)...")

	loadConfig("config.toml")

	for {
		if isVPNAlive() {
			log.Println("VPN ist aktiv.")
			currentInterval = time.Duration(config.Monitor.InitialIntervalSecs) * time.Second
		} else {
			log.Println("VPN ist inaktiv.")
			restartWireGuard()
			currentInterval *= 2
			maxInterval := time.Duration(config.Monitor.MaxIntervalSecs) * time.Second
			if currentInterval > maxInterval {
				currentInterval = maxInterval
			}
			log.Printf("Neues Prüfintervall: %s", currentInterval)
		}
		time.Sleep(currentInterval)
	}
}

