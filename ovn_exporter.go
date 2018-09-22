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

package main

import (
	"flag"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/greenpau/ovsdb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	namespace = "ovn"
)

var (
	appName    = "ovn-exporter"
	appVersion = "[untracked]"
	gitBranch  string
	gitCommit  string
	buildUser  string // whoami
	buildDate  string // date -u
)

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
	client               *ovsdb.OvnClient
	timeout              int
	pollInterval         int64
	errors               int64
	errorsLocker         sync.RWMutex
	nextCollectionTicker int64
	metrics              []prometheus.Metric
}

type exporterOpts struct {
	timeout int
}

// NewExporter returns an initialized Exporter.
func NewExporter(opts exporterOpts) (*Exporter, error) {
	e := Exporter{
		timeout: opts.timeout,
	}
	client := ovsdb.NewOvnClient()
	client.Timeout = opts.timeout
	e.client = client
	e.client.GetSystemID()
	log.Debugf("%s: NewExporter() calls Connect()", e.client.System.ID)
	if err := client.Connect(); err != nil {
		return &e, err
	}
	log.Debugf("%s: NewExporter() calls GetSystemInfo()", e.client.System.ID)
	if err := e.client.GetSystemInfo(); err != nil {
		return &e, err
	}
	log.Debugf("%s: NewExporter() initialized successfully", e.client.System.ID)
	return &e, nil
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
	log.Debugf("%s: Collect() calls RLock()", e.client.System.ID)
	e.RLock()
	defer e.RUnlock()
	if len(e.metrics) == 0 {
		log.Debugf("%s: Collect() no metrics found", e.client.System.ID)
		ch <- prometheus.MustNewConstMetric(
			up,
			prometheus.GaugeValue,
			0,
		)
		ch <- prometheus.MustNewConstMetric(
			info,
			prometheus.GaugeValue,
			1,
			e.client.System.ID, e.client.System.RunDir, e.client.System.Hostname,
			e.client.System.Type, e.client.System.Version,
			e.client.Database.Vswitch.Version, e.client.Database.Vswitch.Schema.Version,
		)
		ch <- prometheus.MustNewConstMetric(
			requestErrors,
			prometheus.CounterValue,
			float64(e.errors),
			e.client.System.ID,
		)
		ch <- prometheus.MustNewConstMetric(
			nextPoll,
			prometheus.CounterValue,
			float64(e.nextCollectionTicker),
			e.client.System.ID,
		)
		return
	}
	log.Debugf("%s: Collect() sends %d metrics to a shared channel", e.client.System.ID, len(e.metrics))
	for _, m := range e.metrics {
		ch <- m
	}
}

