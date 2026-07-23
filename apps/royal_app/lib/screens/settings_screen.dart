import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../theme.dart';
import '../main.dart';

/// SettingsScreen manages settingsscreen functionality.
class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});
  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  Map<String, dynamic>? _info, _emStop, _rateLimit;
  Timer? _timer;

  @override
  void initState() { super.initState(); _refresh(); }
  @override
  void dispose() { _timer?.cancel(); super.dispose(); }

  Future<void> _refresh() async {
    final api = context.read<AppState>().api;
    try {
      final r = await Future.wait([
        api.getInfo(), api.getEmergencyStop(), api.getRateLimitStats(),
      ], eagerError: false);
      if (mounted) setState(() { _info = r[0]; _emStop = r[1]; _rateLimit = r[2]; });
    } catch (_) {}
  }

  Future<void> _toggleEmergencyStop() async {
    final current = _emStop?['emergency_stop'] == true;
    try {
      await context.read<AppState>().api.setEmergencyStop(!current);
      _refresh();
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(
        content: Text('Emergency Stop ${current ? 'DISABLED' : 'ENABLED'}'),
        backgroundColor: (!current ? RoyalTheme.red : RoyalTheme.green).withAlpha(40),
      ));
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e')));
    }
  }

  @override
  Widget build(BuildContext context) {
    final info = _info ?? {};
    final emStop = _emStop?['emergency_stop'] == true;
    final rl = _rateLimit ?? {};
    return Scaffold(
      appBar: AppBar(
        title: Row(children: [
          Container(padding: const EdgeInsets.all(6), decoration: BoxDecoration(color: RoyalTheme.silver.withAlpha(30), borderRadius: BorderRadius.circular(8)),
            child: const Icon(Icons.settings, color: RoyalTheme.silver, size: 20)),
          const SizedBox(width: 12),
          const Text('Settings'),
        ]),
      ),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        Container(
          padding: const EdgeInsets.all(16),
          decoration: RoyalTheme.glowBorder(),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Row(children: [
              Container(width: 40, height: 40, decoration: BoxDecoration(shape: BoxShape.circle, border: Border.all(color: RoyalTheme.gold, width: 2)),
                child: Center(child: Text('♚', style: TextStyle(fontSize: 20, color: RoyalTheme.gold)))),
              const SizedBox(width: 12),
              Text('Royal Node', style: Theme.of(context).textTheme.titleLarge),
            ]),
            const SizedBox(height: 12),
            _kv('Build', info['build'] ?? info['version'] ?? '—'),
            _kv('Go', info['go'] ?? '—'),
            _kv('Started', info['started'] ?? '—'),
            _kv('Data Dir', info['data_dir'] ?? '—'),
          ]),
        ),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: (emStop ? RoyalTheme.red : RoyalTheme.darkCard).withAlpha(emStop ? 20 : 0),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: emStop ? RoyalTheme.red.withAlpha(80) : const Color(0xFF30363D)),
          ),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Row(children: [
              Icon(Icons.warning_amber, color: emStop ? RoyalTheme.red : RoyalTheme.gold, size: 20),
              const SizedBox(width: 8),
              Text('Emergency Stop', style: Theme.of(context).textTheme.titleMedium),
              const Spacer(),
              Switch(
                value: emStop,
                activeThumbColor: RoyalTheme.red,
                onChanged: (_) => _toggleEmergencyStop(),
              ),
            ]),
            Text(emStop
              ? 'ALL treasury operations are SUSPENDED'
              : 'Toggle to halt all treasury mint/burn operations',
              style: TextStyle(fontSize: 12, color: emStop ? RoyalTheme.red : const Color(0xFF8B949E))),
          ]),
        ),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: RoyalTheme.gradientCard(),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Row(children: [
              Icon(Icons.speed, color: RoyalTheme.accent, size: 18),
              const SizedBox(width: 8),
              Text('Rate Limits', style: Theme.of(context).textTheme.titleMedium),
            ]),
            const SizedBox(height: 8),
            _kv('Max RPM', '${rl['max_rpm'] ?? '—'}'),
            _kv('Endpoints', '${rl['endpoints'] ?? '—'}'),
            _kv('Policy', rl['policy'] ?? '—'),
          ]),
        ),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: RoyalTheme.gradientCard(),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Row(children: [
              Icon(Icons.history, color: RoyalTheme.silver, size: 18),
              const SizedBox(width: 8),
              Text('Audit Export', style: Theme.of(context).textTheme.titleMedium),
            ]),
            const SizedBox(height: 8),
            Text('Download full audit log as JSON or CSV',
              style: Theme.of(context).textTheme.bodyMedium),
          ]),
        ),
      ]),
    );
  }

  Widget _kv(String k, String v) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3),
      child: Row(children: [
        Text('$k: ', style: const TextStyle(color: Color(0xFF8B949E), fontSize: 13)),
        Text(v, style: const TextStyle(color: Colors.white, fontSize: 13)),
      ]),
    );
  }
}
