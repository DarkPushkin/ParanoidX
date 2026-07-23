import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../theme.dart';
import '../main.dart';

/// SystemScreen manages systemscreen functionality.
class SystemScreen extends StatefulWidget {
  const SystemScreen({super.key});
  @override
  State<SystemScreen> createState() => _SystemScreenState();
}

class _SystemScreenState extends State<SystemScreen> {
  Map<String, dynamic>? _health, _checks, _metrics, _docker, _paranoidx;
  Timer? _timer;

  @override
  void initState() { super.initState(); _refresh(); _timer = Timer.periodic(const Duration(seconds: 15), (_) => _refresh()); }
  @override
  void dispose() { _timer?.cancel(); super.dispose(); }

  Future<void> _refresh() async {
    final api = context.read<AppState>().api;
    try {
      final r = await Future.wait([
        api.getHealth(), api.getHealthChecks(), api.getSystemMetrics(),
        api.getDockerStatus(), api.getParanoidXStatus(),
      ], eagerError: false);
      if (mounted) setState(() { _health = r[0]; _checks = r[1]; _metrics = r[2]; _docker = r[3]; _paranoidx = r[4]; });
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Row(children: [
          Container(padding: const EdgeInsets.all(6), decoration: BoxDecoration(color: RoyalTheme.accent.withAlpha(30), borderRadius: BorderRadius.circular(8)),
            child: const Icon(Icons.monitor_heart, color: RoyalTheme.accent, size: 20)),
          const SizedBox(width: 12),
          const Text('System'),
        ]),
        actions: [IconButton(icon: const Icon(Icons.refresh), onPressed: _refresh)],
      ),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        _buildHealthOverview(),
        const SizedBox(height: 16),
        _buildMetricsGrid(),
        const SizedBox(height: 16),
        _buildDockerCard(),
        const SizedBox(height: 16),
        _buildParanoidXCard(),
        const SizedBox(height: 16),
        _buildCheckList(),
      ]),
    );
  }

  Widget _buildHealthOverview() {
    final h = _health ?? {};
    final healthy = h['healthy'] == true;
    final uptime = h['uptime_hours'] ?? 0;
    final server = h['server'] ?? h['status'] ?? 'ok';
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: RoyalTheme.glowBorder(color: healthy ? RoyalTheme.green : RoyalTheme.orange),
      child: Row(children: [
        Container(
          width: 48, height: 48,
          decoration: BoxDecoration(shape: BoxShape.circle, color: (healthy ? RoyalTheme.green : RoyalTheme.orange).withAlpha(30)),
          child: Icon(healthy ? Icons.check_circle : Icons.warning, color: healthy ? RoyalTheme.green : RoyalTheme.orange, size: 28),
        ),
        const SizedBox(width: 16),
        Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(healthy ? 'All Systems Operational' : 'System Degraded', style: Theme.of(context).textTheme.titleMedium),
          Text('Uptime: ${uptime}h | Server: $server', style: Theme.of(context).textTheme.bodyMedium),
        ])),
      ]),
    );
  }

  Widget _buildMetricsGrid() {
    final m = _metrics ?? {};
    final cpu = m['cpu'] ?? m['cpu_percent'] ?? '—';
    final ram = m['ram'] ?? m['memory'] ?? m['memory_percent'] ?? '—';
    final disk = m['disk'] ?? m['disk_percent'] ?? '—';
    return Row(children: [
      Expanded(child: _metricCard('CPU', '$cpu%', RoyalTheme.accent, Icons.memory)),
      const SizedBox(width: 12),
      Expanded(child: _metricCard('RAM', '$ram%', RoyalTheme.green, Icons.storage)),
      const SizedBox(width: 12),
      Expanded(child: _metricCard('Disk', '$disk%', RoyalTheme.orange, Icons.disc_full)),
    ]);
  }

  Widget _metricCard(String label, String value, Color color, IconData icon) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: RoyalTheme.gradientCard(),
      child: Column(children: [
        Icon(icon, color: color, size: 20),
        const SizedBox(height: 6),
        Text(value, style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: color)),
        Text(label, style: Theme.of(context).textTheme.bodyMedium),
      ]),
    );
  }

  Widget _buildDockerCard() {
    final d = _docker ?? {};
    final containers = d['containers'] as List? ?? [];
    final running = containers.where((c) => (c is Map && c['state'] == 'running')).length;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: RoyalTheme.gradientCard(),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Icon(Icons.view_in_ar, color: RoyalTheme.accent, size: 18),
          const SizedBox(width: 8),
          Text('Docker ($running/${containers.length} running)', style: Theme.of(context).textTheme.titleMedium),
        ]),
        const SizedBox(height: 12),
        if (containers.isEmpty)
          Text('No containers', style: Theme.of(context).textTheme.bodyMedium)
        else
          ...containers.take(4).map((c) {
            final m = c as Map<String, dynamic>;
            final ok = m['state'] == 'running' || m['healthy'] == true;
            return Padding(
              padding: const EdgeInsets.symmetric(vertical: 3),
              child: Row(children: [
                Container(width: 8, height: 8, decoration: BoxDecoration(
                  shape: BoxShape.circle, color: ok ? RoyalTheme.green : RoyalTheme.red)),
                const SizedBox(width: 8),
                Text('${m['name'] ?? m['service'] ?? '—'}', style: const TextStyle(fontSize: 13)),
                const Spacer(),
                Text('${m['state'] ?? '—'}', style: TextStyle(fontSize: 12, color: ok ? RoyalTheme.green : RoyalTheme.red)),
              ]),
            );
          }),
      ]),
    );
  }

  Widget _buildParanoidXCard() {
    final p = _paranoidx ?? {};
    final layers = p['layers'] as List? ?? [];
    final overall = p['overall'] ?? p['status'] ?? 'unknown';
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: RoyalTheme.glowBorder(color: RoyalTheme.purple),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Icon(Icons.shield, color: RoyalTheme.purple, size: 18),
          const SizedBox(width: 8),
          Text('ParanoidX Security', style: Theme.of(context).textTheme.titleMedium),
          const Spacer(),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
            decoration: BoxDecoration(
              color: (overall == 'healthy' ? RoyalTheme.green : RoyalTheme.orange).withAlpha(30),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Text('$overall', style: TextStyle(fontSize: 11, color: overall == 'healthy' ? RoyalTheme.green : RoyalTheme.orange)),
          ),
        ]),
        const SizedBox(height: 12),
        if (layers.isEmpty)
          Text('No layer data', style: Theme.of(context).textTheme.bodyMedium)
        else
          ...layers.map((l) {
            final m = l is Map<String, dynamic> ? l : {'name': '$l'};
            final ok = m['healthy'] == true || m['status'] == 'healthy';
            return Padding(
              padding: const EdgeInsets.symmetric(vertical: 3),
              child: Row(children: [
                Icon(ok ? Icons.check_circle : Icons.error, size: 14, color: ok ? RoyalTheme.green : RoyalTheme.red),
                const SizedBox(width: 8),
                Text('${m['name'] ?? '—'}', style: const TextStyle(fontSize: 13)),
                const Text('  :10810', style: TextStyle(fontSize: 11, color: Color(0xFF8B949E))),
              ]),
            );
          }),
      ]),
    );
  }

  Widget _buildCheckList() {
    final checks = (_checks?['checks'] as List?) ?? [];
    if (checks.isEmpty) return const SizedBox.shrink();
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: RoyalTheme.gradientCard(),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text('Health Checks (${checks.length})', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 12),
        ...checks.map((c) {
          final m = c is Map<String, dynamic> ? c : {'name': '$c'};
          final ok = m['healthy'] == true || m['ok'] == true || m['status'] == 'healthy';
          return Padding(
            padding: const EdgeInsets.symmetric(vertical: 3),
            child: Row(children: [
              Icon(ok ? Icons.check_circle_outline : Icons.error_outline, size: 14, color: ok ? RoyalTheme.green : RoyalTheme.red),
              const SizedBox(width: 8),
              Text('${m['name'] ?? '—'}', style: const TextStyle(fontSize: 13)),
              const Spacer(),
              Text('${m['status'] ?? m['value'] ?? ''}', style: TextStyle(fontSize: 11, color: ok ? RoyalTheme.green : RoyalTheme.red)),
            ]),
          );
        }),
      ]),
    );
  }
}