// GatherMetrics collect data from OVN server and stores them
// as Prometheus metrics.
func (e *Exporter) GatherMetrics() {
	log.Debugf("%s: GatherMetrics() called", e.client.System.ID)
	if time.Now().Unix() < e.nextCollectionTicker {
		return
	}
	e.Lock()
	log.Debugf("%s: GatherMetrics() locked", e.client.System.ID)
	defer e.Unlock()
	if len(e.metrics) > 0 {
		e.metrics = e.metrics[:0]
		log.Debugf("%s: GatherMetrics() cleared metrics", e.client.System.ID)
	}
	upValue := 1
	isClusterEnabled := false

	var err error

	err = e.client.GetSystemInfo()
	if err != nil {
		log.Errorf("%s: %v", e.client.Database.Vswitch.Name, err)
		e.IncrementErrorCounter()
		upValue = 0
	} else {
		log.Debugf("%s: system-id: %s", e.client.Database.Vswitch.Name, e.client.System.ID)
	}

	components := []string{
		"ovsdb-server",
		"ovsdb-server-southbound",
		"ovsdb-server-southbound-monitoring",
		"ovsdb-server-northbound",
		"ovsdb-server-northbound-monitoring",
		"ovn-northd",
		"ovn-northd-monitoring",
		"ovs-vswitchd",
	}
	for _, component := range components {
		p, err := e.client.GetProcessInfo(component)
		log.Debugf("%s: GatherMetrics() calls GetProcessInfo(%s)", e.client.System.ID, component)
		if err != nil {
			log.Errorf("%s: pid-%v", component, err)
			e.IncrementErrorCounter()
			upValue = 0
		}
		e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
			pid,
			prometheus.GaugeValue,
			float64(p.ID),
			e.client.System.ID,
			component,
			p.User,
			p.Group,
		))
		log.Debugf("%s: GatherMetrics() completed GetProcessInfo(%s)", e.client.System.ID, component)
	}

	components = []string{
		"ovsdb-server",
		"ovsdb-server-southbound",
		"ovsdb-server-northbound",
		"ovn-northd",
		"ovs-vswitchd",
	}
	for _, component := range components {
		log.Debugf("%s: GatherMetrics() calls GetLogFileInfo(%s)", e.client.System.ID, component)
		file, err := e.client.GetLogFileInfo(component)
		if err != nil {
			log.Errorf("%s: log-file-%v", component, err)
			e.IncrementErrorCounter()
			continue
		}
		log.Debugf("%s: GatherMetrics() completed GetLogFileInfo(%s)", e.client.System.ID, component)
		e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
			logFileSize,
			prometheus.GaugeValue,
			float64(file.Info.Size()),
			e.client.System.ID,
			file.Component,
			file.Path,
		))
		log.Debugf("%s: GatherMetrics() calls GetLogFileEventStats(%s)", e.client.System.ID, component)
		eventStats, err := e.client.GetLogFileEventStats(component)
		if err != nil {
			log.Errorf("%s: log-event-stat: %v", component, err)
			e.IncrementErrorCounter()
			continue
		}
		log.Debugf("%s: GatherMetrics() completed GetLogFileEventStats(%s)", e.client.System.ID, component)
		for sev, sources := range eventStats {
			for source, count := range sources {
				e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
					logEventStat,
					prometheus.GaugeValue,
					float64(count),
					e.client.System.ID,
					component,
					sev,
					source,
				))
			}
		}
	}

	log.Debugf("%s: GatherMetrics() calls GetChassis()", e.client.System.ID)
	if vteps, err := e.client.GetChassis(); err != nil {
		log.Errorf("%s: %v", e.client.Database.Southbound.Name, err)
		e.IncrementErrorCounter()
		upValue = 0
	} else {
		for _, vtep := range vteps {
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				chassisInfo,
				prometheus.GaugeValue,
				float64(vtep.Up),
				e.client.System.ID,
				vtep.UUID,
				vtep.Name,
				vtep.IPAddress,
			))
		}
	}
	log.Debugf("%s: GatherMetrics() completed GetChassis()", e.client.System.ID)

	log.Debugf("%s: GatherMetrics() calls GetLogicalSwitches()", e.client.System.ID)
	lsws, err := e.client.GetLogicalSwitches()
	if err != nil {
		log.Errorf("%s: %v", e.client.Database.Southbound.Name, err)
		e.IncrementErrorCounter()
		upValue = 0
	} else {
		for _, lsw := range lsws {
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				logicalSwitchInfo,
				prometheus.GaugeValue,
				1,
				e.client.System.ID,
				lsw.UUID,
				lsw.Name,
			))
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				logicalSwitchPorts,
				prometheus.GaugeValue,
				float64(len(lsw.Ports)),
				e.client.System.ID,
				lsw.UUID,
			))
			if len(lsw.Ports) > 0 {
				for _, p := range lsw.Ports {
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						logicalSwitchPortBinding,
						prometheus.GaugeValue,
						1,
						e.client.System.ID,
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
						e.client.System.ID,
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
				e.client.System.ID,
				lsw.UUID,
			))
		}
	}
	log.Debugf("%s: GatherMetrics() completed GetLogicalSwitches()", e.client.System.ID)

	log.Debugf("%s: GatherMetrics() calls GetLogicalSwitchPorts()", e.client.System.ID)
	lswps, err := e.client.GetLogicalSwitchPorts()
	if err != nil {
		log.Errorf("%s: %v", e.client.Database.Southbound.Name, err)
		e.IncrementErrorCounter()
		upValue = 0
	} else {
		for _, port := range lswps {
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				logicalSwitchPortInfo,
				prometheus.GaugeValue,
				float64(1),
				e.client.System.ID,
				port.UUID,
				port.Name,
				port.ChassisUUID,
				port.LogicalSwitchName,
				port.DatapathUUID,
				port.PortBindingUUID,
				port.MacAddress.String(),
				port.IPAddress.String(),
			))
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				logicalSwitchPortTunnelKey,
				prometheus.GaugeValue,
				float64(port.TunnelKey),
				e.client.System.ID,
				port.UUID,
			))
		}
	}
	log.Debugf("%s: GatherMetrics() completed GetLogicalSwitchPorts()", e.client.System.ID)

	northClusterID := ""
	southClusterID := ""

	components = []string{
		"ovsdb-server-southbound",
		"ovsdb-server-northbound",
	}

	for _, component := range components {
		log.Debugf("%s: GatherMetrics() calls AppListCommands(%s)", e.client.System.ID, component)
		if cmds, err := e.client.AppListCommands(component); err != nil {
			log.Errorf("%s: %v", component, err)
			e.IncrementErrorCounter()
			log.Debugf("%s: GatherMetrics() completed AppListCommands(%s)", e.client.System.ID, component)
		} else {
			log.Debugf("%s: GatherMetrics() completed AppListCommands(%s)", e.client.System.ID, component)
			if cmds["coverage/show"] {
				log.Debugf("%s: GatherMetrics() calls GetAppCoverageMetrics(%s)", e.client.System.ID, component)
				if metrics, err := e.client.GetAppCoverageMetrics(component); err != nil {
					log.Errorf("%s: %v", component, err)
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
									e.client.System.ID,
									component,
									event,
								))
							} else {
								e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
									covAvg,
									prometheus.GaugeValue,
									value,
									e.client.System.ID,
									component,
									event,
									period,
								))
							}
						}
					}
				}
				log.Debugf("%s: GatherMetrics() completed GetAppCoverageMetrics(%s)", e.client.System.ID, component)
			}
			if cmds["cluster/status DB"] {
				log.Debugf("%s: GatherMetrics() calls GetAppClusteringInfo(%s)", e.client.System.ID, component)
				if cluster, err := e.client.GetAppClusteringInfo(component); err != nil {
					isClusterEnabled = false
					log.Errorf("%s: %v", component, err)
					//e.IncrementErrorCounter()
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterEnabled,
						prometheus.GaugeValue,
						0,
						e.client.System.ID,
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
						e.client.System.ID,
						component,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterRole,
						prometheus.GaugeValue,
						float64(cluster.Role),
						e.client.System.ID,
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
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterTerm,
						prometheus.CounterValue,
						float64(cluster.Term),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterNotCommittedEntryCount,
						prometheus.GaugeValue,
						float64(cluster.NotCommittedEntries),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterNotAppliedEntryCount,
						prometheus.GaugeValue,
						float64(cluster.NotAppliedEntries),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterNextIndex,
						prometheus.CounterValue,
						float64(cluster.NextIndex),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterMatchIndex,
						prometheus.CounterValue,
						float64(cluster.MatchIndex),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterLogLowIndex,
						prometheus.CounterValue,
						float64(cluster.Log.Low),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterLogHighIndex,
						prometheus.CounterValue,
						float64(cluster.Log.High),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterLeaderSelf,
						prometheus.GaugeValue,
						float64(cluster.IsLeaderSelf),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterVoteSelf,
						prometheus.GaugeValue,
						float64(cluster.IsVotedSelf),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterPeerCount,
						prometheus.GaugeValue,
						float64(len(cluster.Peers)),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterPeerInConnTotal,
						prometheus.GaugeValue,
						float64(cluster.Connections.Inbound),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
						clusterPeerOutConnTotal,
						prometheus.GaugeValue,
						float64(cluster.Connections.Outbound),
						e.client.System.ID,
						component,
						cluster.ID,
						cluster.ClusterID,
					))
					for peerID, peer := range cluster.Peers {
						e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
							clusterPeerNextIndex,
							prometheus.CounterValue,
							float64(peer.NextIndex),
							e.client.System.ID,
							component,
							cluster.ID,
							cluster.ClusterID,
							peerID,
						))
						e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
							clusterPeerMatchIndex,
							prometheus.CounterValue,
							float64(peer.MatchIndex),
							e.client.System.ID,
							component,
							cluster.ID,
							cluster.ClusterID,
							peerID,
						))
						e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
							clusterPeerInConnInfo,
							prometheus.GaugeValue,
							float64(peer.Connection.Inbound),
							e.client.System.ID,
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
							e.client.System.ID,
							component,
							cluster.ID,
							cluster.ClusterID,
							peerID,
							peer.Address,
						))
					}
					//log.Infof("%s: %v", component, cluster)
				}
				log.Debugf("%s: GatherMetrics() completed GetAppClusteringInfo(%s)", e.client.System.ID, component)
			}
		}
	}

	if northClusterID != "" && southClusterID != "" {
		e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
			clusterGroup,
			prometheus.GaugeValue,
			1,
			e.client.System.ID,
			northClusterID+southClusterID,
		))
	}

	components = []string{
		"ovsdb-server-southbound",
		"ovsdb-server-northbound",
	}

	for _, component := range components {
		log.Debugf("%s: GatherMetrics() calls IsDefaultPortUp(%s)", e.client.System.ID, component)
		defaultPortUp, err := e.client.IsDefaultPortUp(component)
		if err != nil {
			log.Errorf("%s: %v", component, err)
			e.IncrementErrorCounter()
		}
		e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
			networkPortUp,
			prometheus.GaugeValue,
			float64(defaultPortUp),
			e.client.System.ID,
			component,
			"default",
		))
		log.Debugf("%s: GatherMetrics() completed IsDefaultPortUp(%s)", e.client.System.ID, component)
		log.Debugf("%s: GatherMetrics() calls IsSslPortUp(%s)", e.client.System.ID, component)
		sslPortUp, err := e.client.IsSslPortUp(component)
		if err != nil {
			log.Errorf("%s: %v", component, err)
			e.IncrementErrorCounter()
		}
		e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
			networkPortUp,
			prometheus.GaugeValue,
			float64(sslPortUp),
			e.client.System.ID,
			component,
			"ssl",
		))
		log.Debugf("%s: GatherMetrics() completed IsSslPortUp(%s)", e.client.System.ID, component)

		if isClusterEnabled {
			log.Debugf("%s: GatherMetrics() calls IsRaftPortUp(%s)", e.client.System.ID, component)
			raftPortUp, err := e.client.IsRaftPortUp(component)
			if err != nil {
				log.Errorf("%s: %v", component, err)
				e.IncrementErrorCounter()
			}
			e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
				networkPortUp,
				prometheus.GaugeValue,
				float64(raftPortUp),
				e.client.System.ID,
				component,
				"raft",
			))
			log.Debugf("%s: GatherMetrics() completed IsRaftPortUp(%s)", e.client.System.ID, component)
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
		e.client.System.ID, e.client.System.RunDir, e.client.System.Hostname,
		e.client.System.Type, e.client.System.Version,
		e.client.Database.Vswitch.Version, e.client.Database.Vswitch.Schema.Version,
	))

	e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
		requestErrors,
		prometheus.CounterValue,
		float64(e.errors),
		e.client.System.ID,
	))

	e.metrics = append(e.metrics, prometheus.MustNewConstMetric(
		nextPoll,
		prometheus.CounterValue,
		float64(e.nextCollectionTicker),
		e.client.System.ID,
	))

	e.nextCollectionTicker = time.Now().Add(time.Duration(e.pollInterval) * time.Second).Unix()

	log.Debugf("%s: GatherMetrics() returns", e.client.System.ID)
	return
}

