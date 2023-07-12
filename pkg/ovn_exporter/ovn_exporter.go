// Copyright 2018 Paul Greenberg (greenpau@outlook.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ovn_exporter

import (
	_ "net/http/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/greenpau/ovsdb"
	"github.com/greenpau/versioned"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/version"
)

const (
	namespace = "ovn"
)

var (
	app        *versioned.PackageManager
	appVersion string
	gitBranch  string
	gitCommit  string
	buildUser  string
	buildDate  string
)

func init() {
	app = versioned.NewPackageManager("ovn-exporter")
	app.Description = "Prometheus Exporter for Open Virtual Network (OVN)"
	app.Documentation = "https://github.com/greenpau/ovn_exporter/"
	app.SetVersion(appVersion, "1.0.6")
	app.SetGitBranch(gitBranch, "")
	app.SetGitCommit(gitCommit, "")
	app.SetBuildUser(buildUser, "")
	app.SetBuildDate(buildDate, "")
}

var (
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Is OVN stack up (1) or is it down (0).",
		nil, nil,
	)
	info = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "info"),
		"This metric provides basic information about OVN stack. It is always set to 1.",
		[]string{
			"system_id",
			"rundir",
			"hostname",
			"system_type",
			"system_version",
			"ovs_version",
			"db_version",
		}, nil,
	)
	requestErrors = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "failed_req_count"),
		"The number of failed requests to OVN stack.",
		[]string{"system_id"}, nil,
	)
	nextPoll = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "next_poll"),
		"The timestamp of the next potential poll of OVN stack.",
		[]string{"system_id"}, nil,
	)
	pid = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "pid"),
		"The process ID of a running OVN component. If the component is not running, then the ID is 0.",
		[]string{"system_id", "component", "user", "group"}, nil,
	)
	logFileSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "log_file_size"),
		"The size of a log file associated with an OVN component.",
		[]string{"system_id", "component", "filename"}, nil,
	)
	logEventStat = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "log_event_count"),
		"The number of recorded log meessage associated with an OVN component by log severity level and source.",
		[]string{"system_id", "component", "severity", "source"}, nil,
	)
	dbFileSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "db_file_size"),
		"The size of a database file associated with an OVN component.",
		[]string{"system_id", "component", "filename"}, nil,
	)
	chassisInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "chassis_info"),
		"Whether the OVN chassis is up (1) or down (0), together with additional information about the chassis.",
		[]string{"system_id", "uuid", "name", "ip"}, nil,
	)
	logicalSwitchInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "logical_switch_info"),
		"The information about OVN logical switch. This metric is always up (1).",
		[]string{"system_id", "uuid", "name"}, nil,
	)
	logicalSwitchExternalIDs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "logical_switch_external_id"),
		"Provides the external IDs and values associated with OVN logical switches. This metric is always up (1).",
		[]string{"system_id", "uuid", "key", "value"}, nil,
	)
	logicalSwitchPortBinding = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "logical_switch_port_binding"),
		"Provides the association between a logical switch and a logical switch port. This metric is always up (1).",
		[]string{"system_id", "uuid", "port"}, nil,
	)
	logicalSwitchTunnelKey = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "logical_switch_tunnel_key"),
		"The value of the tunnel key associated with the logical switch.",
		[]string{"system_id", "uuid"}, nil,
	)
	logicalSwitchPorts = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "logical_switch_ports"),
		"The number of logical switch ports connected to the OVN logical switch.",
		[]string{"system_id", "uuid"}, nil,
	)
	logicalSwitchPortInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "logical_switch_port_info"),
		"The information about OVN logical switch port. This metric is always up (1).",
		[]string{
			"system_id",
			"uuid",
			"name",
			"chassis",
			"logical_switch",
			"datapath",
			"port_binding",
			"mac_address",
			"ip_address",
		}, nil,
	)
	logicalSwitchPortTunnelKey = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "logical_switch_port_tunnel_key"),
		"The value of the tunnel key associated with the logical switch port.",
		[]string{"system_id", "uuid"}, nil,
	)
	networkPortUp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "network_port"),
		"The TCP port used for database connection. If the value is 0, then the port is not in use.",
		[]string{"system_id", "component", "usage"}, nil,
	)
	covAvg = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "coverage_avg"),
		"The average rate of the number of times particular events occur during a OVSDB daemon's runtime.",
		[]string{"system_id", "component", "event", "interval"}, nil,
	)
	covTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "coverage_total"),
		"The total number of times particular events occur during a OVSDB daemon's runtime.",
		[]string{"system_id", "component", "event"}, nil,
	)
	memUsage = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "memory_usage"),
		"The memory usage.",
		[]string{"system_id", "component", "facility"}, nil,
	)

	clusterEnabled = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_enabled"),
		"Is OVN clustering enabled (1) or not (0).",
		[]string{"system_id", "component"}, nil,
	)
	clusterPeerInConnInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_inbound_peer_connected"),
		"This metric appears when a cluster peer is connected to this server. This metric is always 1.",
		[]string{"system_id", "component", "server_id", "cluster_id", "peer_id", "peer_address"}, nil,
	)
	clusterPeerOutConnInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_outbound_peer_connected"),
		"This metric appears when this server connects to a cluster peer. This metric is always 1.",
		[]string{"system_id", "component", "server_id", "cluster_id", "peer_id", "peer_address"}, nil,
	)
	clusterPeerInConnTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_inbound_peer_conn_total"),
		"The total number of outbound connections to cluster peers.",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterPeerOutConnTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_outbound_peer_conn_total"),
		"The total number of inbound connections from cluster peers.",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterPeerCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_peer_count"),
		"The total number of peers in this server's cluster.",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterRole = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_role"),
		"The role of this server in the cluster. The values are: 3 - leader, 2 - candidate, 1 - follower, 0 - other.",
		[]string{"system_id", "component", "server_id", "server_uuid", "cluster_id", "cluster_uuid"}, nil,
	)
	clusterStatus = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_status"),
		"The status of this server in the cluster. The values are: 1 - cluster member, 0 - other.",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterTerm = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_term"),
		"The current raft term known by this server.",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterNotCommittedEntryCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_uncommitted_entry_count"),
		"The number of raft entries not yet committed by this server.",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterNotAppliedEntryCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_pending_entry_count"),
		"The number of raft entries not yet applied by this server.",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterPeerNextIndex = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_peer_next_index"),
		"The raft's next index associated with this cluster peer.",
		[]string{"system_id", "component", "server_id", "cluster_id", "peer_id"}, nil,
	)
	clusterPeerMatchIndex = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_peer_match_index"),
		"The raft's match index associated with this cluster peer.",
		[]string{"system_id", "component", "server_id", "cluster_id", "peer_id"}, nil,
	)
	clusterLeaderSelf = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_leader_self"),
		"Is this server consider itself a leader (1) or not (0).",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterVoteSelf = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_vote_self"),
		"Is this server voted itself as a leader (1) or not (0).",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterNextIndex = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_next_index"),
		"The raft's next index associated with this server.",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterMatchIndex = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_match_index"),
		"The raft's match index associated with this server.",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterLogLowIndex = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_log_low_index"),
		"The raft's low number associated with this server.",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterLogHighIndex = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_log_high_index"),
		"The raft's low number associated with this server.",
		[]string{"system_id", "component", "server_id", "cluster_id"}, nil,
	)
	clusterGroup = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "cluster_group"),
		"The cluster group in which this server participates. It is a combination of SB and NB cluster IDs. This metric is always up (1).",
		[]string{"system_id", "cluster_group"}, nil,
	)
)

