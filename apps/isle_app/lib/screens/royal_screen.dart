import 'dart:async';
import 'package:flutter/material.dart';
import 'package:api_client/api_client.dart';
import 'package:http/http.dart' as http;
import 'package:intl/intl.dart';

/// RoyalScreen manages the royal mesh node registry and command interface.
class RoyalScreen extends StatefulWidget {
  final RoyalClient royal;
  final String pubkey;
  final http.Client httpClient;
  const RoyalScreen({super.key, required this.royal, required this.pubkey, required this.httpClient});
  @override
  State<RoyalScreen> createState() => _RoyalScreenState();
}

class _RoyalScreenState extends State<RoyalScreen> {
  List<dynamic> _nodes = [];
  String _royalPubkey = '';
  bool _loading = true;
  String? _error;
  Timer? _refreshTimer;

  @override
  void initState() {
    super.initState();
    _load();
    _refreshTimer = Timer.periodic(const Duration(seconds: 30), (_) => _load());
  }

  @override
  void dispose() {
    _refreshTimer?.cancel();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final data = await widget.royal.nodes();
      if (mounted) {
        setState(() {
          _nodes = data['nodes'] as List<dynamic>? ?? [];
          _royalPubkey = data['royal_pubkey'] as String? ?? '';
          _loading = false;
          _error = null;
        });
      }
    } catch (e) {
      if (mounted) setState(() { _error = '$e'; _loading = false; });
    }
  }

  Color _statusColor(String status) {
    switch (status) {
      case 'online': return Colors.green;
      case 'stale': return Colors.orange;
      case 'offline': return Colors.red;
      default: return Colors.grey;
    }
  }

  void _showRegisterDialog() {
    final pkCtrl = TextEditingController();
    final labelCtrl = TextEditingController();
    final addrCtrl = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Register Sub-Node'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: pkCtrl, decoration: const InputDecoration(labelText: 'Node Pubkey', border: OutlineInputBorder(), isDense: true)),
            const SizedBox(height: 8),
            TextField(controller: labelCtrl, decoration: const InputDecoration(labelText: 'Label', border: OutlineInputBorder(), isDense: true)),
            const SizedBox(height: 8),
            TextField(controller: addrCtrl, decoration: const InputDecoration(labelText: 'Address (SMP/onion)', border: OutlineInputBorder(), isDense: true)),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')),
          FilledButton(onPressed: () async {
            Navigator.pop(ctx);
            try {
              await widget.royal.register(pkCtrl.text.trim(), labelCtrl.text.trim(), addrCtrl.text.trim());
              _load();
              if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Node registered'), backgroundColor: Colors.green));
            } catch (e) {
              if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red));
            }
          }, child: const Text('Register')),
        ],
      ),
    );
  }

  void _showSendCommandDialog() {
    final cmdCtrl = TextEditingController();
    final targetCtrl = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Send Command'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: targetCtrl, decoration: const InputDecoration(labelText: 'Target Node Pubkey', border: OutlineInputBorder(), isDense: true)),
            const SizedBox(height: 8),
            TextField(controller: cmdCtrl, decoration: const InputDecoration(labelText: 'Command', border: OutlineInputBorder(), isDense: true), maxLines: 3),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')),
          FilledButton(onPressed: () async {
            Navigator.pop(ctx);
            try {
              final result = await widget.royal.sendCommand(cmdCtrl.text.trim(), targetCtrl.text.trim());
              if (mounted) {
                showDialog(
                  context: context,
                  builder: (ctx) => AlertDialog(
                    title: const Text('Command Result'),
                    content: SelectableText(result.toString()),
                    actions: [FilledButton(onPressed: () => Navigator.pop(ctx), child: const Text('OK'))],
                  ),
                );
              }
            } catch (e) {
              if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red));
            }
          }, child: const Text('Send')),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    if (_loading) return const Center(child: CircularProgressIndicator());
    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.error_outline, size: 48, color: theme.colorScheme.error),
            const SizedBox(height: 12),
            Text('Connection error', style: theme.textTheme.titleMedium),
            const SizedBox(height: 4),
            Text(_error!, style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
            const SizedBox(height: 16),
            FilledButton.icon(icon: const Icon(Icons.refresh), label: const Text('Retry'), onPressed: () { setState(() { _loading = true; _error = null; }); _load(); }),
          ],
        ),
      );
    }
    return Column(
      children: [
        // Header
        Container(
          width: double.infinity,
          padding: const EdgeInsets.fromLTRB(16, 10, 16, 8),
          child: Row(
            children: [
              const Icon(Icons.account_tree, size: 20),
              const SizedBox(width: 8),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Royal Mesh', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
                    if (_royalPubkey.isNotEmpty)
                      Text('Key: ${_royalPubkey.substring(0, 16)}...', style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey, fontFamily: 'monospace', fontSize: 10)),
                  ],
                ),
              ),
              IconButton(icon: const Icon(Icons.refresh, size: 20), onPressed: _load, tooltip: 'Refresh'),
            ],
          ),
        ),
        const Divider(height: 1),

        // Action bar
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
          child: Row(
            children: [
              FilledButton.tonalIcon(
                icon: const Icon(Icons.add, size: 18),
                label: const Text('Register Node', style: TextStyle(fontSize: 13)),
                onPressed: _showRegisterDialog,
              ),
              const SizedBox(width: 8),
              OutlinedButton.icon(
                icon: const Icon(Icons.send, size: 18),
                label: const Text('Send Command', style: TextStyle(fontSize: 13)),
                onPressed: _showSendCommandDialog,
              ),
              const Spacer(),
              Text('${_nodes.length} node${_nodes.length == 1 ? '' : 's'}', style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
            ],
          ),
        ),
        const Divider(height: 1),

        // Node list
        Expanded(
          child: _nodes.isEmpty
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.hub, size: 64, color: Colors.grey[400]),
                      const SizedBox(height: 12),
                      Text('No nodes registered', style: theme.textTheme.titleSmall?.copyWith(color: Colors.grey)),
                      const SizedBox(height: 4),
                      Text('Register a sub-node to start building the mesh', style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
                    ],
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: _nodes.length,
                  itemBuilder: (ctx, i) {
                    final node = _nodes[i] as Map<String, dynamic>;
                    final pubkey = node['pubkey'] as String? ?? '';
                    final label = node['label'] as String? ?? '';
                    final addr = node['addr'] as String? ?? '';
                    final status = node['status'] as String? ?? 'unknown';
                    final lastSeen = node['last_seen'] as String? ?? '';
                    final registered = node['registered'] as String? ?? '';
                    return Card(
                      margin: const EdgeInsets.only(bottom: 8),
                      child: Padding(
                        padding: const EdgeInsets.all(14),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Container(
                                  width: 12, height: 12,
                                  decoration: BoxDecoration(shape: BoxShape.circle, color: _statusColor(status)),
                                ),
                                const SizedBox(width: 10),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(label.isNotEmpty ? label : 'Unnamed Node', style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w600)),
                                      if (addr.isNotEmpty)
                                        Text(addr, style: const TextStyle(fontSize: 10, fontFamily: 'monospace', color: Colors.grey)),
                                    ],
                                  ),
                                ),
                                Chip(
                                  label: Text(status.toUpperCase(), style: TextStyle(fontSize: 10, color: _statusColor(status), fontWeight: FontWeight.w600)),
                                  backgroundColor: _statusColor(status).withOpacity(0.1),
                                  side: BorderSide.none,
                                  visualDensity: VisualDensity.compact,
                                  padding: EdgeInsets.zero,
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),
                            Row(
                              children: [
                                Icon(Icons.vpn_key, size: 12, color: Colors.grey),
                                const SizedBox(width: 4),
                                Expanded(
                                  child: Text(pubkey.length > 32 ? '${pubkey.substring(0, 32)}...' : pubkey,
                                      style: const TextStyle(fontSize: 9, fontFamily: 'monospace', color: Colors.grey)),
                                ),
                              ],
                            ),
                            if (lastSeen.isNotEmpty) ...[
                              const SizedBox(height: 4),
                              Row(
                                children: [
                                  Icon(Icons.access_time, size: 12, color: Colors.grey),
                                  const SizedBox(width: 4),
                                  Text('Last seen: ${_formatTime(lastSeen)}',
                                      style: TextStyle(fontSize: 10, color: Colors.grey[600])),
                                ],
                              ),
                            ],
                          ],
                        ),
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }

  String _formatTime(String iso) {
    try {
      final dt = DateTime.parse(iso);
      final now = DateTime.now().toUtc();
      final diff = now.difference(dt);
      if (diff.inMinutes < 1) return 'just now';
      if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
      if (diff.inHours < 24) return '${diff.inHours}h ago';
      return DateFormat('MMM d HH:mm').format(dt.toLocal());
    } catch (_) {
      return iso;
    }
  }
}
