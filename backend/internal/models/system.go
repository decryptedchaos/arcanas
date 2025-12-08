package models

import "time"

type SystemStats struct {
	CPU     CPUStats     `json:"cpu"`
	Memory  MemoryStats  `json:"memory"`
	Network NetworkStats `json:"network"`
	Storage StorageStats `json:"storage"`
	System  SystemInfo   `json:"system"`
}

type CPUStats struct {
	Usage       float64     `json:"usage"`
	Cores       int         `json:"cores"`
	Model       string      `json:"model"`
	Frequency   string      `json:"frequency"`
	Temperature float64     `json:"temperature"`
	LoadAverage []float64   `json:"load_average"`
	Processes   ProcessInfo `json:"processes"`
}

type ProcessInfo struct {
	Total    int `json:"total"`
	Running  int `json:"running"`
	Sleeping int `json:"sleeping"`
}

type MemoryStats struct {
	Total     int64    `json:"total"`
	Used      int64    `json:"used"`
	Available int64    `json:"available"`
	Usage     float64  `json:"usage"`
	Swap      SwapInfo `json:"swap"`
}

type SwapInfo struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
}

type NetworkStats struct {
	Interfaces []NetworkInterface `json:"interfaces"`
	TotalRx    int64              `json:"total_rx"`
	TotalTx    int64              `json:"total_tx"`
	RxRate     int64              `json:"rx_rate"`
	TxRate     int64              `json:"tx_rate"`
}

type NetworkInterface struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Speed   string `json:"speed"`
	Rx      int64  `json:"rx"`
	Tx      int64  `json:"tx"`
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
}

type StorageStats struct {
	Disks []DiskHealth `json:"disks"`
}

type DiskHealth struct {
	Device      string  `json:"device"`
	Model       string  `json:"model"`
	Size        int64   `json:"size"`
	Used        int64   `json:"used"`
	Temperature float64 `json:"temperature"`
	Health      int     `json:"health"`
	SmartStatus string  `json:"smart_status"`
}

type SystemInfo struct {
	Hostname     string    `json:"hostname"`
	Uptime       int64     `json:"uptime"`
	OS           string    `json:"os"`
	Kernel       string    `json:"kernel"`
	Architecture string    `json:"architecture"`
	LastBoot     time.Time `json:"last_boot"`
}