// Exporter collects OVN data from the given server and exports them using
// the prometheus metrics package.
type Exporter struct {
	sync.RWMutex
	Client               *ovsdb.OvnClient
	timeout              int
	pollInterval         int64
	errors               int64
	errorsLocker         sync.RWMutex
	nextCollectionTicker int64
	metrics              []prometheus.Metric
	logger               log.Logger
	disableOvsdbServer 	 bool
	disableOvsVswitchd bool
	disableNorthd bool
	disableNorthbound bool
	disableSouthbound bool
}

type Options struct {
	Timeout int
	Logger  log.Logger
	DisableOvsdbServer bool
	DisableOvsVswitchd bool
	DisableNorthd bool
	DisableNorthbound bool
	DisableSouthbound bool
}

// NewLogger returns an instance of logger.
func NewLogger(logLevel string) (log.Logger, error) {
	allowedLogLevel := &promlog.AllowedLevel{}
	if err := allowedLogLevel.Set(logLevel); err != nil {
		return nil, err
	}
	promlogConfig := &promlog.Config{
		Level: allowedLogLevel,
	}
	logger := promlog.New(promlogConfig)
	return logger, nil
}

// NewExporter returns an initialized Exporter.
func NewExporter(opts Options) (*Exporter, error) {
	version.Version = appVersion
	version.Revision = gitCommit
	version.Branch = gitBranch
	version.BuildUser = buildUser
	version.BuildDate = buildDate
	e := Exporter{
		timeout: opts.Timeout,
		logger:  opts.Logger,
		disableOvsdbServer: opts.DisableOvsdbServer,
		disableOvsVswitchd: opts.DisableOvsVswitchd,
		disableNorthd: opts.DisableNorthd,
		disableNorthbound: opts.DisableNorthbound,
		disableSouthbound: opts.DisableSouthbound,
	}
	client := ovsdb.NewOvnClient()
	client.Timeout = opts.Timeout
	e.Client = client
	e.Client.GetSystemID()
	return &e, nil
}

