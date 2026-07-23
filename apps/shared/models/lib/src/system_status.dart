/// SystemStatus manages systemstatus functionality.
class SystemStatus {
  final String status;
  final int uptimeSeconds;
  final int totalBanknotes;
  final int totalHolders;
  final int reserveNg;
  final int supplyNg;
  final int activeUsers;

  const SystemStatus({
    required this.status,
    this.uptimeSeconds = 0,
    this.totalBanknotes = 0,
    this.totalHolders = 0,
    this.reserveNg = 0,
    this.supplyNg = 0,
    this.activeUsers = 0,
  });

  factory SystemStatus.fromJson(Map<String, dynamic> json) {
    return SystemStatus(
      status: json['status'] ?? 'unknown',
      uptimeSeconds: (json['uptime'] ?? 0).toInt(),
      totalBanknotes: (json['total_banknotes'] ?? 0).toInt(),
      totalHolders: (json['total_holders'] ?? 0).toInt(),
      reserveNg: (json['reserve_ng'] ?? 0).toInt(),
      supplyNg: (json['supply_ng'] ?? 0).toInt(),
      activeUsers: (json['active_users'] ?? 0).toInt(),
    );
  }

  String get uptimeFormatted {
    final h = uptimeSeconds ~/ 3600;
    final m = (uptimeSeconds % 3600) ~/ 60;
    final s = uptimeSeconds % 60;
    return '${h}h ${m}m ${s}s';
  }
}
