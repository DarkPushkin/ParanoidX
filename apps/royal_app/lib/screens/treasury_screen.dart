import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../theme.dart';
import '../main.dart';

/// TreasuryScreen manages treasuryscreen functionality.
class TreasuryScreen extends StatefulWidget {
  const TreasuryScreen({super.key});
  @override
  State<TreasuryScreen> createState() => _TreasuryScreenState();
}

class _TreasuryScreenState extends State<TreasuryScreen> {
  Map<String, dynamic>? _reserve, _oracle, _tokenomics, _forecast, _audit;
  Timer? _timer;

  final _mintCtrl = TextEditingController();
  final _burnCtrl = TextEditingController();
  final _priceCtrl = TextEditingController();

  @override
  void initState() {
    super.initState();
    _refresh();
    _timer = Timer.periodic(const Duration(seconds: 30), (_) => _refresh());
  }

  @override
  void dispose() { _timer?.cancel(); _mintCtrl.dispose(); _burnCtrl.dispose(); _priceCtrl.dispose(); super.dispose(); }

  Future<void> _refresh() async {
    final api = context.read<AppState>().api;
    try {
      final r = await Future.wait([
        api.getReserve(), api.getSilverOracle(), api.getTokenomics(),
        api.getForecast(), api.getAuditLog(),
      ], eagerError: false);
      if (mounted) setState(() {
        _reserve = r[0]; _oracle = r[1]; _tokenomics = r[2];
        _forecast = r[3]; _audit = r[4];
      });
    } catch (_) {}
  }