// ExporterPerformClientCalls creates client connection.
func ExporterPerformClientCalls(e *Exporter) (*Exporter, error) {
	level.Debug(e.logger).Log(
		"msg", "NewExporter() calls Connect()",
		"system_id", e.Client.System.ID,
	)
	if err := e.Client.Connect(); err != nil {
		return e, err
	}
	level.Debug(e.logger).Log(
		"msg", "NewExporter() calls GetSystemInfo()",
		"system_id", e.Client.System.ID,
	)
	if err := e.Client.GetSystemInfo(); err != nil {
		return e, err
	}
	level.Debug(e.logger).Log(
		"msg", "NewExporter() initialized successfully",
		"system_id", e.Client.System.ID,
	)
	return e, nil
}

// Describe describes all the metrics ever exported by the OVN exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- info
	ch <- requestErrors
	ch <- nextPoll
	ch <- pid
	ch <- logFileSize
	ch <- dbFileSize
	ch <- logEventStat
	ch <- chassisInfo
	ch <- logicalSwitchInfo
	ch <- logicalSwitchExternalIDs
	ch <- logicalSwitchPorts
	ch <- logicalSwitchPortBinding
	ch <- logicalSwitchTunnelKey
	ch <- logicalSwitchPortInfo
	ch <- logicalSwitchPortTunnelKey
	ch <- networkPortUp
	ch <- covAvg
	ch <- covTotal
	ch <- memUsage
	ch <- clusterEnabled
	ch <- clusterRole
	ch <- clusterStatus
	ch <- clusterTerm
	ch <- clusterNotCommittedEntryCount
	ch <- clusterNotAppliedEntryCount
	ch <- clusterNextIndex
	ch <- clusterMatchIndex
	ch <- clusterLogLowIndex
	ch <- clusterLogHighIndex
	ch <- clusterLeaderSelf
	ch <- clusterVoteSelf
	ch <- clusterPeerCount
	ch <- clusterPeerInConnTotal
	ch <- clusterPeerOutConnTotal
	ch <- clusterPeerNextIndex
	ch <- clusterPeerMatchIndex
	ch <- clusterPeerInConnInfo
	ch <- clusterPeerOutConnInfo
	ch <- clusterGroup
}

// IncrementErrorCounter increases the counter of failed queries
// to OVN server.
func (e *Exporter) IncrementErrorCounter() {
	e.errorsLocker.Lock()
	defer e.errorsLocker.Unlock()
	atomic.AddInt64(&e.errors, 1)
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.GatherMetrics()
	level.Debug(e.logger).Log(
		"msg", "Collect() calls RLock()",
		"system_id", e.Client.System.ID,
	  )	  
	e.RLock()
	defer e.RUnlock()
	if len(e.metrics) == 0 {
		level.Debug(e.logger).Log(
			"msg", "Collect() no metrics found",
			"system_id", e.Client.System.ID,
		  )	  
		ch <- prometheus.MustNewConstMetric(
			up,
			prometheus.GaugeValue,
			0,
		)
		ch <- prometheus.MustNewConstMetric(
			info,
			prometheus.GaugeValue,
			1,
			e.Client.System.ID, e.Client.System.RunDir, e.Client.System.Hostname,
			e.Client.System.Type, e.Client.System.Version,
			e.Client.Database.Vswitch.Version, e.Client.Database.Vswitch.Schema.Version,
		)
		ch <- prometheus.MustNewConstMetric(
			requestErrors,
			prometheus.CounterValue,
			float64(e.errors),
			e.Client.System.ID,
		)
		ch <- prometheus.MustNewConstMetric(
			nextPoll,
			prometheus.CounterValue,
			float64(e.nextCollectionTicker),
			e.Client.System.ID,
		)
		return
	}
	level.Debug(e.logger).Log(
		"msg", "Collect() sends metrics to a shared channel",
		"system_id", e.Client.System.ID,
		"metric_count", len(e.metrics),
	)
	for _, m := range e.metrics {
		ch <- m
	}
}

