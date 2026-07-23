import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:http/http.dart' as http;
import '../widgets/nt_display.dart';

/// MarketScreen manages the silver coin marketplace for buying and redeeming coins.
class MarketScreen extends StatefulWidget {
  final String serverUrl;
  final http.Client httpClient;
  const MarketScreen({super.key, required this.serverUrl, required this.httpClient});

  @override
  State<MarketScreen> createState() => _MarketScreenState();
}

class _MarketScreenState extends State<MarketScreen> {
  Map<String, dynamic> _shopData = {};
  List<Map<String, dynamic>> _coins = [];
  List<Map<String, dynamic>> _myCoins = [];
  bool _loading = true;
  String? _error;

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

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final pk = await _ensurePubkey();
      final results = await Future.wait([
        _get('/api/silver/shop'),
        _get('/api/silver/my-coins?pubkey=$pk'),
      ]);
      final shopResult = results[0];
      final myResult = results[1];
      setState(() {
        _shopData = shopResult;
        _coins = (shopResult['coins'] as List<dynamic>?)?.cast<Map<String, dynamic>>() ?? [];
        _myCoins = (myResult['coins'] as List<dynamic>?)?.cast<Map<String, dynamic>>() ?? [];
        _loading = false;
        _error = null;
      });
    } catch (e) {
      setState(() {
        _loading = false;
        _error = e.toString();
      });
    }
  }

  String? _cachedPubkey;

  Future<String> _ensurePubkey() async {
    if (_cachedPubkey != null) return _cachedPubkey!;
    final uri = Uri.parse('${widget.serverUrl}/api/wallet/create');
    final resp = await widget.httpClient.get(uri).timeout(const Duration(seconds: 10));
    if (resp.statusCode != 200) throw Exception('wallet create failed: ${resp.body}');
    final d = jsonDecode(resp.body) as Map<String, dynamic>;
    _cachedPubkey = d['pubkey'] as String;
    return _cachedPubkey!;
  }

  Future<void> _buyCoin(String coinId) async {
    try {
      final pk = await _ensurePubkey();
      final uri = Uri.parse('${widget.serverUrl}/api/silver/buy');
      final resp = await widget.httpClient
          .post(uri,
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'coin_id': coinId, 'buyer': pk}))
          .timeout(const Duration(seconds: 15));
      final d = jsonDecode(resp.body) as Map<String, dynamic>;
      if (d['ok'] != true) throw Exception(d['error'] ?? 'buy failed');
      _load();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('✅ ${d['coin']['serial']} purchased for ${d['coin']['price_nt']} nt')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('❌ $e')),
        );
      }
    }
  }

  Future<void> _redeemCoin(String coinId) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Redeem Coin'),
        content: const Text(
          'Redeeming this coin will return 99% of its NT value to your wallet (1% fee). '
          'The coin will return to the shop inventory.',
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Cancel')),
          FilledButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Redeem')),
        ],
      ),
    );
    if (confirm != true) return;
    try {
      final uri = Uri.parse('${widget.serverUrl}/api/silver/redeem');
      final resp = await widget.httpClient
          .post(uri,
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'coin_id': coinId}))
          .timeout(const Duration(seconds: 15));
      final d = jsonDecode(resp.body) as Map<String, dynamic>;
      if (d['ok'] != true) throw Exception(d['error'] ?? 'redeem failed');
      _load();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('✅ Coin redeemed — ${d['refund_nt']} NT returned')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('❌ $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final body = _loading
        ? const Center(child: CircularProgressIndicator())
        : _error != null
            ? Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.cloud_off, size: 48, color: theme.colorScheme.error),
                    const SizedBox(height: 8),
                    Text('$_error', style: TextStyle(color: theme.colorScheme.error)),
                    const SizedBox(height: 16),
                    FilledButton.icon(
                      onPressed: _load,
                      icon: const Icon(Icons.refresh),
                      label: const Text('Retry'),
                    ),
                  ],
                ),
              )
            : _buildShop(theme);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Silver Coin Shop'),
        actions: [
          IconButton(icon: const Icon(Icons.refresh), onPressed: _load),
        ],
      ),
      body: body,
    );
  }

  Widget _buildShop(ThemeData theme) {
    final reserve = _shopData['reserve_ng'] as int? ?? 0;
    final ratio = (_shopData['backing_ratio'] as num?)?.toDouble() ?? 0;
    final spot = (_shopData['silver_spot_usd'] as num?)?.toDouble() ?? 0;
    final ntPerTlr = _shopData['nt_per_tlr'] as int? ?? NtDisplay.ntPerTlr;

    return RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // ── Reserve Banner ──
          Card(
            color: ratio >= 0.7 ? Colors.green.shade50 : Colors.orange.shade50,
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.security, color: ratio >= 0.7 ? Colors.green : Colors.orange),
                      const SizedBox(width: 8),
                      Text('Silver Reserve', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
                      const Spacer(),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                        decoration: BoxDecoration(
                          color: ratio >= 0.7 ? Colors.green : Colors.orange,
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Text(
                          '${(ratio * 100).toStringAsFixed(1)}%',
                          style: const TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.bold),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  NtDisplay(amountNt: reserve),
                  const SizedBox(height: 4),
                  Text('Silver spot: \$${spot.toStringAsFixed(2)}/oz',
                      style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),

          // ── Island Silver Coins ──
          Row(
            children: [
              Icon(Icons.store, color: theme.colorScheme.primary),
              const SizedBox(width: 8),
              Text('Island Silver Coins', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: Colors.amber.shade100,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text('${_coins.length} available',
                    style: TextStyle(fontSize: 12, color: Colors.brown.shade700)),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            'Each physical silver coin = 1 TLR (1 troy oz). '
            'Backed by island silver reserves. Tokenized into Nanotalers (NT).',
            style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey),
          ),
          const SizedBox(height: 12),

          if (_coins.isEmpty)
            Card(
              child: Padding(
                padding: const EdgeInsets.all(32),
                child: Column(
                  children: [
                    Icon(Icons.inventory_2_outlined, size: 64, color: Colors.grey.shade300),
                    const SizedBox(height: 16),
                    Text('No coins currently minted', style: theme.textTheme.bodyLarge),
                    const SizedBox(height: 8),
                    Text('Island silver coins will appear here when the treasury mints them.',
                        style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
                  ],
                ),
              ),
            )
          else
            ..._coins.map((coin) => _CoinCard(
                  coin: coin,
                  onBuy: () => _buyCoin(coin['id'] as String),
                  theme: theme,
                )),

          const SizedBox(height: 24),

          // ── What is Nanotaler? ──
          Card(
            color: Colors.blue.shade50,
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.info_outline, color: Colors.blue.shade700),
                      const SizedBox(width: 8),
                      Text('What is a Nanotaler?',
                          style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold)),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Nanotaler (NT) is the base unit of the Island economy.\n\n'
                    '• 1 TLR (Taler) = $ntPerTlr NT\n'
                    '• 1 TLR = 1 troy oz of silver\n'
                    '• Each coin is 1 TLR, fully silver-backed\n'
                    '• Premium: 5% over spot for minting & logistics\n\n'
                    'Holders can redeem physical coins at the island treasury.',
                    style: theme.textTheme.bodySmall,
                  ),
                ],
              ),
            ),
          ),

          const SizedBox(height: 16),

          // ── My Coins ──
          if (_myCoins.isNotEmpty) ...[
            const SizedBox(height: 16),
            Row(
              children: [
                Icon(Icons.account_balance_wallet, color: Colors.green.shade700),
                const SizedBox(width: 8),
                Text('My Coins', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
                const Spacer(),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                  decoration: BoxDecoration(
                    color: Colors.green.shade100,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text('${_myCoins.length} owned',
                      style: TextStyle(fontSize: 12, color: Colors.green.shade800)),
                ),
              ],
            ),
            const SizedBox(height: 8),
            ..._myCoins.map((coin) => _MyCoinCard(
                  coin: coin,
                  onRedeem: () => _redeemCoin(coin['id'] as String),
                  theme: theme,
                )),
          ],
        ],
      ),
    );
  }
}

