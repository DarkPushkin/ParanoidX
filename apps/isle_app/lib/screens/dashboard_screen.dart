import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import '../widgets/metrics_panel.dart';

const _gold = Color(0xFFFFD700);

const _silver = Color(0xFFC0C0C0);

const _green = Color(0xFF3FB950);

const _red = Color(0xFFF85149);

const _blue = Color(0xFF58A6FF);

const _bg = Color(0xFF0D1117);

const _card = Color(0xFF161B22);

/// DashboardScreen manages the server dashboard with status, system, treasury and radio tabs.
class DashboardScreen extends StatefulWidget {
  final String serverUrl;
  final String pubkey;

  const DashboardScreen({super.key, this.serverUrl = 'http://127.0.0.1:8080', this.pubkey = ''});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> with SingleTickerProviderStateMixin {
  late TabController _tabs;
  Map<String, dynamic>? _version, _health, _sysMetrics;
  Map<String, dynamic>? _balance, _economy, _oracle;
  List<dynamic> _services = [];
  bool _connected = false;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 5, vsync: this);
    _connect();
  }

  @override
  void dispose() { _tabs.dispose(); super.dispose(); }

  Future<void> _connect() async {
    setState(() => _loading = true);
    final client = http.Client();
    try {
      final base = 'http://127.0.0.1:8080';
      final r = await Future.wait([
        client.get(Uri.parse('$base/api/version')).timeout(const Duration(seconds: 3)),
        client.get(Uri.parse('$base/api/health')).timeout(const Duration(seconds: 3)),
        client.get(Uri.parse('$base/api/admin/metrics/system')).timeout(const Duration(seconds: 3)),
        client.get(Uri.parse('$base/api/economy/oracle')).timeout(const Duration(seconds: 3)),
        client.get(Uri.parse('$base/api/admin/service/status')).timeout(const Duration(seconds: 3)),
      ], eagerError: false);
      if (mounted) setState(() {
        _version = _tryJson(r[0]);
        _health = _tryJson(r[1]);
        _sysMetrics = _tryJson(r[2]);
        _oracle = _tryJson(r[3]);
        final svc = _tryJson(r[4]);
        _services = (svc?['services'] as List?) ?? [];
        _connected = r[0].statusCode == 200;
        _loading = false;
      });
    } catch (_) {
      if (mounted) setState(() { _connected = false; _loading = false; });
    } finally { client.close(); }
  }

  Map<String, dynamic>? _tryJson(http.Response r) {
    try { return jsonDecode(r.body) as Map<String, dynamic>; } catch (_) { return null; }
  }

