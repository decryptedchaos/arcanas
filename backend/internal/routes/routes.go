package routes

import (
	"net/http"

	"arcanas/internal/handlers"
)

func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Hello endpoint (existing)
	mux.HandleFunc("/api/hello", handlers.Hello)

	// Disk storage endpoints
	mux.HandleFunc("/api/disk-stats", handlers.GetDiskStats)
	mux.HandleFunc("/api/disk/smart", handlers.GetSmartStatus)
	mux.HandleFunc("/api/disk/partitions", handlers.GetPartitions)

	// SCSI targets endpoints
	mux.HandleFunc("/api/scsi-targets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetSCSITargets(w, r)
		case http.MethodPost:
			handlers.CreateSCSITarget(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/scsi-targets/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handlers.UpdateSCSITarget(w, r)
		case http.MethodDelete:
			handlers.DeleteSCSITarget(w, r)
		case http.MethodPost:
			handlers.ToggleSCSITarget(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/scsi-targets/sessions", handlers.GetSCSISessions)

	// Samba shares endpoints
	mux.HandleFunc("/api/samba-shares", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetSambaShares(w, r)
		case http.MethodPost:
			handlers.CreateSambaShare(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/samba-shares/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handlers.UpdateSambaShare(w, r)
		case http.MethodDelete:
			handlers.DeleteSambaShare(w, r)
		case http.MethodPost:
			handlers.ToggleSambaShare(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/samba-shares/connections", handlers.GetSambaConnections)
	mux.HandleFunc("/api/samba-shares/test", handlers.TestSambaConfig)

	// NFS exports endpoints
	mux.HandleFunc("/api/nfs-exports", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetNFSExports(w, r)
		case http.MethodPost:
			handlers.CreateNFSExport(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/nfs-exports/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handlers.UpdateNFSExport(w, r)
		case http.MethodDelete:
			handlers.DeleteNFSExport(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/nfs-exports/status", handlers.GetNFSExportStatus)
	mux.HandleFunc("/api/nfs-exports/reload", handlers.ReloadNFSConfig)

	// System statistics endpoints
	mux.HandleFunc("/api/system/overview", handlers.GetSystemStats)
	mux.HandleFunc("/api/system/cpu", handlers.GetCPUStats)
	mux.HandleFunc("/api/system/memory", handlers.GetMemoryStats)
	mux.HandleFunc("/api/system/network", handlers.GetNetworkStats)
	mux.HandleFunc("/api/system/storage-health", handlers.GetStorageHealth)
	mux.HandleFunc("/api/system/processes", handlers.GetSystemProcesses)
	mux.HandleFunc("/api/system/logs", handlers.GetSystemLogs)
	mux.HandleFunc("/api/system/reboot", handlers.RebootSystem)
	mux.HandleFunc("/api/system/shutdown", handlers.ShutdownSystem)

	// I/O rate endpoints
	mux.HandleFunc("/api/system/disk-io", handlers.GetDiskIORates)
	mux.HandleFunc("/api/system/network-io", handlers.GetNetworkIORates)

	return mux
}
