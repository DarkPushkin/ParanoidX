import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../theme.dart';
import '../main.dart';

/// DashboardScreen manages the server dashboard with status, system, treasury and radio tabs.
class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key});
  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  Map<String, dynamic>? _version, _health, _economy, _alerts, _ping;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    _refresh();
    _timer = Timer.periodic(const Duration(seconds: 15), (_) => _refresh());
  }

  @override
  void dispose() { _timer?.cancel(); super.dispose(); }

  Future<void> _refresh() async {
    final api = context.read<AppState>().api;
    try {
      final results = await Future.wait([
        api.getVersion(), api.getHealth(), api.getEconomyReport(),
        api.getAlertRules(), api.ping(),
      ], eagerError: false);
      if (mounted) setState(() {
        _version = results[0];
        _health = results[1];
        _economy = results[2];
        _alerts = results[3];
        _ping = results[4];
      });
      context.read<AppState>().setOffline(false);
    } catch (_) {
      context.read<AppState>().setOffline(true);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Isle Royal Dashboard'),
        actions: [
          if (_ping?['pong'] == true)
            Container(
              margin: const EdgeInsets.only(right: 12),
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: RoyalTheme.green.withAlpha(30),
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: RoyalTheme.green.withAlpha(80)),
              ),
              child: Row(mainAxisSize: MainAxisSize.min, children: [
                Container(width: 8, height: 8, decoration: const BoxDecoration(shape: BoxShape.circle, color: RoyalTheme.green)),
                const SizedBox(width: 6),
                Text('Online', style: TextStyle(fontSize: 12, color: RoyalTheme.green)),
              ]),
            )
          else
            Container(
              margin: const EdgeInsets.only(right: 12),
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: RoyalTheme.red.withAlpha(30),
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: RoyalTheme.red.withAlpha(80)),
              ),
              child: Row(mainAxisSize: MainAxisSize.min, children: [
                Container(width: 8, height: 8, decoration: const BoxDecoration(shape: BoxShape.circle, color: RoyalTheme.red)),
                const SizedBox(width: 6),
                Text('Offline', style: TextStyle(fontSize: 12, color: RoyalTheme.red)),
              ]),
            ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _refresh,
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [
            _buildBuildCard(),
            const SizedBox(height: 16),
            _buildHealthGrid(),
            const SizedBox(height: 16),
            _buildEconomyCard(),
            const SizedBox(height: 16),
            _buildAlertCard(),
          ],
        ),
      ),
    );
  }

  Widget _buildBuildCard() {
    final ver = _version?['build'] ?? '—';
    final uptime = _health?['uptime_hours'] ?? 0;
    final started = _version?['started'] ?? '—';
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: RoyalTheme.glowBorder(),
      child: Row(children: [
        Expanded(
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text('Sovereign Node', style: Theme.of(context).textTheme.titleLarge),
            const SizedBox(height: 4),
            Text(ver, style: TextStyle(color: RoyalTheme.gold, fontWeight: FontWeight.w600)),
            const SizedBox(height: 2),
            Text('Started: $started', style: Theme.of(context).textTheme.bodyMedium),
          ]),
        ),
        Column(children: [
          Text('${uptime}h', style: TextStyle(fontSize: 36, fontWeight: FontWeight.w700, color: RoyalTheme.gold)),
          Text('uptime', style: Theme.of(context).textTheme.bodyMedium),
        ]),
      ]),
    );
  }

  Widget _buildHealthGrid() {
    final h = _health ?? {};
    final bridgeOk = h['bridge'] == true || h['bridge_connected'] == true;
    final healthy = h['healthy'] ?? false;
    final msgCount = h['message_count'] ?? h['total_messages'] ?? 0;
    return Row(children: [
      Expanded(child: _statCard('Server', healthy == true ? 'Healthy' : 'Degraded', healthy == true ? RoyalTheme.green : RoyalTheme.orange, Icons.check_circle)),
      const SizedBox(width: 12),
      Expanded(child: _statCard('Bridge', bridgeOk ? 'Connected' : 'Disconnected', bridgeOk ? RoyalTheme.green : RoyalTheme.red, Icons.wifi)),
      const SizedBox(width: 12),
      Expanded(child: _statCard('Messages', '$msgCount', RoyalTheme.accent, Icons.email)),
    ]);
  }

  Widget _statCard(String label, String value, Color color, IconData icon) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: RoyalTheme.gradientCard(),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Icon(icon, size: 18, color: color),
          const SizedBox(width: 6),
          Text(label, style: Theme.of(context).textTheme.bodyMedium),
        ]),
        const SizedBox(height: 8),
        Text(value, style: TextStyle(fontSize: 20, fontWeight: FontWeight.w700, color: color)),
      ]),
    );
  }

  Widget _buildEconomyCard() {
    final e = _economy ?? {};
    final reserve = e['reserve_ng'] ?? e['reserve'] ?? 0;
    final supply = e['supply_ng'] ?? e['supply'] ?? e['total_supply'] ?? 0;
    final price = e['oracle_price'] ?? e['price'] ?? 0;
    final ratio = supply is int && supply > 0
        ? (reserve is int ? (reserve / supply * 100).toStringAsFixed(1) : '0')
        : '100';
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: RoyalTheme.gradientCard(),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Icon(Icons.account_balance, size: 20, color: RoyalTheme.gold),
          const SizedBox(width: 8),
          Text('Treasury Overview', style: Theme.of(context).textTheme.titleLarge),
        ]),
        const SizedBox(height: 16),
        Row(children: [
          Expanded(child: _metric('Reserve', '${reserve} ng', RoyalTheme.gold)),
          Expanded(child: _metric('Supply', '${supply} ng', RoyalTheme.silver)),
          Expanded(child: _metric('Price', '\$$price', RoyalTheme.green)),
          Expanded(child: _metric('Backing', '$ratio%', RoyalTheme.accent)),
        ]),
      ]),
    );
  }

  Widget _metric(String label, String value, Color color) {
    return Column(children: [
      Text(value, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: color)),
      const SizedBox(height: 2),
      Text(label, style: Theme.of(context).textTheme.bodyMedium),
    ]);
  }

  Widget _buildAlertCard() {
    final rules = _alerts?['rules'] as List? ?? [];
    final count = rules.length;
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: RoyalTheme.gradientCard(colors: [RoyalTheme.darkCard, const Color(0xFF1E1515)]),
      child: Row(children: [
        Icon(Icons.notifications_active, color: RoyalTheme.orange, size: 24),
        const SizedBox(width: 12),
        Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text('$count Alert Rule${count == 1 ? '' : 's'} Active',
              style: Theme.of(context).textTheme.titleMedium),
          Text('Monitoring treasury, bridge & system health',
              style: Theme.of(context).textTheme.bodyMedium),
        ])),
        Icon(Icons.chevron_right, color: RoyalTheme.gold),
      ]),
    );
  }
}