func init() {
	prometheus.MustRegister(version.NewCollector(namespace + "_exporter"))
}

func main() {
	var listenAddress string
	var metricsPath string
	var pollTimeout int
	var pollInterval int
	var isShowVersion bool
	var logLevel string
	var opts exporterOpts

	var systemRunDir string
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

	var usageHelp = func() {
		fmt.Fprintf(os.Stderr, "\n%s - Prometheus Exporter for Open Virtual Network (OVN)\n\n", appName)
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments]\n\n", appName)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nDocumentation: https://github.com/greenpau/ovn_exporter/\n\n")
	}
	flag.Usage = usageHelp
	flag.Parse()
	opts.timeout = pollTimeout
	version.Version = appVersion
	version.Revision = gitCommit
	version.Branch = gitBranch
	version.BuildUser = buildUser
	version.BuildDate = buildDate

	if err := log.Base().SetLevel(logLevel); err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}

	if isShowVersion {
		fmt.Fprintf(os.Stdout, "%s %s", appName, version.Version)
		if version.Revision != "" {
			fmt.Fprintf(os.Stdout, ", commit: %s\n", version.Revision)
		} else {
			fmt.Fprint(os.Stdout, "\n")
		}
		os.Exit(0)
	}

	log.Infof("Starting %s %s", appName, version.Info())
	log.Infof("Build context %s", version.BuildContext())

	exporter, err := NewExporter(opts)
	if err != nil {
		log.Errorf("%s failed to init properly: %s", appName, err)
	}

	exporter.client.System.RunDir = systemRunDir

	exporter.client.Database.Vswitch.Name = databaseVswitchName
	exporter.client.Database.Vswitch.Socket.Remote = databaseVswitchSocketRemote
	exporter.client.Database.Vswitch.File.Data.Path = databaseVswitchFileDataPath
	exporter.client.Database.Vswitch.File.Log.Path = databaseVswitchFileLogPath
	exporter.client.Database.Vswitch.File.Pid.Path = databaseVswitchFilePidPath
	exporter.client.Database.Vswitch.File.SystemID.Path = databaseVswitchFileSystemIDPath

	exporter.client.Database.Northbound.Name = databaseNorthboundName
	exporter.client.Database.Northbound.Socket.Remote = databaseNorthboundSocketRemote
	exporter.client.Database.Northbound.Socket.Control = databaseNorthboundSocketControl
	exporter.client.Database.Northbound.File.Data.Path = databaseNorthboundFileDataPath
	exporter.client.Database.Northbound.File.Log.Path = databaseNorthboundFileLogPath
	exporter.client.Database.Northbound.File.Pid.Path = databaseNorthboundFilePidPath
	exporter.client.Database.Northbound.Port.Default = databaseNorthboundPortDefault
	exporter.client.Database.Northbound.Port.Ssl = databaseNorthboundPortSsl
	exporter.client.Database.Northbound.Port.Raft = databaseNorthboundPortRaft

	exporter.client.Database.Southbound.Name = databaseSouthboundName
	exporter.client.Database.Southbound.Socket.Remote = databaseSouthboundSocketRemote
	exporter.client.Database.Southbound.Socket.Control = databaseSouthboundSocketControl
	exporter.client.Database.Southbound.File.Data.Path = databaseSouthboundFileDataPath
	exporter.client.Database.Southbound.File.Log.Path = databaseSouthboundFileLogPath
	exporter.client.Database.Southbound.File.Pid.Path = databaseSouthboundFilePidPath
	exporter.client.Database.Southbound.Port.Default = databaseSouthboundPortDefault
	exporter.client.Database.Southbound.Port.Ssl = databaseSouthboundPortSsl
	exporter.client.Database.Southbound.Port.Raft = databaseSouthboundPortRaft

	exporter.client.Service.Vswitchd.File.Log.Path = serviceVswitchdFileLogPath
	exporter.client.Service.Vswitchd.File.Pid.Path = serviceVswitchdFilePidPath

	exporter.client.Service.Northd.File.Log.Path = serviceNorthdFileLogPath
	exporter.client.Service.Northd.File.Pid.Path = serviceNorthdFilePidPath

	log.Infof("OVS system-id: %s", exporter.client.System.ID)
	exporter.pollInterval = int64(pollInterval)
	prometheus.MustRegister(exporter)

	http.Handle(metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>OVN Exporter</title></head>
             <body>
             <h1>OVN Exporter</h1>
             <p><a href='` + metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	log.Infoln("Listening on", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
