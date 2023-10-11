# Open Virtual Network (OVN) Exporter

<a href="https://github.com/greenpau/ovn_exporter/actions/" target="_blank"><img src="https://github.com/greenpau/ovn_exporter/workflows/build/badge.svg?branch=main"></a>

Export Open Virtual Network (OVN) data to Prometheus.

## Introduction

This exporter exports metrics from the following OVN components:
* OVN `northd` service
* OVS `vswitchd` service
* `OVN Northbound` database
* `OVN Southbound` database
* `Open_vSwitch` database

## Getting Started

Run the following commands to install it:

```bash
wget https://github.com/greenpau/ovn_exporter/releases/download/v1.0.0/ovn-exporter-1.0.0.linux-amd64.tar.gz
tar xvzf ovn-exporter-1.0.0.linux-amd64.tar.gz
cd ovn-exporter*
./install.sh
cd ..
rm -rf ovn-exporter-1.0.0.linux-amd64*
systemctl status ovn-exporter -l
curl -s localhost:9476/metrics | grep server_id
```

Run the following commands to build and test it:

```bash
cd $GOPATH/src
mkdir -p github.com/greenpau
cd github.com/greenpau
git clone https://github.com/greenpau/ovn_exporter.git
cd ovn_exporter
make
make qtest
```

## Exported Metrics

| Metric | Meaning | Labels |
| ------ | ------- | ------ |
| `ovn_chassis_info` | Whether the OVN chassis is up (1) or down (0), together with additional information about the chassis. | `system_id` |
| `ovn_cluster_enabled` |  Is OVN clustering enabled (1) or not (0). | `system_id` |
| `ovn_cluster_inbound_peer_conn_total` |  The total number of outbound connections to cluster peers. | `system_id` |
| `ovn_cluster_leader_self` |  Is this server consider itself a leader (1) or not (0). | `system_id` |
| `ovn_cluster_log_high_index` |  The raft's high number associated with this server. | `system_id` |
| `ovn_cluster_log_low_index` |  The raft's low number associated with this server. | `system_id` |
| `ovn_cluster_match_index` |  The raft's match index associated with this server. | `system_id` |
| `ovn_cluster_next_index` |  The raft's next index associated with this server. | `system_id` |
| `ovn_cluster_outbound_peer_conn_total` |  The total number of inbound connections from cluster peers. | `system_id` |
| `ovn_cluster_peer_count` |  The total number of peers in this server's cluster. | `system_id` |
| `ovn_cluster_pending_entry_count` |  The number of raft entries not yet applied by this server. | `system_id` |
| `ovn_cluster_role` |  The role of this server in the cluster. The values are: 3 - leader, 2 - candidate, 1 - follower, 0 - other. | `system_id` |
| `ovn_cluster_status` |  The status of this server in the cluster. The values are: 1 - cluster member, 0 - other. | `system_id` |
| `ovn_cluster_term` |  The current raft term known by this server. | `system_id` |
| `ovn_cluster_uncommitted_entry_count` |  The number of raft entries not yet committed by this server. | `system_id` |
| `ovn_cluster_vote_self` |  Is this server voted itself as a leader (1) or not (0). | `system_id` |
| `ovn_coverage_avg` |  The average rate of the number of times particular events occur during a OVSDB daemon's runtime. | `system_id` |
| `ovn_coverage_total` |  The total number of times particular events occur during a OVSDB daemon's runtime. | `system_id` |
| `ovn_exporter_build_info` |  A metric with a constant '1' value labeled by version, revision, branch, and goversion from which ovn_exporter was built. | `system_id` |
| `ovn_failed_req_count` |  The number of failed requests to OVN stack. | `system_id` |
| `ovn_info` |  This metric provides basic information about OVN stack. It is always set to 1. | `system_id` |
| `ovn_log_file_size` |  The size of a log file associated with an OVN component. | `system_id` |
| `ovn_logical_switch_external_id` |  Provides the external IDs and values associated with OVN logical switches. This metric is always up (1). | `system_id` |
| `ovn_logical_switch_info` |  The information about OVN logical switch. This metric is always up (1). | `system_id` |
| `ovn_logical_switch_port_binding` |  Provides the association between a logical switch and a logical switch port. This metric is always up (1). | `system_id` |
| `ovn_logical_switch_port_info` |  The information about OVN logical switch port. This metric is always up (1). | `system_id` |
| `ovn_logical_switch_port_tunnel_key` |  The value of the tunnel key associated with the logical switch port. | `system_id` |
| `ovn_logical_switch_ports` |  The number of logical switch ports connected to the OVN logical switch. | `system_id` |
| `ovn_logical_switch_tunnel_key` |  The value of the tunnel key associated with the logical switch. | `system_id` |
| `ovn_network_port` |  The TCP port used for database connection. If the value is 0, then the port is not in use. | `system_id` |
| `ovn_next_poll` |  The timestamp of the next potential poll of OVN stack. | `system_id` |
| `ovn_pid` |  The process ID of a running OVN component. If the component is not running, then the ID is 0. | `system_id` |
| `ovn_cluster_group` | The cluster group in which this server participates. It is a combination of SB and NB cluster IDs. This metric is always up (1). | `system_id`, `cluster_group` |
| `ovn_up` |  Is OVN stack up (1) or is it down (0). | `system_id` |

