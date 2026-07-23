import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:models/models.dart';
import 'package:widgets/widgets.dart';
import '../services/isle_api_service.dart';
import '../widgets/nt_display.dart';

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
  List<TokenBalance> _tokenBalances = [];
  List<ExternalWallet> _externalWallets = [];
  final _toCtrl = TextEditingController();
  final _amountCtrl = TextEditingController();
  bool _tokensExpanded = false;
  bool _extWalletsExpanded = false;
  final _ewTypeCtrl = TextEditingController();
  final _ewAddressCtrl = TextEditingController();
  final _ewLabelCtrl = TextEditingController();
  final _ewChainCtrl = TextEditingController();
  final _ctSymbolCtrl = TextEditingController();
  final _ctNameCtrl = TextEditingController();
  final _ctDecimalsCtrl = TextEditingController(text: '18');
  final _ctChainCtrl = TextEditingController(text: 'custom');
  final _ctContractCtrl = TextEditingController();

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
    final d = await _get('/api/wallet/create');
    return d['pubkey'] as String;
  }

  Future<void> _load() async {
    try {
      final b = await widget.api.getBalance();
      final h = await widget.api.getHoldings();
      List<Map<String, dynamic>> coins = [];
      List<TokenBalance> tokens = [];
      List<ExternalWallet> extWallets = [];
      try {
        final pk = await _ensurePubkey();
        final mc = await _get('/api/silver/my-coins?pubkey=$pk');
        coins = (mc['coins'] as List<dynamic>?)?.cast<Map<String, dynamic>>() ?? [];
      } catch (_) {}
      try {
        final tb = await widget.api.token.balances(pubkey: await _ensurePubkey());
        tokens = (tb['balances'] as List?)
                ?.map((e) => TokenBalance.fromJson(e as Map<String, dynamic>))
                .toList() ??
            [];
      } catch (_) {}
      try {
        final ew = await widget.api.externalWallet.list(pubkey: await _ensurePubkey());
        extWallets = (ew['wallets'] as List?)
                ?.map((e) => ExternalWallet.fromJson(e as Map<String, dynamic>))
                .toList() ??
            [];
      } catch (_) {}
      setState(() {
        _balance = b;
        _holdings = h;
        _myCoins = coins;
        _tokenBalances = tokens;
        _externalWallets = extWallets;
      });
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
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Transfer sent')));
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e')));
    }
  }

  Future<void> _linkExternalWallet() async {
    final type = _ewTypeCtrl.text.trim();
    final address = _ewAddressCtrl.text.trim();
    if (type.isEmpty || address.isEmpty) return;
    try {
      await widget.api.externalWallet.link(
        pubkey: await _ensurePubkey(),
        walletType: type,
        walletAddress: address,
        label: _ewLabelCtrl.text.trim(),
        chain: _ewChainCtrl.text.trim(),
      );
      _ewTypeCtrl.clear(); _ewAddressCtrl.clear(); _ewLabelCtrl.clear(); _ewChainCtrl.clear();
      _load();
      if (mounted) { Navigator.of(context).pop(); ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Wallet linked'))); }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e')));
    }
  }

  Future<void> _unlinkWallet(ExternalWallet w) async {
    try {
      await widget.api.externalWallet.unlink(await _ensurePubkey(), w.walletType);
      _load();
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('${w.displayName} unlinked')));
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e')));
    }
  }

  Future<void> _addCustomToken() async {
    final symbol = _ctSymbolCtrl.text.trim().toUpperCase();
    final name = _ctNameCtrl.text.trim();
    if (symbol.isEmpty || name.isEmpty) return;
    try {
      await widget.api.token.addCustom(
        symbol: symbol, name: name,
        decimals: int.tryParse(_ctDecimalsCtrl.text.trim()) ?? 18,
        chain: _ctChainCtrl.text.trim(),
        contractAddress: _ctContractCtrl.text.trim(),
      );
      _ctSymbolCtrl.clear(); _ctNameCtrl.clear(); _ctChainCtrl.text = 'custom'; _ctDecimalsCtrl.text = '18'; _ctContractCtrl.clear();
      _load();
      if (mounted) { Navigator.of(context).pop(); ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Token added'))); }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e')));
    }
  }

  void _showLinkWalletDialog() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Link External Wallet'),
        content: SingleChildScrollView(child: Column(mainAxisSize: MainAxisSize.min, children: [
          TextField(controller: _ewTypeCtrl, decoration: const InputDecoration(labelText: 'Wallet type', hintText: 'trust_wallet, telegram_wallet', border: OutlineInputBorder())),
          const SizedBox(height: 8),
          TextField(controller: _ewAddressCtrl, decoration: const InputDecoration(labelText: 'Wallet address / public key', border: OutlineInputBorder())),
          const SizedBox(height: 8),
          TextField(controller: _ewLabelCtrl, decoration: const InputDecoration(labelText: 'Label (optional)', border: OutlineInputBorder())),
          const SizedBox(height: 8),
          TextField(controller: _ewChainCtrl, decoration: const InputDecoration(labelText: 'Chain (optional)', hintText: 'all', border: OutlineInputBorder())),
        ])),
        actions: [
          TextButton(onPressed: () => Navigator.of(ctx).pop(), child: const Text('Cancel')),
          FilledButton(onPressed: _linkExternalWallet, child: const Text('Link')),
        ],
      ),
    );
  }

  void _showAddTokenDialog() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Add Custom Token'),
        content: SingleChildScrollView(child: Column(mainAxisSize: MainAxisSize.min, children: [
          TextField(controller: _ctSymbolCtrl, decoration: const InputDecoration(labelText: 'Symbol', hintText: 'SHIB', border: OutlineInputBorder())),
          const SizedBox(height: 8),
          TextField(controller: _ctNameCtrl, decoration: const InputDecoration(labelText: 'Name', hintText: 'Shiba Inu', border: OutlineInputBorder())),
          const SizedBox(height: 8),
          TextField(controller: _ctDecimalsCtrl, decoration: const InputDecoration(labelText: 'Decimals', border: OutlineInputBorder()), keyboardType: TextInputType.number),
          const SizedBox(height: 8),
          TextField(controller: _ctChainCtrl, decoration: const InputDecoration(labelText: 'Chain', border: OutlineInputBorder())),
          const SizedBox(height: 8),
          TextField(controller: _ctContractCtrl, decoration: const InputDecoration(labelText: 'Contract address (optional)', border: OutlineInputBorder())),
        ])),
        actions: [
          TextButton(onPressed: () => Navigator.of(ctx).pop(), child: const Text('Cancel')),
          FilledButton(onPressed: _addCustomToken, child: const Text('Add Token')),
        ],
      ),
    );
  }

  Future<void> _confirmUnlink(ExternalWallet w) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Unlink Wallet'),
        content: Text('Remove ${w.displayName} (${w.walletAddress.substring(0, 8)}...)?'),
        actions: [
          TextButton(onPressed: () => Navigator.of(ctx).pop(false), child: const Text('Cancel')),
          FilledButton(onPressed: () => Navigator.of(ctx).pop(true), child: const Text('Unlink')),
        ],
      ),
    );
    if (ok == true) _unlinkWallet(w);
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final body = RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(12),
        children: [
          _balanceCard(theme), const SizedBox(height: 10),
          _transferCard(theme), const SizedBox(height: 10),
          _holdingsCard(theme), const SizedBox(height: 10),
          _silverCoinsCard(theme), const SizedBox(height: 10),
          _tokenBalancesCard(theme), const SizedBox(height: 10),
          _externalWalletsCard(theme),
        ],
      ),
    );
    return Scaffold(appBar: AppBar(title: const Text('Wallet')), body: body);
  }

  Widget _balanceCard(ThemeData theme) => Card(child: Padding(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
    Row(children: [Icon(Icons.account_balance_wallet, color: theme.colorScheme.primary), const SizedBox(width: 8), Text('Balance', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold))]),
    const SizedBox(height: 12),
    _balance == null ? const Text('Loading...') : Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      NtDisplay(amountNt: _balance!['liquid_balance_ng'] ?? 0),
      const SizedBox(height: 4),
      Text('Frozen: ${_balance!['banknotes_count'] ?? 0} banknotes', style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
    ]),
  ])));

  Widget _transferCard(ThemeData theme) => Card(child: Padding(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
    Row(children: [Icon(Icons.send, color: theme.colorScheme.primary), const SizedBox(width: 8), Text('Transfer', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold))]),
    const SizedBox(height: 12),
    TextField(controller: _toCtrl, decoration: const InputDecoration(labelText: 'Recipient pubkey', border: OutlineInputBorder())),
    const SizedBox(height: 8),
    TextField(controller: _amountCtrl, decoration: const InputDecoration(labelText: 'Amount (NT)', border: OutlineInputBorder()), keyboardType: TextInputType.number),
    const SizedBox(height: 12),
    FilledButton.icon(onPressed: _transfer, icon: const Icon(Icons.send, size: 16), label: const Text('Send')),
  ])));

  Widget _holdingsCard(ThemeData theme) => Card(child: Padding(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
    Row(children: [Icon(Icons.inventory_2, color: theme.colorScheme.primary), const SizedBox(width: 8), Text('Holdings', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold))]),
    const SizedBox(height: 12),
    _holdings == null ? const Text('Loading...') : _holdings!.containsKey('liquid_ng')
        ? Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            NtDisplay(amountNt: (_holdings!['liquid_ng'] as num?)?.toInt() ?? 0),
            Text('Frozen: ${(_holdings!['frozen_ng'] as num?)?.toInt() ?? 0} nt', style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey)),
            const Divider(),
            ...(_holdings!['banknotes'] as List? ?? []).map((bn) => Padding(padding: const EdgeInsets.symmetric(vertical: 2), child: Text('${bn['serial']} — ${bn['rarity']} (${bn['denomination_ng']} nt)', style: const TextStyle(fontFamily: 'monospace', fontSize: 12)))),
          ])
        : const Text('No holdings data'),
  ])));

  Widget _silverCoinsCard(ThemeData theme) => Card(child: Padding(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
    Row(children: [
      Icon(Icons.monetization_on, color: Colors.amber.shade700), const SizedBox(width: 8),
      Text('Silver Coins', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)), const Spacer(),
      Container(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2), decoration: BoxDecoration(color: _myCoins.isNotEmpty ? Colors.amber.shade100 : Colors.grey.shade100, borderRadius: BorderRadius.circular(12)),
        child: Text('${_myCoins.length}', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: _myCoins.isNotEmpty ? Colors.brown.shade700 : Colors.grey))),
    ]),
    const SizedBox(height: 12),
    if (_myCoins.isEmpty) Text('No silver coins yet.', style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey))
    else ..._myCoins.map((c) => Padding(padding: const EdgeInsets.symmetric(vertical: 4), child: Row(children: [
      Icon(Icons.circle, size: 10, color: Colors.amber.shade600), const SizedBox(width: 8),
      Text('${c['serial']}', style: const TextStyle(fontFamily: 'monospace', fontSize: 12, fontWeight: FontWeight.bold)), const Spacer(),
      Text('${c['price_nt']} nt', style: TextStyle(fontSize: 11, color: Colors.grey.shade600)),
    ]))),
  ])));

  Widget _tokenBalancesCard(ThemeData theme) => Card(child: Padding(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
    InkWell(onTap: () => setState(() => _tokensExpanded = !_tokensExpanded), child: Row(children: [
      Icon(Icons.token, color: Colors.indigo), const SizedBox(width: 8),
      Text('Token Balances', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)), const Spacer(),
      if (_tokenBalances.isNotEmpty) Container(margin: const EdgeInsets.only(right: 8), padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2), decoration: BoxDecoration(color: Colors.indigo.shade50, borderRadius: BorderRadius.circular(12)),
        child: Text('${_tokenBalances.length}', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Colors.indigo.shade700))),
      Icon(_tokensExpanded ? Icons.expand_less : Icons.expand_more),
    ])),
    if (_tokensExpanded) ...[
      const SizedBox(height: 12),
      if (_tokenBalances.isEmpty) Text('No token balances yet.', style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey))
      else ..._tokenBalances.map((tb) => Padding(padding: const EdgeInsets.symmetric(vertical: 4), child: Row(children: [
        Container(width: 28, height: 28, decoration: BoxDecoration(color: _tokenColor(tb.symbol).withValues(alpha: 0.15), borderRadius: BorderRadius.circular(14)),
          child: Center(child: Text(tb.symbol.substring(0, 1), style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: _tokenColor(tb.symbol))))),
        const SizedBox(width: 10),
        Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(tb.symbol, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
          Text(tb.name, style: TextStyle(fontSize: 11, color: Colors.grey.shade600)),
        ])),
        Text(tb.formattedBalance, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
      ]))),
      const SizedBox(height: 8),
      OutlinedButton.icon(onPressed: _showAddTokenDialog, icon: const Icon(Icons.add, size: 16), label: const Text('Add Custom Token'), style: OutlinedButton.styleFrom(foregroundColor: Colors.indigo)),
    ],
  ])));

  Widget _externalWalletsCard(ThemeData theme) => Card(child: Padding(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
    InkWell(onTap: () => setState(() => _extWalletsExpanded = !_extWalletsExpanded), child: Row(children: [
      Icon(Icons.wallet, color: Colors.teal), const SizedBox(width: 8),
      Text('External Wallets', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)), const Spacer(),
      if (_externalWallets.isNotEmpty) Container(margin: const EdgeInsets.only(right: 8), padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2), decoration: BoxDecoration(color: Colors.teal.shade50, borderRadius: BorderRadius.circular(12)),
        child: Text('${_externalWallets.length}', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Colors.teal.shade700))),
      Icon(_extWalletsExpanded ? Icons.expand_less : Icons.expand_more),
    ])),
    if (_extWalletsExpanded) ...[
      const SizedBox(height: 12),
      if (_externalWallets.isEmpty) Text('No external wallets linked.', style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey))
      else ..._externalWallets.map((w) => Padding(padding: const EdgeInsets.symmetric(vertical: 4), child: Card(elevation: 0, color: theme.colorScheme.surfaceContainerHighest, child: ListTile(dense: true,
        leading: _walletIcon(w.walletType, theme),
        title: Text(w.displayName, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
        subtitle: Text('${w.walletAddress.substring(0, 8)}...${w.walletAddress.length > 8 ? w.walletAddress.substring(w.walletAddress.length - 4) : ''}', style: const TextStyle(fontFamily: 'monospace', fontSize: 11)),
        trailing: Row(mainAxisSize: MainAxisSize.min, children: [
          w.isVerified ? Icon(Icons.verified, size: 16, color: Colors.teal.shade400) : Icon(Icons.verified_outlined, size: 16, color: Colors.grey),
          const SizedBox(width: 4),
          IconButton(icon: const Icon(Icons.delete_outline, size: 16), onPressed: () => _confirmUnlink(w), padding: EdgeInsets.zero, constraints: const BoxConstraints()),
        ]),
      )))),
      const SizedBox(height: 8),
      OutlinedButton.icon(onPressed: _showLinkWalletDialog, icon: const Icon(Icons.link, size: 16), label: const Text('Link Wallet'), style: OutlinedButton.styleFrom(foregroundColor: Colors.teal)),
    ],
  ])));

  Widget _walletIcon(String type, ThemeData theme) {
    final icon = switch (type.toLowerCase()) {
      'trust_wallet' || 'trustwallet' => Icons.verified,
      'telegram_wallet' || 'telegram' => Icons.telegram,
      'metamask' => Icons.hexagon, 'phantom' => Icons.bolt,
      _ => Icons.account_balance_wallet,
    };
    return Container(width: 32, height: 32, decoration: BoxDecoration(color: Colors.teal.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(8)), child: Icon(icon, size: 18, color: Colors.teal));
  }

  Color _tokenColor(String symbol) => switch (symbol.toUpperCase()) {
    'BTC' => Colors.orange, 'ETH' => const Color(0xFF627EEA), 'USDT' => Colors.green,
    'SOL' => const Color(0xFF9945FF), 'XRP' => const Color(0xFF00AAE4), 'ADA' => const Color(0xFF0033AD),
    'AVAX' => const Color(0xFFE84142), 'DOT' => Colors.magenta, 'LINK' => const Color(0xFF375BD2),
    'NG' => const Color(0xFF1A1A2E), _ => Colors.indigo,
  };

  @override
  void dispose() {
    _toCtrl.dispose(); _amountCtrl.dispose();
    _ewTypeCtrl.dispose(); _ewAddressCtrl.dispose(); _ewLabelCtrl.dispose(); _ewChainCtrl.dispose();
    _ctSymbolCtrl.dispose(); _ctNameCtrl.dispose(); _ctDecimalsCtrl.dispose(); _ctChainCtrl.dispose(); _ctContractCtrl.dispose();
    super.dispose();
  }
}