class _MyCoinCard extends StatelessWidget {
  final Map<String, dynamic> coin;
  final VoidCallback onRedeem;
  final ThemeData theme;

  const _MyCoinCard({required this.coin, required this.onRedeem, required this.theme});

  @override
  Widget build(BuildContext context) {
    final priceNt = coin['price_nt'] as int? ?? 0;
    final serial = coin['serial'] as String? ?? 'N/A';
    final boughtAt = coin['bought_at'] as String? ?? '';
    final f = DateFormat('yyyy-MM-dd HH:mm');
    final date = boughtAt.isNotEmpty ? f.tryParse(boughtAt.substring(0, 19)) : null;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      color: Colors.green.shade50,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 48, height: 48,
                  decoration: BoxDecoration(
                    color: Colors.green.shade200,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Icon(Icons.workspace_premium, color: Colors.white, size: 28),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(serial,
                          style: theme.textTheme.titleSmall?.copyWith(
                            fontWeight: FontWeight.bold,
                            fontFamily: 'monospace',
                          )),
                      if (date != null)
                        Text('Bought: ${f.format(date)}',
                            style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
                    ],
                  ),
                ),
              ],
            ),
            const Divider(),
            Row(
              children: [
                Expanded(child: NtDisplay(amountNt: priceNt)),
                FilledButton.tonalIcon(
                  onPressed: onRedeem,
                  icon: const Icon(Icons.replay, size: 18),
                  label: const Text('Redeem'),
                  style: FilledButton.styleFrom(backgroundColor: Colors.amber.shade100),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _CoinCard extends StatelessWidget {
  final Map<String, dynamic> coin;
  final VoidCallback onBuy;
  final ThemeData theme;

  const _CoinCard({required this.coin, required this.onBuy, required this.theme});

  @override
  Widget build(BuildContext context) {
    final priceNt = coin['price_nt'] as int? ?? 0;
    final spot = (coin['silver_spot_usd'] as num?)?.toDouble() ?? 0;
    final serial = coin['serial'] as String? ?? 'N/A';
    final status = coin['status'] as String? ?? 'unknown';
    final createdAt = coin['created_at'] as String? ?? '';

    final f = DateFormat('yyyy-MM-dd HH:mm');
    final date = createdAt.isNotEmpty ? f.tryParse(createdAt.substring(0, 19)) : null;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 48,
                  height: 48,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade200,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(Icons.monetization_on, color: Colors.amber.shade700, size: 28),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(serial,
                          style: theme.textTheme.titleSmall?.copyWith(
                            fontWeight: FontWeight.bold,
                            fontFamily: 'monospace',
                          )),
                      const SizedBox(height: 2),
                      Text('1 TLR Physical Silver Coin',
                          style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                  decoration: BoxDecoration(
                    color: status == 'available' ? Colors.green.shade100 : Colors.grey.shade100,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(status,
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                        color: status == 'available' ? Colors.green.shade800 : Colors.grey,
                      )),
                ),
              ],
            ),
            const Divider(),
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Price', style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
                      NtDisplay(amountNt: priceNt),
                      const SizedBox(height: 2),
                      Text('\$${spot.toStringAsFixed(2)} spot',
                          style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
                    ],
                  ),
                ),
                if (status == 'available')
                  FilledButton.tonalIcon(
                    onPressed: onBuy,
                    icon: const Icon(Icons.shopping_cart, size: 18),
                    label: const Text('Buy'),
                  ),
              ],
            ),
            if (date != null)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text('Minted: ${f.format(date)}',
                    style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
              ),
          ],
        ),
      ),
    );
  }
}