For example:

```bash
$ curl localhost:9476/metrics | grep ovn
# HELP ovn_chassis_info Whether the OVN chassis is up (1) or down (0), together with additional information about the chassis.
# TYPE ovn_chassis_info gauge
ovn_chassis_info{ip="172.16.10.1",name="nyrtr1-6500120-vlan-20",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="c9b7412f-2c27-4191-b2ab-729b93ffa3cd"} 0
ovn_chassis_info{ip="172.16.10.10",name="7592b50a-c201-48ea-8737-4748c185237f",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="935fe428-4adb-47ed-b3e4-1497655ffa79"} 0
ovn_chassis_info{ip="172.16.10.2",name="nyrtr2-6500120-vlan-20",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="bb41cb2c-ea2a-4743-bb2c-6fb0ebf4900d"} 0
ovn_chassis_info{ip="172.16.10.20",name="fa2b92b1-83ff-47a4-ad4a-da219df28a91",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="bc3d1542-a85e-47e6-8d33-412735eaa664"} 0
# HELP ovn_cluster_enabled Is OVN clustering enabled (1) or not (0).
# TYPE ovn_cluster_enabled gauge
ovn_cluster_enabled{component="ovsdb-server-northbound",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1
ovn_cluster_enabled{component="ovsdb-server-southbound",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1
# HELP ovn_cluster_inbound_peer_conn_total The total number of outbound connections to cluster peers.
# TYPE ovn_cluster_inbound_peer_conn_total gauge
ovn_cluster_inbound_peer_conn_total{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_cluster_inbound_peer_conn_total{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
# HELP ovn_cluster_leader_self Is this server consider itself a leader (1) or not (0).
# TYPE ovn_cluster_leader_self gauge
ovn_cluster_leader_self{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1
ovn_cluster_leader_self{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1
# HELP ovn_cluster_log_high_index The raft's high number associated with this server.
# TYPE ovn_cluster_log_high_index counter
ovn_cluster_log_high_index{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 95
ovn_cluster_log_high_index{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 121
# HELP ovn_cluster_log_low_index The raft's low number associated with this server.
# TYPE ovn_cluster_log_low_index counter
ovn_cluster_log_low_index{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 67
ovn_cluster_log_low_index{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 93
# HELP ovn_cluster_match_index The raft's match index associated with this server.
# TYPE ovn_cluster_match_index counter
ovn_cluster_match_index{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 94
ovn_cluster_match_index{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 120
# HELP ovn_cluster_next_index The raft's next index associated with this server.
# TYPE ovn_cluster_next_index counter
ovn_cluster_next_index{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 94
ovn_cluster_next_index{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 120
# HELP ovn_cluster_outbound_peer_conn_total The total number of inbound connections from cluster peers.
# TYPE ovn_cluster_outbound_peer_conn_total gauge
ovn_cluster_outbound_peer_conn_total{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_cluster_outbound_peer_conn_total{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
# HELP ovn_cluster_peer_count The total number of peers in this server's cluster.
# TYPE ovn_cluster_peer_count gauge
ovn_cluster_peer_count{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_cluster_peer_count{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
# HELP ovn_cluster_pending_entry_count The number of raft entries not yet applied by this server.
# TYPE ovn_cluster_pending_entry_count gauge
ovn_cluster_pending_entry_count{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_cluster_pending_entry_count{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
# HELP ovn_cluster_role The role of this server in the cluster. The values are: 3 - leader, 2 - candidate, 1 - follower, 0 - other.
# TYPE ovn_cluster_role gauge
ovn_cluster_role{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 3
ovn_cluster_role{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 3
# HELP ovn_cluster_status The status of this server in the cluster. The values are: 1 - cluster member, 0 - other.
# TYPE ovn_cluster_status gauge
ovn_cluster_status{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1
ovn_cluster_status{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1
# HELP ovn_cluster_term The current raft term known by this server.
# TYPE ovn_cluster_term counter
ovn_cluster_term{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 29
ovn_cluster_term{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 29
# HELP ovn_cluster_uncommitted_entry_count The number of raft entries not yet committed by this server.
# TYPE ovn_cluster_uncommitted_entry_count gauge
ovn_cluster_uncommitted_entry_count{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_cluster_uncommitted_entry_count{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
# HELP ovn_cluster_vote_self Is this server voted itself as a leader (1) or not (0).
# TYPE ovn_cluster_vote_self gauge
ovn_cluster_vote_self{cluster_id="5c1c",component="ovsdb-server-northbound",server_id="983c",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1
ovn_cluster_vote_self{cluster_id="b3be",component="ovsdb-server-southbound",server_id="6ef",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1
# HELP ovn_coverage_avg The average rate of the number of times particular events occur during a OVSDB daemon's runtime.
# TYPE ovn_coverage_avg gauge
ovn_coverage_avg{component="ovsdb-server-northbound",event="hmap_expand",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 8.4267
ovn_coverage_avg{component="ovsdb-server-northbound",event="hmap_expand",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 12.117
ovn_coverage_avg{component="ovsdb-server-northbound",event="hmap_expand",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 8.4
ovn_coverage_avg{component="ovsdb-server-northbound",event="hmap_pathological",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="hmap_pathological",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="hmap_pathological",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="lockfile_lock",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="lockfile_lock",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="lockfile_lock",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="poll_create_node",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 22.2428
ovn_coverage_avg{component="ovsdb-server-northbound",event="poll_create_node",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 25.85
ovn_coverage_avg{component="ovsdb-server-northbound",event="poll_create_node",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 26.6
ovn_coverage_avg{component="ovsdb-server-northbound",event="poll_zero_timeout",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.1325
ovn_coverage_avg{component="ovsdb-server-northbound",event="poll_zero_timeout",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.167
ovn_coverage_avg{component="ovsdb-server-northbound",event="poll_zero_timeout",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="pstream_open",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="pstream_open",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="pstream_open",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="seq_change",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 3.6892
ovn_coverage_avg{component="ovsdb-server-northbound",event="seq_change",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 4.017
ovn_coverage_avg{component="ovsdb-server-northbound",event="seq_change",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 3.8
ovn_coverage_avg{component="ovsdb-server-northbound",event="unixctl_received",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.0058
ovn_coverage_avg{component="ovsdb-server-northbound",event="unixctl_received",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.1
ovn_coverage_avg{component="ovsdb-server-northbound",event="unixctl_received",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="unixctl_replied",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.0058
ovn_coverage_avg{component="ovsdb-server-northbound",event="unixctl_replied",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.1
ovn_coverage_avg{component="ovsdb-server-northbound",event="unixctl_replied",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-northbound",event="util_xalloc",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 216.5706
ovn_coverage_avg{component="ovsdb-server-northbound",event="util_xalloc",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 297.617
ovn_coverage_avg{component="ovsdb-server-northbound",event="util_xalloc",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 221.4
ovn_coverage_avg{component="ovsdb-server-southbound",event="hmap_expand",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 8.4058
ovn_coverage_avg{component="ovsdb-server-southbound",event="hmap_expand",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 11.967
ovn_coverage_avg{component="ovsdb-server-southbound",event="hmap_expand",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 8.4
ovn_coverage_avg{component="ovsdb-server-southbound",event="hmap_pathological",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="hmap_pathological",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="hmap_pathological",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="lockfile_lock",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="lockfile_lock",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="lockfile_lock",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="poll_create_node",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 22.2553
ovn_coverage_avg{component="ovsdb-server-southbound",event="poll_create_node",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 25.8
ovn_coverage_avg{component="ovsdb-server-southbound",event="poll_create_node",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 26.6
ovn_coverage_avg{component="ovsdb-server-southbound",event="poll_zero_timeout",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.1411
ovn_coverage_avg{component="ovsdb-server-southbound",event="poll_zero_timeout",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.15
ovn_coverage_avg{component="ovsdb-server-southbound",event="poll_zero_timeout",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="pstream_open",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="pstream_open",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="pstream_open",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="seq_change",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 3.6922
ovn_coverage_avg{component="ovsdb-server-southbound",event="seq_change",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 4.017
ovn_coverage_avg{component="ovsdb-server-southbound",event="seq_change",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 3.8
ovn_coverage_avg{component="ovsdb-server-southbound",event="unixctl_received",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.0058
ovn_coverage_avg{component="ovsdb-server-southbound",event="unixctl_received",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.1
ovn_coverage_avg{component="ovsdb-server-southbound",event="unixctl_received",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="unixctl_replied",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.0058
ovn_coverage_avg{component="ovsdb-server-southbound",event="unixctl_replied",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0.1
ovn_coverage_avg{component="ovsdb-server-southbound",event="unixctl_replied",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
ovn_coverage_avg{component="ovsdb-server-southbound",event="util_xalloc",interval="1h",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 216.5233
ovn_coverage_avg{component="ovsdb-server-southbound",event="util_xalloc",interval="5m",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 299.017
ovn_coverage_avg{component="ovsdb-server-southbound",event="util_xalloc",interval="5s",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 221.4
# HELP ovn_coverage_total The total number of times particular events occur during a OVSDB daemon's runtime.
# TYPE ovn_coverage_total counter
ovn_coverage_total{component="ovsdb-server-northbound",event="hmap_expand",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 679295
ovn_coverage_total{component="ovsdb-server-northbound",event="hmap_pathological",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 3
ovn_coverage_total{component="ovsdb-server-northbound",event="lockfile_lock",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1
ovn_coverage_total{component="ovsdb-server-northbound",event="poll_create_node",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1.810543e+06
ovn_coverage_total{component="ovsdb-server-northbound",event="poll_zero_timeout",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 11286
ovn_coverage_total{component="ovsdb-server-northbound",event="pstream_open",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 4
ovn_coverage_total{component="ovsdb-server-northbound",event="seq_change",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 301169
ovn_coverage_total{component="ovsdb-server-northbound",event="unixctl_received",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 190
ovn_coverage_total{component="ovsdb-server-northbound",event="unixctl_replied",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 190
ovn_coverage_total{component="ovsdb-server-northbound",event="util_xalloc",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1.7508459e+07
ovn_coverage_total{component="ovsdb-server-southbound",event="hmap_expand",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 679519
ovn_coverage_total{component="ovsdb-server-southbound",event="hmap_pathological",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 5
ovn_coverage_total{component="ovsdb-server-southbound",event="lockfile_lock",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1
ovn_coverage_total{component="ovsdb-server-southbound",event="poll_create_node",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1.814127e+06
ovn_coverage_total{component="ovsdb-server-southbound",event="poll_zero_timeout",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 11379
ovn_coverage_total{component="ovsdb-server-southbound",event="pstream_open",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 4
ovn_coverage_total{component="ovsdb-server-southbound",event="seq_change",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 301582
ovn_coverage_total{component="ovsdb-server-southbound",event="unixctl_received",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 190
ovn_coverage_total{component="ovsdb-server-southbound",event="unixctl_replied",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 190
ovn_coverage_total{component="ovsdb-server-southbound",event="util_xalloc",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1.75451e+07
# HELP ovn_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which ovn_exporter was built.
# TYPE ovn_exporter_build_info gauge
ovn_exporter_build_info{branch="master",goversion="go1.10.2",revision="2f70c95",version="1.0.0"} 1
# HELP ovn_failed_req_count The number of failed requests to OVN stack.
# TYPE ovn_failed_req_count counter
ovn_failed_req_count{system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 0
# HELP ovn_info This metric provides basic information about OVN stack. It is always set to 1.
# TYPE ovn_info gauge
ovn_info{db_version="7.16.1",hostname="godev.local",ovs_version="2.10.90",rundir="/var/run/openvswitch",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",system_type="CentOS",system_version="7.5.1804-Core"} 1
# HELP ovn_log_file_size The size of a log file associated with an OVN component.
# TYPE ovn_log_file_size gauge
ovn_log_file_size{component="ovn-northd",filename="/var/log/openvswitch/ovn-northd.log",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 93
ovn_log_file_size{component="ovs-vswitchd",filename="/var/log/openvswitch/ovs-vswitchd.log",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 95
ovn_log_file_size{component="ovsdb-server",filename="/var/log/openvswitch/ovsdb-server.log",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 95
ovn_log_file_size{component="ovsdb-server-northbound",filename="/var/log/openvswitch/ovsdb-server-nb.log",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 98
ovn_log_file_size{component="ovsdb-server-southbound",filename="/var/log/openvswitch/ovsdb-server-sb.log",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 98
# HELP ovn_logical_switch_external_id Provides the external IDs and values associated with OVN logical switches. This metric is always up (1).
# TYPE ovn_logical_switch_external_id gauge
ovn_logical_switch_external_id{key="gateway_ip",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40",value="10.10.10.1"} 1
ovn_logical_switch_external_id{key="subnet",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40",value="10.10.10.0/24"} 1
ovn_logical_switch_external_id{key="subnet_context",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40",value="default"} 1
# HELP ovn_logical_switch_info The information about OVN logical switch. This metric is always up (1).
# TYPE ovn_logical_switch_info gauge
ovn_logical_switch_info{name="19a05268b5eb3df10e2d50b8220505ea0026679bb62eb39d71c8707dd5165248",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40"} 1
# HELP ovn_logical_switch_port_binding Provides the association between a logical switch and a logical switch port. This metric is always up (1).
# TYPE ovn_logical_switch_port_binding gauge
ovn_logical_switch_port_binding{port="1080f06b-3252-4b7d-8e49-d146c43010a3",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40"} 1
ovn_logical_switch_port_binding{port="1660dfb8-3bfb-4ffd-ba3a-203df17e2235",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40"} 1
ovn_logical_switch_port_binding{port="5de3762a-cf08-458b-87bb-89a1fd226bcb",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40"} 1
ovn_logical_switch_port_binding{port="73dbf01c-acb3-4b0f-90bc-ce7b11aea128",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40"} 1
ovn_logical_switch_port_binding{port="a510e307-faa6-4707-bf22-c9b28c1f0d00",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40"} 1
ovn_logical_switch_port_binding{port="d2676f53-5daf-4514-8a5b-abbc92504894",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40"} 1
# HELP ovn_logical_switch_port_info The information about OVN logical switch port. This metric is always up (1).
# TYPE ovn_logical_switch_port_info gauge
ovn_logical_switch_port_info{chassis="935fe428-4adb-47ed-b3e4-1497655ffa79",datapath="a7b76868-1725-418d-8289-175798c19db7",ip_address="10.10.10.111",logical_switch="",mac_address="02:54:b4:11:3b:e6",name="9da77936277dcf536dd03fa0351578948aaf9b9e599063fb9b305e4b2ef977a8",port_binding="2ea56412-37c7-447b-ab1c-b0a86ed42573",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="d2676f53-5daf-4514-8a5b-abbc92504894"} 1
ovn_logical_switch_port_info{chassis="935fe428-4adb-47ed-b3e4-1497655ffa79",datapath="a7b76868-1725-418d-8289-175798c19db7",ip_address="10.10.10.112",logical_switch="",mac_address="02:27:5b:bd:a9:70",name="0375c97d5224fbe7cd2d10bfe4c14340b89b288ad20bd759b4a1e385fbb81395",port_binding="20fe176d-22ed-4d0c-9209-d2a2d68f060d",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="5de3762a-cf08-458b-87bb-89a1fd226bcb"} 1
ovn_logical_switch_port_info{chassis="bb41cb2c-ea2a-4743-bb2c-6fb0ebf4900d",datapath="a7b76868-1725-418d-8289-175798c19db7",ip_address="10.10.10.2",logical_switch="",mac_address="b7:65:9d:71:22:07",name="nyrtr2-6500120-vlan-20-p1",port_binding="59188438-cb9a-4ae8-bc3d-d62c2ca012a3",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="1660dfb8-3bfb-4ffd-ba3a-203df17e2235"} 1
ovn_logical_switch_port_info{chassis="bc3d1542-a85e-47e6-8d33-412735eaa664",datapath="a7b76868-1725-418d-8289-175798c19db7",ip_address="10.10.10.121",logical_switch="",mac_address="02:e9:3d:b0:f9:f5",name="024126f4fe4cc21a95e8aa686d9e30767f3222a2f2adac5d9d1c85f63c27bfd7",port_binding="76ef669e-518b-40d9-9ed8-15de5f246749",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="a510e307-faa6-4707-bf22-c9b28c1f0d00"} 1
ovn_logical_switch_port_info{chassis="bc3d1542-a85e-47e6-8d33-412735eaa664",datapath="a7b76868-1725-418d-8289-175798c19db7",ip_address="10.10.10.122",logical_switch="",mac_address="02:32:6b:ca:25:9e",name="2b2bf7e475a74b6f48b8c92c750a13188b1c09f4c7e6342ed8aaa62a00628969",port_binding="23146d8f-d30e-4e61-964f-17c4695eae38",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="73dbf01c-acb3-4b0f-90bc-ce7b11aea128"} 1
ovn_logical_switch_port_info{chassis="c9b7412f-2c27-4191-b2ab-729b93ffa3cd",datapath="a7b76868-1725-418d-8289-175798c19db7",ip_address="10.10.10.1",logical_switch="",mac_address="b7:65:9d:71:22:07",name="nyrtr1-6500120-vlan-20-p1",port_binding="d73c06e9-0bee-480a-9467-45965845f243",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="1080f06b-3252-4b7d-8e49-d146c43010a3"} 1
# HELP ovn_logical_switch_port_tunnel_key The value of the tunnel key associated with the logical switch port.
# TYPE ovn_logical_switch_port_tunnel_key gauge
ovn_logical_switch_port_tunnel_key{system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="1080f06b-3252-4b7d-8e49-d146c43010a3"} 1
ovn_logical_switch_port_tunnel_key{system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="1660dfb8-3bfb-4ffd-ba3a-203df17e2235"} 2
ovn_logical_switch_port_tunnel_key{system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="5de3762a-cf08-458b-87bb-89a1fd226bcb"} 4
ovn_logical_switch_port_tunnel_key{system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="73dbf01c-acb3-4b0f-90bc-ce7b11aea128"} 6
ovn_logical_switch_port_tunnel_key{system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="a510e307-faa6-4707-bf22-c9b28c1f0d00"} 5
ovn_logical_switch_port_tunnel_key{system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="d2676f53-5daf-4514-8a5b-abbc92504894"} 3
# HELP ovn_logical_switch_ports The number of logical switch ports connected to the OVN logical switch.
# TYPE ovn_logical_switch_ports gauge
ovn_logical_switch_ports{system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40"} 6
# HELP ovn_logical_switch_tunnel_key The value of the tunnel key associated with the logical switch.
# TYPE ovn_logical_switch_tunnel_key gauge
ovn_logical_switch_tunnel_key{system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",uuid="61f9dca6-2339-4b07-a1eb-7cb2e2fa0a40"} 6.50012e+06
# HELP ovn_network_port The TCP port used for database connection. If the value is 0, then the port is not in use.
# TYPE ovn_network_port gauge
ovn_network_port{component="ovsdb-server-northbound",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",usage="default"} 6641
ovn_network_port{component="ovsdb-server-northbound",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",usage="raft"} 6643
ovn_network_port{component="ovsdb-server-northbound",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",usage="ssl"} 0
ovn_network_port{component="ovsdb-server-southbound",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",usage="default"} 6642
ovn_network_port{component="ovsdb-server-southbound",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",usage="raft"} 6644
ovn_network_port{component="ovsdb-server-southbound",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",usage="ssl"} 0
# HELP ovn_next_poll The timestamp of the next potential poll of OVN stack.
# TYPE ovn_next_poll counter
ovn_next_poll{system_id="bea816d9-f201-4d69-a609-5b03f278f5b9"} 1.537492716e+09
# HELP ovn_pid The process ID of a running OVN component. If the component is not running, then the ID is 0.
# TYPE ovn_pid gauge
ovn_pid{component="ovn-northd",group="root",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",user="root"} 30210
ovn_pid{component="ovn-northd-monitoring",group="root",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",user="root"} 30209
ovn_pid{component="ovs-vswitchd",group="openvswitch",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",user="openvswitch"} 2326
ovn_pid{component="ovsdb-server",group="openvswitch",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",user="openvswitch"} 427
ovn_pid{component="ovsdb-server-northbound",group="root",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",user="root"} 30185
ovn_pid{component="ovsdb-server-northbound-monitoring",group="root",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",user="root"} 30184
ovn_pid{component="ovsdb-server-southbound",group="root",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",user="root"} 30197
ovn_pid{component="ovsdb-server-southbound-monitoring",group="root",system_id="bea816d9-f201-4d69-a609-5b03f278f5b9",user="root"} 30196
# HELP ovn_up Is OVN stack up (1) or is it down (0).
# TYPE ovn_up gauge
ovn_up 1
```