// GatherMetrics collect data from OVN server and stores them
// as Prometheus metrics.
func (e *Exporter) GatherMetrics() {
	level.Debug(e.logger).Log(
		"msg", "GatherMetrics() called",
		"system_id", e.Client.System.ID,
	)
	if time.Now().Unix() < e.nextCollectionTicker {
		return
	}
	e.Lock()
	level.Debug(e.logger).Log(
		"msg", "GatherMetrics() locked",
		"system_id", e.Client.System.ID,
	)
	defer e.Unlock()
	if len(e.metrics) > 0 {
		e.metrics = e.metrics[:0]
		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() cleared metrics",
			"system_id", e.Client.System.ID,
		)
	}
	upValue := 1
	isClusterEnabled := false

	var err error

	err = e.Client.GetSystemInfo()
	if err != nil {
		level.Debug(e.logger).Log(
			"msg", "GetSystemInfo() failed",
			"vswitch_name", e.Client.Database.Vswitch.Name,
			"system_id", e.Client.System.ID,
			"error", err.Error(),
		)
		e.IncrementErrorCounter()
		upValue = 0
	} else {
		level.Debug(e.logger).Log(
			"msg", "GetSystemInfo() successful",
			"vswitch_name", e.Client.Database.Vswitch.Name,
			"system_id", e.Client.System.ID,
		)
	}

	components := []string{	}

	if (! e.disableOvsdbServer) { components = append(components, "ovsdb-server")	}
	if (! e.disableOvsVswitchd) { components = append(components, "ovs-vswitchd")	}
	if (! e.disableNorthd) { components = append(components, "ovn-northd") }
	if (! e.disableNorthd) { components = append(components, "ovn-northd-monitoring")	}
	if (! e.disableNorthbound) { components = append(components, "ovsdb-server-northbound") }
	if (! e.disableNorthbound) { components = append(components, "ovsdb-server-northbound-monitoring") }
	if (! e.disableSouthbound) { components = append(components, "ovsdb-server-southbound") }
	if (! e.disableSouthbound) { components = append(components, "ovsdb-server-southbound-monitoring") }

	for _, component := range components {
		p, err := e.Client.GetProcessInfo(component)
		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() calls GetProcessInfo()",
			"component", component,
			"system_id", e.Client.System.ID,
		)
		if err != nil {
			level.Error(e.logger).Log(
				"msg", "GetProcessInfo() failed",
				"component", component,
				"system_id", e.Client.System.ID,
				"error", err.Error(),
			)
			e.IncrementErrorCounter()
			upValue = 0
		}
		e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
			pid,
			prometheus.GaugeValue,
			float64(p.ID),
			e.Client.System.ID,
			component,
			p.User,
			p.Group,
		))
		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() completed GetProcessInfo()",
			"component", component,
			"system_id", e.Client.System.ID,
		)
	}
	
	components = []string{ }
	if (! e.disableOvsdbServer) { components = append(components, "ovsdb-server")	}
	if (! e.disableOvsVswitchd) { components = append(components, "ovs-vswitchd")	}
	if (! e.disableNorthd) { components = append(components, "ovn-northd") }
	if (! e.disableNorthbound) { components = append(components, "ovsdb-server-northbound") }
	if (! e.disableSouthbound) { components = append(components, "ovsdb-server-southbound") }

	for _, component := range components {
		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() calls GetLogFileInfo()",
			"component", component,
			"system_id", e.Client.System.ID,
		)
		file, err := e.Client.GetLogFileInfo(component)
		if err != nil {
			level.Error(e.logger).Log(
				"msg", "GetLogFileInfo() failed",
				"component", component,
				"system_id", e.Client.System.ID,
				"error", err.Error(),
			)
			e.IncrementErrorCounter()
			continue
		}
		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() completed GetLogFileInfo()",
			"component", component,
			"system_id", e.Client.System.ID,
		)
		e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
			logFileSize,
			prometheus.GaugeValue,
			float64(file.Info.Size()),
			e.Client.System.ID,
			file.Component,
			file.Path,
		))
		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() calls GetLogFileEventStats()",
			"component", component,
			"system_id", e.Client.System.ID,
		)
		eventStats, err := e.Client.GetLogFileEventStats(component)
		if err != nil {
			level.Error(e.logger).Log(
				"msg", "GetLogFileEventStats() failed",
				"component", component,
				"system_id", e.Client.System.ID,
				"error", err.Error(),
			)
			e.IncrementErrorCounter()
			continue
		}
		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() completed GetLogFileEventStats()",
			"component", component,
			"system_id", e.Client.System.ID,
		)
		for sev, sources := range eventStats {
			for source, count := range sources {
				e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
					logEventStat,
					prometheus.GaugeValue,
					float64(count),
					e.Client.System.ID,
					component,
					sev,
					source,
				))
			}
		}
	}

	level.Debug(e.logger).Log(
		"msg", "GatherMetrics() calls GetChassis()",
		"system_id", e.Client.System.ID,
	)
	if vteps, err := e.Client.GetChassis(); err != nil {
		level.Error(e.logger).Log(
			"msg", "GetChassis() failed",
			"southbound_db_name", e.Client.Database.Southbound.Name,
			"system_id", e.Client.System.ID,
			"error", err.Error(),
		)
		e.IncrementErrorCounter()
		upValue = 0
	} else {
		for _, vtep := range vteps {
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				chassisInfo,
				prometheus.GaugeValue,
				float64(vtep.Up),
				e.Client.System.ID,
				vtep.UUID,
				vtep.Name,
				vtep.IPAddress.String(),
			))
		}
	}
	level.Debug(e.logger).Log(
		"msg", "GatherMetrics() completed GetChassis()",
		"system_id", e.Client.System.ID,
	)

	level.Debug(e.logger).Log(
		"msg", "GatherMetrics() calls GetLogicalSwitches()",
		"system_id", e.Client.System.ID,
	)
	lsws, err := e.Client.GetLogicalSwitches()
	if err != nil {
		level.Error(e.logger).Log(
			"msg", "GetLogicalSwitches() failed",
			"southbound_db_name", e.Client.Database.Southbound.Name,
			"system_id", e.Client.System.ID,
			"error", err.Error(),
		)
		e.IncrementErrorCounter()
		upValue = 0
	} else {
		for _, lsw := range lsws {
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				logicalSwitchInfo,
				prometheus.GaugeValue,
				1,
				e.Client.System.ID,
				lsw.UUID,
				lsw.Name,
			))
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				logicalSwitchPorts,
				prometheus.GaugeValue,
				float64(len(lsw.Ports)),
				e.Client.System.ID,
				lsw.UUID,
			))
			if len(lsw.Ports) > 0 {
				for _, p := range lsw.Ports {
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						logicalSwitchPortBinding,
						prometheus.GaugeValue,
						1,
						e.Client.System.ID,
						lsw.UUID,
						p,
					))
				}
			}
			if len(lsw.ExternalIDs) > 0 {
				for k, v := range lsw.ExternalIDs {
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						logicalSwitchExternalIDs,
						prometheus.GaugeValue,
						1,
						e.Client.System.ID,
						lsw.UUID,
						k,
						v,
					))
				}
			}
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				logicalSwitchTunnelKey,
				prometheus.GaugeValue,
				float64(lsw.TunnelKey),
				e.Client.System.ID,
				lsw.UUID,
			))
		}
	}
	level.Debug(e.logger).Log(
		"msg", "GatherMetrics() completed GetLogicalSwitches()",
		"system_id", e.Client.System.ID,
	)

	level.Debug(e.logger).Log(
		"msg", "GatherMetrics() calls GetLogicalSwitchPorts()",
		"system_id", e.Client.System.ID,
	)
	lswps, err := e.Client.GetLogicalSwitchPorts()
	if err != nil {
		level.Error(e.logger).Log(
			"msg", "GetLogicalSwitchPorts() failed",
			"southbound_db_name", e.Client.Database.Southbound.Name,
			"system_id", e.Client.System.ID,
			"error", err.Error(),
		)
		e.IncrementErrorCounter()
		upValue = 0
	} else {
		for _, port := range lswps {
			macAddr := "<nil>"
			ipAddr := "<nil>"

			// Find first MAC address
			for _, a := range port.Addresses {
				if a.MacAddress != nil {
					macAddr = a.MacAddress.String()
					break
				}
			}

			// Find first IP address
			for _, a := range port.Addresses {
				if len(a.IpAddresses) > 0 {
					ipAddr = a.IpAddresses[0].String()
					break
				}
			}

			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				logicalSwitchPortInfo,
				prometheus.GaugeValue,
				float64(1),
				e.Client.System.ID,
				port.UUID,
				port.Name,
				port.ChassisUUID,
				port.LogicalSwitchName,
				port.DatapathUUID,
				port.PortBindingUUID,
				macAddr,
				ipAddr,
			))
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				logicalSwitchPortTunnelKey,
				prometheus.GaugeValue,
				float64(port.TunnelKey),
				e.Client.System.ID,
				port.UUID,
			))
		}
	}
	level.Debug(e.logger).Log(
		"msg", "GatherMetrics() completed GetLogicalSwitchPorts()",
		"system_id", e.Client.System.ID,
	)

	northClusterID := ""
	southClusterID := ""

	components = []string{ }
	if (! e.disableOvsdbServer) { components = append(components, "ovsdb-server")	}
	if (! e.disableNorthbound) { components = append(components, "ovsdb-server-northbound") }
	if (! e.disableSouthbound) { components = append(components, "ovsdb-server-southbound") }

	for _, component := range components {
		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() calls AppListCommands()",
			"component", component,
			"system_id", e.Client.System.ID,
		)
		if cmds, err := e.Client.AppListCommands(component); err != nil {
			level.Error(e.logger).Log(
				"msg", "AppListCommands() failed",
				"component", component,
				"system_id", e.Client.System.ID,
				"error", err.Error(),
			)
			e.IncrementErrorCounter()
			level.Debug(e.logger).Log(
				"msg", "GatherMetrics() completed AppListCommands()",
				"component", component,
				"system_id", e.Client.System.ID,
			)
		} else {
			level.Debug(e.logger).Log(
				"msg", "GatherMetrics() completed AppListCommands()",
				"component", component,
				"system_id", e.Client.System.ID,
			)
			if cmds["coverage/show"] {
				level.Debug(e.logger).Log(
					"msg", "GatherMetrics() calls GetAppCoverageMetrics()",
					"component", component,
					"system_id", e.Client.System.ID,
				)
				if metrics, err := e.Client.GetAppCoverageMetrics(component); err != nil {
					level.Error(e.logger).Log(
						"msg", "GetAppCoverageMetrics() failed",
						"component", component,
						"system_id", e.Client.System.ID,
						"error", err.Error(),
					)
					e.IncrementErrorCounter()
				} else {
					for event, metric := range metrics {
						//log.Infof("%s: %s, %s", component, name, metric)
						for period, value := range metric {
							if period == "total" {
								e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
									covTotal,
									prometheus.CounterValue,
									value,
									e.Client.System.ID,
									component,
									event,
								))
							} else {
								e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
									covAvg,
									prometheus.GaugeValue,
									value,
									e.Client.System.ID,
									component,
									event,
									period,
								))
							}
						}
					}
				}
				level.Debug(e.logger).Log(
					"msg", "GatherMetrics() completed GetAppCoverageMetrics()",
					"component", component,
					"system_id", e.Client.System.ID,
				)
			}
			if cmds["memory/show"] {
				level.Debug(e.logger).Log(
					"msg", "GatherMetrics() calls GetAppMemoryMetrics()",
					"component", component,
					"system_id", e.Client.System.ID,
				)
				if metrics, err := e.Client.GetAppMemoryMetrics(component); err != nil {
					level.Error(e.logger).Log(
						"msg", "GetAppMemoryMetrics() failed",
						"component", component,
						"system_id", e.Client.System.ID,
						"error", err.Error(),
					)
					e.IncrementErrorCounter()
				} else {
					for facility, value := range metrics {
						e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
							memUsage,
							prometheus.GaugeValue,
							value,
							e.Client.System.ID,
							component,
							facility,
						))
					}
				}
				level.Debug(e.logger).Log(
					"msg", "GatherMetrics() completed GetAppMemoryMetrics()",
					"component", component,
					"system_id", e.Client.System.ID,
				)
			}
			if cmds["cluster/status DB"] {
				level.Debug(e.logger).Log(
					"msg", "GatherMetrics() calls GetAppClusteringInfo()",
					"component", component,
					"system_id", e.Client.System.ID,
				)
				if cluster, err := e.Client.GetAppClusteringInfo(component); err != nil {
					isClusterEnabled = false
					level.Error(e.logger).Log(
						"msg", "GetAppClusteringInfo() failed",
						"component", component,
						"system_id", e.Client.System.ID,
						"error", err.Error(),
					)
					//e.IncrementErrorCounter()
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterEnabled,
						prometheus.GaugeValue,
						0,
						e.Client.System.ID,
						component,
					))
				} else {
					isClusterEnabled = true
					switch component {
					case "ovsdb-server-southbound":
						southClusterID = cluster.ClusterID
					case "ovsdb-server-northbound":
						northClusterID = cluster.ClusterID
					}
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterEnabled,
						prometheus.GaugeValue,
						1,
						e.Client.System.ID,
						component,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterRole,
						prometheus.GaugeValue,
						float64(cluster.Role),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.UUID,
						cluster.ClusterID,
						cluster.ClusterUUID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterStatus,
						prometheus.GaugeValue,
						float64(cluster.Status),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterTerm,
						prometheus.CounterValue,
						float64(cluster.Term),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterNotCommittedEntryCount,
						prometheus.GaugeValue,
						float64(cluster.NotCommittedEntries),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterNotAppliedEntryCount,
						prometheus.GaugeValue,
						float64(cluster.NotAppliedEntries),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterNextIndex,
						prometheus.CounterValue,
						float64(cluster.NextIndex),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterMatchIndex,
						prometheus.CounterValue,
						float64(cluster.MatchIndex),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterLogLowIndex,
						prometheus.CounterValue,
						float64(cluster.Log.Low),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterLogHighIndex,
						prometheus.CounterValue,
						float64(cluster.Log.High),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterLeaderSelf,
						prometheus.GaugeValue,
						float64(cluster.IsLeaderSelf),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterVoteSelf,
						prometheus.GaugeValue,
						float64(cluster.IsVotedSelf),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterPeerCount,
						prometheus.GaugeValue,
						float64(len(cluster.Peers)),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterPeerInConnTotal,
						prometheus.GaugeValue,
						float64(cluster.Connections.Inbound),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterPeerOutConnTotal,
						prometheus.GaugeValue,
						float64(cluster.Connections.Outbound),
						e.Client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					for peerID, peer := range cluster.Peers {
						e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
							clusterPeerNextIndex,
							prometheus.CounterValue,
							float64(peer.NextIndex),
							e.Client.System.ID,
							component,
							cluster.ID,
							cluster.ClusterID,
							peerID,
						))
						e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
							clusterPeerMatchIndex,
							prometheus.CounterValue,
							float64(peer.MatchIndex),
							e.Client.System.ID,
							component,
							cluster.ID,
							cluster.ClusterID,
							peerID,
						))
						e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
							clusterPeerInConnInfo,
							prometheus.GaugeValue,
							float64(peer.Connection.Inbound),
							e.Client.System.ID,
							component,
							cluster.ID,
							cluster.ClusterID,
							peerID,
							peer.Address,
						))
						e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
							clusterPeerOutConnInfo,
							prometheus.GaugeValue,
							float64(peer.Connection.Outbound),
							e.Client.System.ID,
							component,
							cluster.ID,
							cluster.ClusterID,
							peerID,
							peer.Address,
						))
					}
					//log.Infof("%s: %v", component, cluster)
				}
				level.Debug(e.logger).Log(
					"msg", "GatherMetrics() completed GetAppClusteringInfo()",
					"component", component,
					"system_id", e.Client.System.ID,
				)
			}
		}
	}

	if northClusterID != "" && southClusterID != "" {
		e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
			clusterGroup,
			prometheus.GaugeValue,
			1,
			e.Client.System.ID,
			northClusterID+southClusterID,
		))
	}

	components = []string{ }
	if (! e.disableNorthbound) { components = append(components, "ovsdb-server-northbound") }
	if (! e.disableSouthbound) { components = append(components, "ovsdb-server-southbound") }

	for _, component := range components {
		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() calls IsDefaultPortUp()",
			"component", component,
			"system_id", e.Client.System.ID,
		)
		defaultPortUp, err := e.Client.IsDefaultPortUp(component)
		if err != nil {
			level.Error(e.logger).Log(
				"msg", "IsDefaultPortUp() failed",
				"component", component,
				"system_id", e.Client.System.ID,
				"error", err.Error(),
			)
			e.IncrementErrorCounter()
		}
		e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
			networkPortUp,
			prometheus.GaugeValue,
			float64(defaultPortUp),
			e.Client.System.ID,
			component,
			"default",
		))
		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() completed IsDefaultPortUp()",
			"component", component,
			"system_id", e.Client.System.ID,
		)

		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() calls IsSslPortUp()",
			"component", component,
			"system_id", e.Client.System.ID,
		)
		sslPortUp, err := e.Client.IsSslPortUp(component)
		if err != nil {
			level.Error(e.logger).Log(
				"msg", "IsSslPortUp() failed",
				"component", component,
				"system_id", e.Client.System.ID,
				"error", err.Error(),
			)
			e.IncrementErrorCounter()
		}
		e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
			networkPortUp,
			prometheus.GaugeValue,
			float64(sslPortUp),
			e.Client.System.ID,
			component,
			"ssl",
		))
		level.Debug(e.logger).Log(
			"msg", "GatherMetrics() completed IsSslPortUp()",
			"component", component,
			"system_id", e.Client.System.ID,
		)

		if isClusterEnabled {
			level.Debug(e.logger).Log(
				"msg", "GatherMetrics() calls IsRaftPortUp()",
				"component", component,
				"system_id", e.Client.System.ID,
			)
			raftPortUp, err := e.Client.IsRaftPortUp(component)
			if err != nil {
				level.Error(e.logger).Log(
					"msg", "IsRaftPortUp() failed",
					"component", component,
					"system_id", e.Client.System.ID,
					"error", err.Error(),
				)
				e.IncrementErrorCounter()
			}
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				networkPortUp,
				prometheus.GaugeValue,
				float64(raftPortUp),
				e.Client.System.ID,
				component,
				"raft",
			))
			level.Debug(e.logger).Log(
				"msg", "GatherMetrics() completed IsRaftPortUp()",
				"component", component,
				"system_id", e.Client.System.ID,
			)
		}
	}

	e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
		up,
		prometheus.GaugeValue,
		float64(upValue),
	))

	e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
		info,
		prometheus.GaugeValue,
		1,
		e.Client.System.ID, e.Client.System.RunDir, e.Client.System.Hostname,
		e.Client.System.Type, e.Client.System.Version,
		e.Client.Database.Vswitch.Version, e.Client.Database.Vswitch.Schema.Version,
	))

	e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
		requestErrors,
		prometheus.CounterValue,
		float64(e.errors),
		e.Client.System.ID,
	))

	e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
		nextPoll,
		prometheus.CounterValue,
		float64(e.nextCollectionTicker),
		e.Client.System.ID,
	))

	e.nextCollectionTicker = time.Now().Add(time.Duration(e.pollInterval) * time.Second).Unix()
	
	level.Debug(e.logger).Log(
		"msg", "GatherMetrics() returns",
		"system_id", e.Client.System.ID,
	)
	return
}

func init() {
	prometheus.MustRegister(version.NewCollector(namespace + "_exporter"))
}

// GetVersionInfo returns exporter info.
func GetVersionInfo() string {
	return version.Info()
}

// GetVersionBuildContext returns exporter build context.
func GetVersionBuildContext() string {
	return version.BuildContext()
}

// GetVersion returns exporter version.
func GetVersion() string {
	return version.Version
}

// GetRevision returns exporter revision.
func GetRevision() string {
	return version.Revision
}

// GetExporterName returns exporter name.
func GetExporterName() string {
	return app.Name
}

// GetExporterVersion return full version info.
func GetExporterVersion() string {
	return app.Banner()
}

// SetPollInterval sets exporter's polling interval.
func (e *Exporter) SetPollInterval(i int64) {
	e.pollInterval = i
}
