package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/log/level"
	ovn "github.com/greenpau/ovn_exporter/pkg/ovn_exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var listenAddress string
	var metricsPath string
	var pollTimeout int
	var pollInterval int
	var isShowVersion bool
	var logLevel string
	var systemRunDir string
	var disableOvsdbServer bool
	var disableOvsVswitchd bool
	var disableNorthd bool
	var disableNorthbound bool
	var disableSouthbound bool
	var databaseVswitchName string
	var databaseVswitchSocketRemote string
	var databaseVswitchFileDataPath string
	var databaseVswitchFileLogPath string
	var databaseVswitchFilePidPath string
	var databaseVswitchFileSystemIDPath string
	var databaseNorthboundName string
	var databaseNorthboundSocketRemote string
	var databaseNorthboundSocketControl string
	var databaseNorthboundFileDataPath string
	var databaseNorthboundFileLogPath string
	var databaseNorthboundFilePidPath string
	var databaseNorthboundPortDefault int
	var databaseNorthboundPortSsl int
	var databaseNorthboundPortRaft int
	var databaseSouthboundName string
	var databaseSouthboundSocketRemote string
	var databaseSouthboundSocketControl string
	var databaseSouthboundFileDataPath string
	var databaseSouthboundFileLogPath string
	var databaseSouthboundFilePidPath string
	var databaseSouthboundPortDefault int
	var databaseSouthboundPortSsl int
	var databaseSouthboundPortRaft int
	var serviceVswitchdFileLogPath string
	var serviceVswitchdFilePidPath string
	var serviceNorthdFileLogPath string
	var serviceNorthdFilePidPath string

	flag.StringVar(&listenAddress, "web.listen-address", ":9476", "Address to listen on for web interface and telemetry.")
	flag.StringVar(&metricsPath, "web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	flag.IntVar(&pollTimeout, "ovn.timeout", 2, "Timeout on gRPC requests to OVN.")
	flag.IntVar(&pollInterval, "ovn.poll-interval", 15, "The minimum interval (in seconds) between collections from OVN server.")
	flag.BoolVar(&isShowVersion, "version", false, "version information")
	flag.StringVar(&logLevel, "log.level", "info", "logging severity level")

	flag.StringVar(&systemRunDir, "system.run.dir", "/var/run/openvswitch", "OVS default run directory.")

	flag.StringVar(&databaseVswitchName, "database.vswitch.name", "Open_vSwitch", "The name of OVS db.")
	flag.StringVar(&databaseVswitchSocketRemote, "database.vswitch.socket.remote", "unix:/var/run/openvswitch/db.sock", "JSON-RPC unix socket to OVS db.")
	flag.StringVar(&databaseVswitchFileDataPath, "database.vswitch.file.data.path", "/etc/openvswitch/conf.db", "OVS db file.")
	flag.StringVar(&databaseVswitchFileLogPath, "database.vswitch.file.log.path", "/var/log/openvswitch/ovsdb-server.log", "OVS db log file.")
	flag.StringVar(&databaseVswitchFilePidPath, "database.vswitch.file.pid.path", "/var/run/openvswitch/ovsdb-server.pid", "OVS db process id file.")
	flag.StringVar(&databaseVswitchFileSystemIDPath, "database.vswitch.file.system.id.path", "/etc/openvswitch/system-id.conf", "OVS system id file.")

	flag.StringVar(&databaseNorthboundName, "database.northbound.name", "OVN_Northbound", "The name of OVN NB (northbound) db.")
	flag.StringVar(&databaseNorthboundSocketRemote, "database.northbound.socket.remote", "unix:/run/openvswitch/ovnnb_db.sock", "JSON-RPC unix socket to OVN NB db.")
	flag.StringVar(&databaseNorthboundSocketControl, "database.northbound.socket.control", "unix:/run/openvswitch/ovnnb_db.ctl", "JSON-RPC unix socket to OVN NB app.")
	flag.StringVar(&databaseNorthboundFileDataPath, "database.northbound.file.data.path", "/var/lib/openvswitch/ovnnb_db.db", "OVN NB db file.")
	flag.StringVar(&databaseNorthboundFileLogPath, "database.northbound.file.log.path", "/var/log/openvswitch/ovsdb-server-nb.log", "OVN NB db log file.")
	flag.StringVar(&databaseNorthboundFilePidPath, "database.northbound.file.pid.path", "/run/openvswitch/ovnnb_db.pid", "OVN NB db process id file.")
	flag.IntVar(&databaseNorthboundPortDefault, "database.northbound.port.default", 6641, "OVN NB db network socket port.")
	flag.IntVar(&databaseNorthboundPortSsl, "database.northbound.port.ssl", 6631, "OVN NB db network socket secure port.")
	flag.IntVar(&databaseNorthboundPortRaft, "database.northbound.port.raft", 6643, "OVN NB db network port for clustering (raft)")

	flag.StringVar(&databaseSouthboundName, "database.southbound.name", "OVN_Southbound", "The name of OVN SB (southbound) db.")
	flag.StringVar(&databaseSouthboundSocketRemote, "database.southbound.socket.remote", "unix:/run/openvswitch/ovnsb_db.sock", "JSON-RPC unix socket to OVN SB db.")
	flag.StringVar(&databaseSouthboundSocketControl, "database.southbound.socket.control", "unix:/run/openvswitch/ovnsb_db.ctl", "JSON-RPC unix socket to OVN SB app.")
	flag.StringVar(&databaseSouthboundFileDataPath, "database.southbound.file.data.path", "/var/lib/openvswitch/ovnsb_db.db", "OVN SB db file.")
	flag.StringVar(&databaseSouthboundFileLogPath, "database.southbound.file.log.path", "/var/log/openvswitch/ovsdb-server-sb.log", "OVN SB db log file.")
	flag.StringVar(&databaseSouthboundFilePidPath, "database.southbound.file.pid.path", "/run/openvswitch/ovnsb_db.pid", "OVN SB db process id file.")
	flag.IntVar(&databaseSouthboundPortDefault, "database.southbound.port.default", 6642, "OVN SB db network socket port.")
	flag.IntVar(&databaseSouthboundPortSsl, "database.southbound.port.ssl", 6632, "OVN SB db network socket secure port.")
	flag.IntVar(&databaseSouthboundPortRaft, "database.southbound.port.raft", 6644, "OVN SB db network port for clustering (raft)")

	flag.StringVar(&serviceVswitchdFileLogPath, "service.vswitchd.file.log.path", "/var/log/openvswitch/ovs-vswitchd.log", "OVS vswitchd daemon log file.")
	flag.StringVar(&serviceVswitchdFilePidPath, "service.vswitchd.file.pid.path", "/var/run/openvswitch/ovs-vswitchd.pid", "OVS vswitchd daemon process id file.")

	flag.StringVar(&serviceNorthdFileLogPath, "service.ovn.northd.file.log.path", "/var/log/openvswitch/ovn-northd.log", "OVN northd daemon log file.")
	flag.StringVar(&serviceNorthdFilePidPath, "service.ovn.northd.file.pid.path", "/run/openvswitch/ovn-northd.pid", "OVN northd daemon process id file.")

	flag.BoolVar(&disableNorthd, "disable.northd", false, "Disable northd")
	flag.BoolVar(&disableNorthbound, "disable.ovsdb.server.northbound", false, "Disable ovsdb-server-northbound")
	flag.BoolVar(&disableSouthbound, "disable.ovsdb.server.southbound", false, "Disable ovsdb-server-southbound")
	flag.BoolVar(&disableOvsdbServer, "disable.ovsdb.server", false, "Disable ovsdb-server")
	flag.BoolVar(&disableOvsVswitchd, "disable.ovs.vswitchd", false, "Disable ovs-vswitchd")

	var usageHelp = func() {
		fmt.Fprintf(os.Stderr, "\n%s - Prometheus Exporter for Open Virtual Network (OVN)\n\n", ovn.GetExporterName())
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments]\n\n", ovn.GetExporterName())
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDocumentation: https://github.com/greenpau/ovn_exporter/\n\n")
	}
	flag.Usage = usageHelp
	flag.Parse()

	if isShowVersion {
		fmt.Fprintf(os.Stdout, "%s %s", ovn.GetExporterName(), ovn.GetVersion())
		if ovn.GetRevision() != "" {
			fmt.Fprintf(os.Stdout, ", commit: %s\n", ovn.GetRevision())
		} else {
			fmt.Fprint(os.Stdout, "\n")
		}
		os.Exit(0)
	}

	logger, err := ovn.NewLogger(logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed initializing logger: %v", err)
		os.Exit(1)
	}

	level.Info(logger).Log(
		"msg", "Starting exporter",
		"exporter", ovn.GetExporterName(),
		"version", ovn.GetVersionInfo(),
		"build_context", ovn.GetVersionBuildContext(),
	)

	opts := ovn.Options{
		Timeout: pollTimeout,
		Logger:  logger,
		DisableOvsdbServer: disableOvsdbServer,
		DisableOvsVswitchd: disableOvsVswitchd,
		DisableNorthd: disableNorthd,
		DisableNorthbound: disableNorthbound,
		DisableSouthbound: disableSouthbound,
	}

	exporter, err := ovn.NewExporter(opts)
	if err != nil {
		level.Error(logger).Log(
			"msg", "failed to init properly",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	exporter.Client.System.RunDir = systemRunDir

	exporter.Client.Database.Vswitch.Name = databaseVswitchName
	exporter.Client.Database.Vswitch.Socket.Remote = databaseVswitchSocketRemote
	exporter.Client.Database.Vswitch.File.Data.Path = databaseVswitchFileDataPath
	exporter.Client.Database.Vswitch.File.Log.Path = databaseVswitchFileLogPath
	exporter.Client.Database.Vswitch.File.Pid.Path = databaseVswitchFilePidPath
	exporter.Client.Database.Vswitch.File.SystemID.Path = databaseVswitchFileSystemIDPath

	exporter.Client.Database.Northbound.Name = databaseNorthboundName
	exporter.Client.Database.Northbound.Socket.Remote = databaseNorthboundSocketRemote
	exporter.Client.Database.Northbound.Socket.Control = databaseNorthboundSocketControl
	exporter.Client.Database.Northbound.File.Data.Path = databaseNorthboundFileDataPath
	exporter.Client.Database.Northbound.File.Log.Path = databaseNorthboundFileLogPath
	exporter.Client.Database.Northbound.File.Pid.Path = databaseNorthboundFilePidPath
	exporter.Client.Database.Northbound.Port.Default = databaseNorthboundPortDefault
	exporter.Client.Database.Northbound.Port.Ssl = databaseNorthboundPortSsl
	exporter.Client.Database.Northbound.Port.Raft = databaseNorthboundPortRaft

	exporter.Client.Database.Southbound.Name = databaseSouthboundName
	exporter.Client.Database.Southbound.Socket.Remote = databaseSouthboundSocketRemote
	exporter.Client.Database.Southbound.Socket.Control = databaseSouthboundSocketControl
	exporter.Client.Database.Southbound.File.Data.Path = databaseSouthboundFileDataPath
	exporter.Client.Database.Southbound.File.Log.Path = databaseSouthboundFileLogPath
	exporter.Client.Database.Southbound.File.Pid.Path = databaseSouthboundFilePidPath
	exporter.Client.Database.Southbound.Port.Default = databaseSouthboundPortDefault
	exporter.Client.Database.Southbound.Port.Ssl = databaseSouthboundPortSsl
	exporter.Client.Database.Southbound.Port.Raft = databaseSouthboundPortRaft

	exporter.Client.Service.Vswitchd.File.Log.Path = serviceVswitchdFileLogPath
	exporter.Client.Service.Vswitchd.File.Pid.Path = serviceVswitchdFilePidPath

	exporter.Client.Service.Northd.File.Log.Path = serviceNorthdFileLogPath
	exporter.Client.Service.Northd.File.Pid.Path = serviceNorthdFilePidPath

	exporter, err = ovn.ExporterPerformClientCalls(exporter)
	if err != nil {
		level.Error(logger).Log(
			"msg", "failed to finalize exporter calls properly",
			"exporter_name", ovn.GetExporterName(),
			"error", err.Error(),
		)
	}

	level.Info(logger).Log("ovs_system_id", exporter.Client.System.ID)

	exporter.SetPollInterval(int64(pollInterval))
	prometheus.MustRegister(exporter)

	http.Handle(metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>OVN Exporter</title></head>
             <body>
             <h1>OVN Exporter</h1>
             <p><a href='` + metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	level.Info(logger).Log("listen_on ", listenAddress)
	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		level.Error(logger).Log(
			"msg", "listener failed",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}
