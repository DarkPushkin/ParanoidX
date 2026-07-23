import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:models/models.dart';
import 'package:widgets/widgets.dart';
import '../services/isle_api_service.dart';
import '../widgets/nt_display.dart';

/// WalletScreen manages the wallet interface for balance, transfers, holdings and silver coins.
class WalletScreen extends StatefulWidget {
  final IsleApiService api;
  final String serverUrl;
  final http.Client httpClient;
  const WalletScreen({super.key, required this.api, required this.serverUrl, required this.httpClient});

  @override
  State<WalletScreen> createState() => _WalletScreenState();
}

class _WalletScreenState extends State<WalletScreen> {
  Map<String, dynamic>? _balance;
  Map<String, dynamic>? _holdings;
  List<Map<String, dynamic>> _myCoins = [];
  final _toCtrl = TextEditingController();
  final _amountCtrl = TextEditingController();
  String? _cachedPubkey;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<Map<String, dynamic>> _get(String path) async {
    final uri = Uri.parse('${widget.serverUrl}$path');
    final resp = await widget.httpClient.get(uri).timeout(const Duration(seconds: 10));
    if (resp.statusCode != 200) throw Exception('${resp.statusCode}: ${resp.body}');
    return jsonDecode(resp.body) as Map<String, dynamic>;
  }

  Future<String> _ensurePubkey() async {
    if (_cachedPubkey != null) return _cachedPubkey!;
    final d = await _get('/api/wallet/create');
    _cachedPubkey = d['pubkey'] as String;
    return _cachedPubkey!;
  }

  Future<void> _load() async {
    try {
      final b = await widget.api.getBalance();
      final h = await widget.api.getHoldings();
      List<Map<String, dynamic>> coins = [];
      try {
        final pk = await _ensurePubkey();
        final mc = await _get('/api/silver/my-coins?pubkey=$pk');
        coins = (mc['coins'] as List<dynamic>?)?.cast<Map<String, dynamic>>() ?? [];
      } catch (_) {}
      setState(() { _balance = b; _holdings = h; _myCoins = coins; });
    } catch (_) {}
  }

  Future<void> _transfer() async {
    final to = _toCtrl.text.trim();
    final amount = int.tryParse(_amountCtrl.text.trim());
    if (to.isEmpty || amount == null || amount <= 0) return;
    try {
      await widget.api.transfer(to, amount);
      _toCtrl.clear();
      _amountCtrl.clear();
      _load();
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Transfer sent')),
      );
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error: $e')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final body = RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(12),
        children: [
          // Balance
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.account_balance_wallet, color: Theme.of(context).colorScheme.primary),
                      const SizedBox(width: 8),
                      Text('Balance', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
                    ],
                  ),
                  const SizedBox(height: 12),
                  _balance == null
                      ? const Text('Loading...')
                      : Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            NtDisplay(amountNt: _balance!['liquid_balance_ng'] ?? 0),
                            const SizedBox(height: 4),
                            Text('Frozen: ${_balance!['banknotes_count'] ?? 0} banknotes',
                                style: Theme.of(context).textTheme.bodySmall?.copyWith(color: Colors.grey)),
                          ],
                        ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 10),

          // Transfer
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.send, color: Theme.of(context).colorScheme.primary),
                      const SizedBox(width: 8),
                      Text('Transfer', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
                    ],
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: _toCtrl,
                    decoration: const InputDecoration(labelText: 'Recipient pubkey', border: OutlineInputBorder()),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: _amountCtrl,
                    decoration: const InputDecoration(labelText: 'Amount (NT)', border: OutlineInputBorder()),
                    keyboardType: TextInputType.number,
                  ),
                  const SizedBox(height: 12),
                  FilledButton.icon(
                    onPressed: _transfer,
                    icon: const Icon(Icons.send, size: 16),
                    label: const Text('Send'),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 10),

          // Holdings
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.inventory_2, color: Theme.of(context).colorScheme.primary),
                      const SizedBox(width: 8),
                      Text('Holdings', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
                    ],
                  ),
                  const SizedBox(height: 12),
                  _holdings == null
                      ? const Text('Loading...')
                      : _holdings!.containsKey('liquid_ng')
                          ? Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                NtDisplay(amountNt: (_holdings!['liquid_ng'] as num?)?.toInt() ?? 0),
                                Text('Frozen: ${(_holdings!['frozen_ng'] as num?)?.toInt() ?? 0} nt',
                                    style: Theme.of(context).textTheme.bodySmall?.copyWith(color: Colors.grey)),
                                const Divider(),
                                ...(_holdings!['banknotes'] as List? ?? [])
                                    .map((bn) => Padding(
                                          padding: const EdgeInsets.symmetric(vertical: 2),
                                          child: Text('${bn['serial']} — ${bn['rarity']} (${bn['denomination_ng']} nt)',
                                              style: const TextStyle(fontFamily: 'monospace', fontSize: 12)),
                                        )),
                              ],
                            )
                          : Text('No holdings data'),
                ],
              ),
            ),
          ),
          const SizedBox(height: 10),

          // Silver Coins
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.monetization_on, color: Colors.amber.shade700),
                      const SizedBox(width: 8),
                      Text('Silver Coins', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
                      const Spacer(),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                        decoration: BoxDecoration(
                          color: _myCoins.isNotEmpty ? Colors.amber.shade100 : Colors.grey.shade100,
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Text('${_myCoins.length}',
                            style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold,
                                color: _myCoins.isNotEmpty ? Colors.brown.shade700 : Colors.grey)),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  if (_myCoins.isEmpty)
                    Text('No silver coins yet. Buy one in the Market tab.',
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(color: Colors.grey))
                  else
                    ..._myCoins.map((c) => Padding(
                          padding: const EdgeInsets.symmetric(vertical: 4),
                          child: Row(
                            children: [
                              Icon(Icons.circle, size: 10, color: Colors.amber.shade600),
                              const SizedBox(width: 8),
                              Text('${c['serial']}',
                                  style: const TextStyle(fontFamily: 'monospace', fontSize: 12, fontWeight: FontWeight.bold)),
                              const Spacer(),
                              Text('${c['price_nt']} nt',
                                  style: TextStyle(fontSize: 11, color: Colors.grey.shade600)),
                            ],
                          ),
                        )),
                ],
              ),
            ),
          ),
        ],
      ),
    );

    return Scaffold(
      appBar: AppBar(title: const Text('Wallet — NT')),
      body: body,
    );
  }

  @override
  void dispose() {
    _toCtrl.dispose();
    _amountCtrl.dispose();
    super.dispose();
  }
}