## Flags

```bash
./bin/ovn-exporter --help

ovn-exporter - Prometheus Exporter for Open Virtual Network (OVN)

Usage: ovn-exporter [arguments]

  -database.northbound.file.data.path string
        OVN NB db file. (default "/var/lib/openvswitch/ovnnb_db.db")
  -database.northbound.file.log.path string
        OVN NB db log file. (default "/var/log/openvswitch/ovsdb-server-nb.log")
  -database.northbound.file.pid.path string
        OVN NB db process id file. (default "/run/openvswitch/ovnnb_db.pid")
  -database.northbound.name string
        The name of OVN NB (northbound) db. (default "OVN_Northbound")
  -database.northbound.port.default int
        OVN NB db network socket port. (default 6641)
  -database.northbound.port.raft int
        OVN NB db network port for clustering (raft) (default 6643)
  -database.northbound.port.ssl int
        OVN NB db network socket secure port. (default 6631)
  -database.northbound.socket.control string
        JSON-RPC unix socket to OVN NB app. (default "unix:/run/openvswitch/ovnnb_db.ctl")
  -database.northbound.socket.remote string
        JSON-RPC unix socket to OVN NB db. (default "unix:/run/openvswitch/ovnnb_db.sock")
  -database.southbound.file.data.path string
        OVN SB db file. (default "/var/lib/openvswitch/ovnsb_db.db")
  -database.southbound.file.log.path string
        OVN SB db log file. (default "/var/log/openvswitch/ovsdb-server-sb.log")
  -database.southbound.file.pid.path string
        OVN SB db process id file. (default "/run/openvswitch/ovnsb_db.pid")
  -database.southbound.name string
        The name of OVN SB (southbound) db. (default "OVN_Southbound")
  -database.southbound.port.default int
        OVN SB db network socket port. (default 6642)
  -database.southbound.port.raft int
        OVN SB db network port for clustering (raft) (default 6644)
  -database.southbound.port.ssl int
        OVN SB db network socket secure port. (default 6632)
  -database.southbound.socket.control string
        JSON-RPC unix socket to OVN SB app. (default "unix:/run/openvswitch/ovnsb_db.ctl")
  -database.southbound.socket.remote string
        JSON-RPC unix socket to OVN SB db. (default "unix:/run/openvswitch/ovnsb_db.sock")
  -database.vswitch.file.data.path string
        OVS db file. (default "/etc/openvswitch/conf.db")
  -database.vswitch.file.log.path string
        OVS db log file. (default "/var/log/openvswitch/ovsdb-server.log")
  -database.vswitch.file.pid.path string
        OVS db process id file. (default "/var/run/openvswitch/ovsdb-server.pid")
  -database.vswitch.file.system.id.path string
        OVS system id file. (default "/etc/openvswitch/system-id.conf")
  -database.vswitch.name string
        The name of OVS db. (default "Open_vSwitch")
  -database.vswitch.socket.remote string
        JSON-RPC unix socket to OVS db. (default "unix:/var/run/openvswitch/db.sock")
  -log.level string
        logging severity level (default "info")
  -ovn.poll-interval int
        The minimum interval (in seconds) between collections from OVN server. (default 15)
  -ovn.timeout int
        Timeout on gRPC requests to OVN. (default 2)
  -service.ovn.northd.file.log.path string
        OVN northd daemon log file. (default "/var/log/openvswitch/ovn-northd.log")
  -service.ovn.northd.file.pid.path string
        OVN northd daemon process id file. (default "/run/openvswitch/ovn-northd.pid")
  -service.vswitchd.file.log.path string
        OVS vswitchd daemon log file. (default "/var/log/openvswitch/ovs-vswitchd.log")
  -service.vswitchd.file.pid.path string
        OVS vswitchd daemon process id file. (default "/var/run/openvswitch/ovs-vswitchd.pid")
  -system.run.dir string
        OVS default run directory. (default "/var/run/openvswitch")
  -version
        version information
  -web.listen-address string
        Address to listen on for web interface and telemetry. (default ":9476")
  -web.telemetry-path string
        Path under which to expose metrics. (default "/metrics")
```