/// Returns the current  sf value.
  double get _sf => MediaQuery.of(context).textScaleFactor.clamp(0.8, 1.5);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Isle Dashboard', style: TextStyle(fontSize: 24 * _sf, fontWeight: FontWeight.w700)),
        actions: [
          if (_loading)
            const Padding(padding: EdgeInsets.all(16), child: SizedBox(width: 22, height: 22, child: CircularProgressIndicator(strokeWidth: 2)))
          else
            Container(
              margin: const EdgeInsets.only(right: 12),
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
              decoration: BoxDecoration(
                color: (_connected ? _green : _red).withAlpha(30),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Row(mainAxisSize: MainAxisSize.min, children: [
                Container(width: 10, height: 10, decoration: BoxDecoration(shape: BoxShape.circle, color: _connected ? _green : _red)),
                const SizedBox(width: 8),
                Text(_connected ? 'ONLINE' : 'OFFLINE', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w700, color: _connected ? _green : _red)),
              ]),
            ),
        ],
        bottom: TabBar(
          controller: _tabs,
          isScrollable: true,
          labelColor: _gold,
          unselectedLabelColor: const Color(0xFF8B949E),
          indicatorColor: _gold,
          labelStyle: TextStyle(fontSize: 16 * _sf, fontWeight: FontWeight.w600),
          tabs: const [
            Tab(icon: Icon(Icons.dashboard), text: 'Status'),
            Tab(icon: Icon(Icons.monitor_heart), text: 'System'),
            Tab(icon: Icon(Icons.wifi), text: 'Network'),
            Tab(icon: Icon(Icons.account_balance), text: 'Treasury'),
            Tab(icon: Icon(Icons.radio), text: 'Radio'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabs,
        children: [
          _buildStatusTab(),
          _buildSystemTab(),
          _buildNetworkTab(),
          _buildTreasuryTab(),
          _buildRadioTab(),
        ],
      ),
    );
  }

  // ── Status Tab ──
  Widget _buildStatusTab() {
    if (_loading) return const Center(child: CircularProgressIndicator());
    if (!_connected) return _offlineView();
    final v = _version ?? {};
    final h = _health ?? {};
    final uptime = h['uptime_hours'] ?? 0;
    return ListView(padding: EdgeInsets.all(20 * _sf), children: [
      _bigCard(children: [
        Row(children: [
          const Icon(Icons.check_circle, size: 32, color: _green),
          const SizedBox(width: 12),
          Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text('All Systems Nominal', style: TextStyle(fontSize: 22 * _sf, fontWeight: FontWeight.w700, color: _green)),
            Text('Build: ${v['build'] ?? '—'}', style: TextStyle(fontSize: 14 * _sf, color: const Color(0xFF8B949E))),
          ]),
          const Spacer(),
          Column(children: [
            Text('${uptime}h', style: TextStyle(fontSize: 28 * _sf, fontWeight: FontWeight.w700, color: _gold)),
            Text('uptime', style: TextStyle(fontSize: 13 * _sf, color: const Color(0xFF8B949E))),
          ]),
        ]),
      ]),
      const SizedBox(height: 16),
      _nodeInfo(v),
      const SizedBox(height: 16),
      MetricsPanel(apiBase: 'http://127.0.0.1:8080'),
    ]);
  }

  Widget _offlineView() {
    return Center(
      child: Padding(
        padding: EdgeInsets.all(40 * _sf),
        child: Column(mainAxisSize: MainAxisSize.min, children: [
          Icon(Icons.cloud_off, size: 80 * _sf, color: _red.withAlpha(120)),
          SizedBox(height: 20 * _sf),
          Text('Server Unreachable', style: TextStyle(fontSize: 28 * _sf, fontWeight: FontWeight.w700, color: _red)),
          SizedBox(height: 12 * _sf),
          Text('Could not connect to http://127.0.0.1:8080', style: TextStyle(fontSize: 16 * _sf, color: const Color(0xFF8B949E)), textAlign: TextAlign.center),
          SizedBox(height: 24 * _sf),
          FilledButton.icon(
            icon: const Icon(Icons.refresh, size: 22),
            label: Text('Retry', style: TextStyle(fontSize: 18 * _sf)),
            onPressed: _connect,
            style: FilledButton.styleFrom(backgroundColor: _blue, padding: EdgeInsets.symmetric(horizontal: 32, vertical: 16)),
          ),
        ]),
      ),
    );
  }

  // ── System Tab ──
  Widget _buildSystemTab() {
    final sm = _sysMetrics ?? {};
    final cpu = (sm['cpu_percent'] as num?)?.toDouble() ?? 0;
    final disk = sm['disk'] as Map<String, dynamic>? ?? {};
    final ram = sm['ram'] as Map<String, dynamic>? ?? {};
    final ramPct = double.tryParse((ram['used_pct'] as String? ?? '0').replaceAll('%', '')) ?? 0;
    final diskPct = double.tryParse((disk['used_pct'] as String? ?? '0').replaceAll('%', '')) ?? 0;
    final svc = _services;

    return ListView(padding: EdgeInsets.all(20 * _sf), children: [
      Row(children: [
        Expanded(child: _gauge('CPU', cpu, _blue, Icons.memory)),
        SizedBox(width: 12 * _sf),
        Expanded(child: _gauge('RAM', ramPct, _green, Icons.storage)),
        SizedBox(width: 12 * _sf),
        Expanded(child: _gauge('DISK', diskPct, diskPct > 90 ? _red : _gold, Icons.disc_full)),
      ]),
      SizedBox(height: 16 * _sf),
      _bigCard(title: 'Services', children: [
        if (svc.isEmpty)
          Text('No service data', style: TextStyle(fontSize: 16 * _sf, color: const Color(0xFF8B949E)))
        else
          ...svc.take(8).map((s) {
            final m = s as Map<String, dynamic>;
            final ok = m['status'] == 'running' || m['healthy'] == true;
            return Padding(
              padding: EdgeInsets.symmetric(vertical: 4 * _sf),
              child: Row(children: [
                Container(width: 12, height: 12, decoration: BoxDecoration(shape: BoxShape.circle, color: ok ? _green : _red)),
                SizedBox(width: 10 * _sf),
                Text('${m['name'] ?? m['service'] ?? '—'}', style: TextStyle(fontSize: 16 * _sf, color: Colors.white)),
                const Spacer(),
                Text('${m['status'] ?? '—'}', style: TextStyle(fontSize: 14 * _sf, color: ok ? _green : _red)),
              ]),
            );
          }),
      ]),
    ]);
  }

  // ── Network Tab ──
  Widget _buildNetworkTab() {
    final v = _version ?? {};
    return ListView(padding: EdgeInsets.all(20 * _sf), children: [
      _bigCard(title: 'Connection', children: [
        _row('Node', '127.0.0.1:8080'),
        _row('Pubkey', widget.pubkey.isNotEmpty ? '${widget.pubkey.substring(0, 16)}...' : '—'),
        _row('Build', v['build'] ?? '—'),
        _row('Go', v['go'] ?? '—'),
        _row('Started', v['started'] ?? '—'),
      ]),
    ]);
  }

  // ── Treasury Tab ──
  Widget _buildTreasuryTab() {
    final o = _oracle ?? {};
    final price = o['price'] ?? o['usd_per_oz'] ?? 0;
    final coin = o['change_pct'] ?? 0;
    return ListView(padding: EdgeInsets.all(20 * _sf), children: [
      _bigCard(title: 'Silver Oracle', children: [
        Center(
          child: Column(children: [
            Text('\$$price', style: TextStyle(fontSize: 48 * _sf, fontWeight: FontWeight.w700, color: _silver)),
            Text('per troy oz', style: TextStyle(fontSize: 16 * _sf, color: const Color(0xFF8B949E))),
            if (coin is num)
              Text('${coin >= 0 ? '+' : ''}${coin.toStringAsFixed(2)}%',
                  style: TextStyle(fontSize: 18 * _sf, color: coin >= 0 ? _green : _red)),
          ]),
        ),
      ]),
    ]);
  }

  // ── Radio Tab ──
  Widget _buildRadioTab() {
    return ListView(padding: EdgeInsets.all(20 * _sf), children: [
      _bigCard(title: 'Isle Radio', children: [
        Center(child: Column(children: [
          Icon(Icons.radio, size: 48 * _sf, color: _gold),
          SizedBox(height: 12 * _sf),
          Text('Isle Radio Station', style: TextStyle(fontSize: 22 * _sf, fontWeight: FontWeight.w600, color: Colors.white)),
          SizedBox(height: 4 * _sf),
          Text('Saint Mary Liberty Island', style: TextStyle(fontSize: 14 * _sf, color: const Color(0xFF8B949E))),
        ])),
      ]),
    ]);
  }

  // ── Shared Widgets ──
  Widget _bigCard({String? title, required List<Widget> children}) {
    return Container(
      padding: EdgeInsets.all(20 * _sf),
      decoration: BoxDecoration(
        color: _card,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFF30363D)),
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        if (title != null) ...[
          Text(title, style: TextStyle(fontSize: 20 * _sf, fontWeight: FontWeight.w700, color: _gold)),
          SizedBox(height: 12 * _sf),
        ],
        ...children,
      ]),
    );
  }

  Widget _nodeInfo(Map<String, dynamic> v) {
    return _bigCard(children: [
      _row('Go', v['go'] ?? '—'),
      _row('Build', v['build'] ?? '—'),
      _row('Uptime', '${_health?['uptime_hours'] ?? 0}h'),
      _row('Server', '127.0.0.1:8080'),
    ]);
  }

  Widget _row(String label, String value) {
    return Padding(
      padding: EdgeInsets.symmetric(vertical: 4 * _sf),
      child: Row(children: [
        Text('$label: ', style: TextStyle(fontSize: 16 * _sf, color: const Color(0xFF8B949E))),
        Expanded(child: Text(value, style: TextStyle(fontSize: 16 * _sf, color: Colors.white, fontWeight: FontWeight.w500))),
      ]),
    );
  }

  Widget _gauge(String label, num value, Color color, IconData icon) {
    final pct = (value as num).toDouble().clamp(0, 100);
    return Container(
      padding: EdgeInsets.all(16 * _sf),
      decoration: BoxDecoration(
        color: _card,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: color.withAlpha(60)),
      ),
      child: Column(children: [
        Icon(icon, size: 28 * _sf, color: color),
        SizedBox(height: 8 * _sf),
        Text('${pct.toStringAsFixed(0)}%', style: TextStyle(fontSize: 28 * _sf, fontWeight: FontWeight.w700, color: color)),
        SizedBox(height: 4 * _sf),
        Text(label, style: TextStyle(fontSize: 14 * _sf, color: const Color(0xFF8B949E), fontWeight: FontWeight.w600)),
        SizedBox(height: 8 * _sf),
        ClipRRect(
          borderRadius: BorderRadius.circular(6),
          child: LinearProgressIndicator(
            value: pct / 100,
            backgroundColor: const Color(0xFF21262D),
            valueColor: AlwaysStoppedAnimation(color),
            minHeight: 8,
          ),
        ),
      ]),
    );
  }
}
