// Package status collects and aggregates server health metrics for the simplex-node
// status API. It provides Collect (node info, Docker SMP/XFTP status, vault usage,
// disk metrics), CheckDiskAndAlert (with USB detection and critical/warning thresholds),
// and reputation scoring (royal node, banknote holders, auditors). Disk usage is
// checked for root, data, smp_state, and xftp_state partitions.
package status
