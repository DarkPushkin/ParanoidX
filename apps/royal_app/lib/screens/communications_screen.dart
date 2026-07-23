import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../theme.dart';
import '../main.dart';

/// CommunicationsScreen manages communicationsscreen functionality.
class CommunicationsScreen extends StatefulWidget {
  const CommunicationsScreen({super.key});
  @override
  State<CommunicationsScreen> createState() => _CommunicationsScreenState();
}

class _CommunicationsScreenState extends State<CommunicationsScreen> {
  final _msgCtrl = TextEditingController();
  Map<String, dynamic>? _status;
  Timer? _timer;

  @override
  void initState() { super.initState(); _refresh(); _timer = Timer.periodic(const Duration(seconds: 30), (_) => _refresh()); }
  @override
  void dispose() { _timer?.cancel(); _msgCtrl.dispose(); super.dispose(); }

  Future<void> _refresh() async {
    try {
      final s = await context.read<AppState>().api.chatStatus();
      if (mounted) setState(() => _status = s);
    } catch (_) {}
  }

  Future<void> _broadcast() async {
    if (_msgCtrl.text.trim().isEmpty) return;
    try {
      final res = await context.read<AppState>().api.chatBroadcast(_msgCtrl.text);
      if (mounted) {
        _msgCtrl.clear();
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(
          content: Text(res['ok'] == true ? 'Broadcast sent' : 'Failed'),
          backgroundColor: res['ok'] == true ? RoyalTheme.green.withAlpha(40) : RoyalTheme.red.withAlpha(40),
        ));
      }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e'), backgroundColor: RoyalTheme.red.withAlpha(40)));
    }
  }

  @override
  Widget build(BuildContext context) {
    final connected = _status?['bridge_connected'] == true || _status?['connected'] == true;
    final msgCount = _status?['message_count'] ?? _status?['total_messages'] ?? 0;
    return Scaffold(
      appBar: AppBar(
        title: Row(children: [
          Container(padding: const EdgeInsets.all(6), decoration: BoxDecoration(color: RoyalTheme.teal.withAlpha(30), borderRadius: BorderRadius.circular(8)),
            child: const Icon(Icons.chat, color: RoyalTheme.teal, size: 20)),
          const SizedBox(width: 12),
          const Text('Communications'),
        ]),
        actions: [
          Container(
            margin: const EdgeInsets.only(right: 12),
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            decoration: BoxDecoration(
              color: (connected ? RoyalTheme.green : RoyalTheme.red).withAlpha(30),
              borderRadius: BorderRadius.circular(20),
            ),
            child: Text(connected ? 'Bridge Online' : 'Bridge Offline',
              style: TextStyle(fontSize: 11, color: connected ? RoyalTheme.green : RoyalTheme.red, fontWeight: FontWeight.w600)),
          ),
        ],
      ),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        Container(
          padding: const EdgeInsets.all(16),
          decoration: RoyalTheme.gradientCard(),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text('SimpleX Bridge', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            Row(children: [
              _miniStat('Messages', '$msgCount', RoyalTheme.accent),
              const SizedBox(width: 24),
              _miniStat('Status', connected ? 'Connected' : 'Disconnected', connected ? RoyalTheme.green : RoyalTheme.red),
            ]),
          ]),
        ),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: RoyalTheme.glowBorder(color: RoyalTheme.teal),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Row(children: [
              Icon(Icons.campaign, color: RoyalTheme.orange, size: 20),
              const SizedBox(width: 8),
              Text('Royal Broadcast', style: Theme.of(context).textTheme.titleMedium),
            ]),
            const SizedBox(height: 12),
            TextField(
              controller: _msgCtrl,
              maxLines: 4,
              decoration: const InputDecoration(
                hintText: 'Type your broadcast message to all contacts...',
              ),
            ),
            const SizedBox(height: 12),
            Align(
              alignment: Alignment.centerRight,
              child: ElevatedButton.icon(
                icon: const Icon(Icons.send, size: 18),
                label: const Text('Broadcast'),
                onPressed: _broadcast,
              ),
            ),
          ]),
        ),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: RoyalTheme.gradientCard(),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Row(children: [
              Icon(Icons.notifications, color: RoyalTheme.gold, size: 18),
              const SizedBox(width: 8),
              Text('Treasury Alert', style: Theme.of(context).textTheme.titleMedium),
            ]),
            const SizedBox(height: 8),
            Text('Send an urgent treasury notification to all SimpleX contacts.',
              style: Theme.of(context).textTheme.bodyMedium),
          ]),
        ),
      ]),
    );
  }

  Widget _miniStat(String label, String value, Color color) {
    return Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Text(value, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: color)),
      Text(label, style: Theme.of(context).textTheme.bodyMedium),
    ]);
  }
}
