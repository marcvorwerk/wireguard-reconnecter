package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/go-ping/ping"
	"github.com/pelletier/go-toml/v2"
)

var version = "testing"

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
		log.Fatalf("Error loading config file: %v", err)
	}

	err = toml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	currentInterval = time.Duration(config.Monitor.InitialIntervalSecs) * time.Second
}

func isVPNAlive() bool {
	pinger, err := ping.NewPinger(config.Monitor.PingTarget)
	if err != nil {
		log.Printf("Error: %v", err)
		return false
	}
	pinger.Count = 1
	pinger.Timeout = 2 * time.Second

	err = pinger.Run()
	if err != nil {
		log.Printf("Ping Error: %v", err)
		return false
	}

	stats := pinger.Statistics()
	return stats.PacketsRecv > 0
}

func restartWireGuard() {
	log.Printf("VPN down. Try restart of interface %s...", config.Monitor.WireguardInterface)
	cmd := exec.Command("systemctl", "restart", "wg-quick@"+config.Monitor.WireguardInterface)
	err := cmd.Run()
	if err != nil {
		log.Printf("Error during restart: %v", err)
	} else {
		log.Println("WireGuard Service successful restarted")
	}
}

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	configPath := flag.String("config", "config.toml", "Path to config file")
	flag.Parse()

	if *showVersion {
		fmt.Println("WireGuard Monitor version:", version)
		os.Exit(0)
	}

	log.Println("Starte WireGuard Monitor")

	loadConfig(*configPath)

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