  Future<void> _action(String label, Future<Map<String, dynamic>> Function() fn) async {
    try {
      final res = await fn();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(
          content: Text('${res['ok'] == true ? '✓' : '✗'} $label'),
          backgroundColor: res['ok'] == true ? RoyalTheme.green.withAlpha(40) : RoyalTheme.red.withAlpha(40),
        ));
        _refresh();
      }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e'), backgroundColor: RoyalTheme.red.withAlpha(40)));
    }
  }

  void _showMintDialog() {
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: RoyalTheme.darkCard,
      title: const Text('Mint Tokens'),
      content: TextField(
        controller: _mintCtrl,
        keyboardType: TextInputType.number,
        decoration: const InputDecoration(labelText: 'Amount (ng)'),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')),
        ElevatedButton(onPressed: () {
          final amt = int.tryParse(_mintCtrl.text);
          if (amt != null && amt > 0) {
            _action('Mint $amt ng', () => context.read<AppState>().api.mint(amt));
            Navigator.pop(ctx);
          }
        }, child: const Text('Mint')),
      ],
    ));
  }

  void _showBurnDialog() {
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: RoyalTheme.darkCard,
      title: const Text('Burn Tokens'),
      content: TextField(
        controller: _burnCtrl,
        keyboardType: TextInputType.number,
        decoration: const InputDecoration(labelText: 'Amount (ng)'),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')),
        ElevatedButton(
          style: ElevatedButton.styleFrom(backgroundColor: RoyalTheme.red),
          onPressed: () {
            final amt = int.tryParse(_burnCtrl.text);
            if (amt != null && amt > 0) {
              _action('Burn $amt ng', () => context.read<AppState>().api.burn(amt));
              Navigator.pop(ctx);
            }
          }, child: const Text('Burn')),
      ],
    ));
  }

  void _showOracleDialog() {
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: RoyalTheme.darkCard,
      title: const Text('Set Oracle Price'),
      content: TextField(
        controller: _priceCtrl,
        keyboardType: TextInputType.number,
        decoration: const InputDecoration(labelText: 'Price (USD/oz)'),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')),
        ElevatedButton(onPressed: () {
          final p = double.tryParse(_priceCtrl.text);
          if (p != null && p > 0) {
            _action('Update Oracle', () => context.read<AppState>().api.updateOracle(p));
            Navigator.pop(ctx);
          }
        }, child: const Text('Update')),
      ],
    ));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Row(children: [
          Container(padding: const EdgeInsets.all(6), decoration: BoxDecoration(color: RoyalTheme.gold.withAlpha(30), borderRadius: BorderRadius.circular(8)),
            child: const Icon(Icons.account_balance, color: RoyalTheme.gold, size: 20)),
          const SizedBox(width: 12),
          const Text('Treasury'),
        ]),
        actions: [
          IconButton(icon: const Icon(Icons.refresh), onPressed: _refresh),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _buildOverviewRow(),
          const SizedBox(height: 16),
          _buildActionGrid(),
          const SizedBox(height: 16),
          _buildTokenomicsCard(),
          const SizedBox(height: 16),
          _buildForecastCard(),
          const SizedBox(height: 16),
          _buildAuditCard(),
        ],
      ),
    );
  }

  Widget _buildOverviewRow() {
    final r = _reserve ?? {};
    final o = _oracle ?? {};
    final reserve = r['reserve_ng'] ?? r['ng'] ?? 0;
    final supply = r['supply_ng'] ?? r['total_supply'] ?? 0;
    final price = o['price'] ?? o['usd_per_oz'] ?? 0;
    final tier = r['tier'] ?? 0;
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: RoyalTheme.glowBorder(),
      child: Column(children: [
        Row(mainAxisAlignment: MainAxisAlignment.spaceAround, children: [
          _overviewItem('Reserve', '${reserve} ng', RoyalTheme.gold),
          _overviewItem('Supply', '${supply} ng', RoyalTheme.silver),
          _overviewItem('Price', '\$$price', RoyalTheme.green),
          _overviewItem('Tier', '$tier', Colors.blue),
        ]),
        const SizedBox(height: 12),
        LinearProgressIndicator(
          value: (tier is int && tier > 0) ? tier / 5 : 0,
          backgroundColor: const Color(0xFF21262D),
          valueColor: const AlwaysStoppedAnimation<Color>(RoyalTheme.gold),
        ),
        const SizedBox(height: 4),
        Text('Tier Progress', style: Theme.of(context).textTheme.bodyMedium),
      ]),
    );
  }

  Widget _overviewItem(String label, String value, Color? color) {
    final c = color ?? RoyalTheme.gold;
    return Column(children: [
      Text(value, style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: c)),
      Text(label, style: Theme.of(context).textTheme.bodyMedium),
    ]);
  }

  Widget _buildActionGrid() {
    return Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Text('Treasury Actions', style: Theme.of(context).textTheme.titleMedium),
      const SizedBox(height: 12),
      Wrap(spacing: 10, runSpacing: 10, children: [
        _actionButton('Mint', Icons.add_circle, RoyalTheme.green, _showMintDialog),
        _actionButton('Burn', Icons.remove_circle, RoyalTheme.red, _showBurnDialog),
        _actionButton('Oracle', Icons.trending_up, RoyalTheme.accent, _showOracleDialog),
        _actionButton('Deflation', Icons.compress, RoyalTheme.orange, () => _action('Deflation', () => context.read<AppState>().api.triggerDeflation())),
        _actionButton('Dividend', Icons.payments, RoyalTheme.gold, () => _action('Dividend', () => context.read<AppState>().api.triggerDividend())),
        _actionButton('Auto-Mint', Icons.settings, RoyalTheme.purple, () {}),
      ]),
    ]);
  }

  Widget _actionButton(String label, IconData icon, Color color, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: (MediaQuery.of(context).size.width - 52) / 3,
        padding: const EdgeInsets.symmetric(vertical: 16),
        decoration: BoxDecoration(
          color: color.withAlpha(15),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: color.withAlpha(50)),
        ),
        child: Column(children: [
          Icon(icon, color: color, size: 24),
          const SizedBox(height: 6),
          Text(label, style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: color)),
        ]),
      ),
    );
  }

  Widget _buildTokenomicsCard() {
    final t = _tokenomics ?? {};
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: RoyalTheme.gradientCard(),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text('Tokenomics', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 12),
        _kvRow('Max Supply', '${t['max_supply'] ?? '—'}'),
        _kvRow('Circulating', '${t['circulating'] ?? t['total_supply'] ?? '—'}'),
        _kvRow('Burned', '${t['burned'] ?? 0}'),
        _kvRow('Mining Rate', '${t['mining_rate'] ?? '—'}'),
      ]),
    );
  }

  Widget _buildForecastCard() {
    final f = _forecast ?? {};
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: RoyalTheme.gradientCard(colors: [RoyalTheme.darkCard, const Color(0xFF111820)]),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Icon(Icons.trending_up, color: RoyalTheme.teal, size: 18),
          const SizedBox(width: 8),
          Text('Forecast', style: Theme.of(context).textTheme.titleMedium),
        ]),
        const SizedBox(height: 12),
        _kvRow('30d Reserve\n', '${f['reserve_30d'] ?? '—'}'),
        _kvRow('Health Score', '${f['health_score'] ?? f['score'] ?? '—'}'),
        _kvRow('Recommendation', '${f['recommendation'] ?? '—'}'),
      ]),
    );
  }

  Widget _buildAuditCard() {
    final entries = (_audit?['entries'] as List?) ?? [];
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: RoyalTheme.gradientCard(),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Icon(Icons.history, color: RoyalTheme.silver, size: 18),
          const SizedBox(width: 8),
          Text('Recent Audit', style: Theme.of(context).textTheme.titleMedium),
        ]),
        const SizedBox(height: 12),
        if (entries.isEmpty)
          Text('No audit entries yet', style: Theme.of(context).textTheme.bodyMedium)
        else
          ...(entries.take(5).map((e) => _auditRow(e as Map<String, dynamic>))),
      ]),
    );
  }

  Widget _auditRow(Map<String, dynamic> e) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(children: [
        Container(width: 6, height: 6, decoration: const BoxDecoration(shape: BoxShape.circle, color: RoyalTheme.gold)),
        const SizedBox(width: 8),
        Text('${e['action'] ?? '—'}', style: const TextStyle(fontSize: 13, color: Colors.white)),
        const Spacer(),
        Text('${e['timestamp'] ?? ''}'.substring(0, 10), style: Theme.of(context).textTheme.bodyMedium),
      ]),
    );
  }

  Widget _kvRow(String k, String v) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3),
      child: Row(children: [
        Text('$k: ', style: const TextStyle(color: Color(0xFF8B949E), fontSize: 13)),
        Text(v, style: const TextStyle(color: Colors.white, fontSize: 13, fontWeight: FontWeight.w500)),
      ]),
    );
  }
}
