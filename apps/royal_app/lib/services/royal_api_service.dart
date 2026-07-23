import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;

/// RoyalApiService manages typed HTTP client for the Royal API endpoints.
class RoyalApiService {
  final String baseUrl;
  final http.Client _client = http.Client();
  StreamSubscription? _sseSub;

  RoyalApiService(this.baseUrl);

  Map<String, String> get _headers => {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  };

  Future<Map<String, dynamic>> _get(String path) async {
    final r = await _client.get(Uri.parse('$baseUrl$path'), headers: _headers);
    if (r.statusCode != 200) throw ApiException(r.statusCode, r.body);
    return jsonDecode(r.body) as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> _post(String path, {Map<String, dynamic>? body}) async {
    final r = await _client.post(
      Uri.parse('$baseUrl$path'),
      headers: _headers,
      body: body != null ? jsonEncode(body) : null,
    );
    if (r.statusCode != 200) throw ApiException(r.statusCode, r.body);
    return jsonDecode(r.body) as Map<String, dynamic>;
  }

  void dispose() {
    _sseSub?.cancel();
    _client.close();
  }

  // ── Health & Version ──
  Future<Map<String, dynamic>> getVersion() => _get('/api/version');
  Future<Map<String, dynamic>> getHealth() => _get('/api/health');
  Future<Map<String, dynamic>> getHealthChecks() => _get('/api/health/checks');

  // ── AI Steward ──
  Future<Map<String, dynamic>> aiChat(String question, {String? context, String? profile, String? userId}) =>
      _post('/api/ai/chat', body: {'question': question, if (context != null) 'context': context, if (profile != null) 'profile': profile, if (userId != null) 'user_id': userId});

  Future<Map<String, dynamic>> aiExplainSilver() => _get('/api/ai/explain-silver');
  Future<Map<String, dynamic>> aiMemoryStats() => _get('/api/ai/memory/stats');

  // ── Royal Treasury ──
  Future<Map<String, dynamic>> getReserve() => _get('/api/royal/reserve');
  Future<Map<String, dynamic>> setReserve(int ng) =>
      _post('/api/royal/reserve', body: {'action': 'set', 'ng': ng});
  Future<Map<String, dynamic>> addReserve(int ng) =>
      _post('/api/royal/reserve', body: {'action': 'add', 'ng': ng});
  Future<Map<String, dynamic>> getSilverOracle() => _get('/api/royal/oracle');
  Future<Map<String, dynamic>> updateOracle(double price) =>
      _post('/api/royal/oracle', body: {'price': price});
  Future<Map<String, dynamic>> getDeflation() => _get('/api/royal/deflation');
  Future<Map<String, dynamic>> triggerDeflation() =>
      _post('/api/royal/deflation', body: {'action': 'trigger'});
  Future<Map<String, dynamic>> getAutoMintConfig() => _get('/api/royal/auto-mint');
  Future<Map<String, dynamic>> updateAutoMint(bool enabled, int intervalMs) =>
      _post('/api/royal/auto-mint', body: {'enabled': enabled, 'interval_ms': intervalMs});
  Future<Map<String, dynamic>> triggerDividend() =>
      _post('/api/royal/dividend', body: {'action': 'trigger'});
  Future<Map<String, dynamic>> getDividendHistory() => _get('/api/royal/dividend-history');
  Future<Map<String, dynamic>> mint(int amount) =>
      _post('/api/royal/mint', body: {'amount': amount});
  Future<Map<String, dynamic>> burn(int amount) =>
      _post('/api/royal/burn', body: {'amount': amount});
  Future<Map<String, dynamic>> getBanknotes() => _get('/api/royal/banknotes');
  Future<Map<String, dynamic>> createBanknote(String owner, int amount) =>
      _post('/api/royal/banknotes', body: {'owner': owner, 'amount': amount});
  Future<Map<String, dynamic>> getProofOfReserve() => _get('/api/royal/proof-of-reserve');
  Future<Map<String, dynamic>> getRates() => _get('/api/royal/rates');
  Future<Map<String, dynamic>> getTokenomics() => _get('/api/royal/tokenomics');
  Future<Map<String, dynamic>> getForecast() => _get('/api/royal/forecast');
  Future<Map<String, dynamic>> getAuditLog() => _get('/api/royal/audit-log');

  // ── Royal Dashboard & Alerts ──
  Future<Map<String, dynamic>> getDashboardUI() => _get('/api/royal/ui');
  Future<Map<String, dynamic>> getAlertRules() => _get('/api/royal/alerts/list');
  Future<Map<String, dynamic>> addAlertRule(Map<String, dynamic> rule) =>
      _post('/api/royal/alerts/add', body: rule);
  Future<Map<String, dynamic>> deleteAlertRule(String id) =>
      _post('/api/royal/alerts/delete', body: {'id': id});

  // ── Multi-Sig ──
  Future<Map<String, dynamic>> getMultiSig() => _get('/api/royal/multisig');
  Future<Map<String, dynamic>> addSigner(String pubkey) =>
      _post('/api/royal/multisig', body: {'action': 'add', 'pubkey': pubkey});
  Future<Map<String, dynamic>> removeSigner(String pubkey) =>
      _post('/api/royal/multisig', body: {'action': 'remove', 'pubkey': pubkey});

  // ── Cron ──
  Future<Map<String, dynamic>> getCronRules() => _get('/api/royal/cron/list');
  Future<Map<String, dynamic>> addCronRule(Map<String, dynamic> rule) =>
      _post('/api/royal/cron/add', body: rule);
  Future<Map<String, dynamic>> deleteCronRule(String id) =>
      _post('/api/royal/cron/delete', body: {'id': id});

  // ── Sync ──
  Future<Map<String, dynamic>> getSync() => _get('/api/royal/sync');

  // ── Scheduled Actions ──
  Future<Map<String, dynamic>> getScheduledActions() => _get('/api/royal/schedule/list');
  Future<Map<String, dynamic>> createScheduledAction(Map<String, dynamic> action) =>
      _post('/api/royal/schedule/create', body: action);

  // ── Audit Export ──
  Future<Map<String, dynamic>> exportAudit({String format = 'json'}) =>
      _get('/api/royal/audit/export?format=$format');

  // ── Node Groups ──
  Future<Map<String, dynamic>> getNodeGroups() => _get('/api/royal/nodes/groups');
  Future<Map<String, dynamic>> createNodeGroup(String name, List<String> nodes) =>
      _post('/api/royal/nodes/groups', body: {'action': 'create', 'name': name, 'nodes': nodes});
  Future<Map<String, dynamic>> deleteNodeGroup(String name) =>
      _post('/api/royal/nodes/groups', body: {'action': 'delete', 'name': name});

  // ── Emergency Stop ──
  Future<Map<String, dynamic>> getEmergencyStop() => _get('/api/royal/emergency-stop');
  Future<Map<String, dynamic>> setEmergencyStop(bool enable) =>
      _post('/api/royal/emergency-stop', body: {'action': enable ? 'enable' : 'disable'});

  // ── Node Reputation ──
  Future<Map<String, dynamic>> getReputations() => _get('/api/royal/nodes/reputation');
  Future<Map<String, dynamic>> sendHeartbeat(String pubkey, {double latencyMs = 0}) =>
      _get('/api/royal/nodes/heartbeat?pubkey=$pubkey&latency_ms=$latencyMs');

  // ── Analytics ──
  Future<Map<String, dynamic>> getTreasuryTrends({int days = 30}) =>
      _get('/api/royal/analytics/treasury-trends?days=$days');

  // ── Crypto Reserves ──
  Future<Map<String, dynamic>> getCryptoReserves() => _get('/api/royal/crypto-reserves');
  Future<Map<String, dynamic>> updateCryptoReserve(Map<String, dynamic> reserve) =>
      _post('/api/royal/crypto-reserves', body: reserve);

  // ── Rate Limit Stats ──
  Future<Map<String, dynamic>> getRateLimitStats() => _get('/api/royal/rate-limit-stats');

  // ── Ping ──
  Future<Map<String, dynamic>> ping() => _get('/api/royal/test/ping');

  // ── DC Cloud ──
  Future<Map<String, dynamic>> dcStatus() => _get('/api/dc/status');
  Future<Map<String, dynamic>> dcList() => _get('/api/dc/list');
  Future<Map<String, dynamic>> dcSwarm() => _get('/api/dc/swarm');
  Future<Map<String, dynamic>> dcSeed(String infohash) =>
      _post('/api/dc/seed', body: {'infohash': infohash});

  // ── Chat Bridge ──
  Future<Map<String, dynamic>> chatStatus() => _get('/api/chat/status');
  Future<Map<String, dynamic>> chatBroadcast(String message) =>
      _post('/api/royal/chat/broadcast', body: {'message': message});
  Future<Map<String, dynamic>> chatTreasuryAlert(String message) =>
      _post('/api/royal/chat/treasury-alert', body: {'message': message});

  // ── Governance ──
  Future<Map<String, dynamic>> getConstitution() => _get('/api/royal/governance');
  Future<Map<String, dynamic>> createProposal(Map<String, dynamic> proposal) =>
      _post('/api/royal/governance/proposals', body: proposal);
  Future<Map<String, dynamic>> getProposals() => _get('/api/royal/governance/proposals');

  // ── Economy ──
  Future<Map<String, dynamic>> getEconomyReport() => _get('/api/royal/economy/report');
  Future<Map<String, dynamic>> getRates2() => _get('/api/economy/rates');
  Future<Map<String, dynamic>> getTreasuryForecast() => _get('/api/economy/treasury-forecast');

  // ── System ──
  Future<Map<String, dynamic>> getInfo() => _get('/api/admin/info');
  Future<Map<String, dynamic>> getSystemMetrics() => _get('/api/admin/metrics/system');
  Future<Map<String, dynamic>> getDockerStatus() => _get('/api/admin/docker');
  Future<Map<String, dynamic>> getServiceStatus() => _get('/api/admin/service/status');
  Future<Map<String, dynamic>> getDiagnostics() => _get('/api/admin/diagnostics');

  // ── ParanoidX ──
  Future<Map<String, dynamic>> getParanoidXStatus() => _get('/api/paranoidx/status');
  Future<Map<String, dynamic>> getParanoidXHistory() => _get('/api/paranoidx/history');

  // ── SSE Events ──
  Stream<Map<String, dynamic>> sseEvents() {
    late StreamController<Map<String, dynamic>> controller;
    controller = StreamController<Map<String, dynamic>>();
    final request = http.Request('GET', Uri.parse('$baseUrl/api/royal/events'));
    request.headers.addAll(_headers);
    _client.send(request).then((response) {
      response.stream
          .transform(utf8.decoder)
          .transform(const LineSplitter())
          .listen((line) {
        if (line.startsWith('data: ')) {
          try {
            final data = jsonDecode(line.substring(6)) as Map<String, dynamic>;
            controller.add(data);
          } catch (_) {}
        }
      }, onError: (e) => controller.addError(e));
    });
    return controller.stream;
  }
}

/// ApiException manages exception representing a non-200 HTTP response from the API.
class ApiException implements Exception {
  final int statusCode;
  final String body;
  ApiException(this.statusCode, this.body);

/// Returns the current message value.
  String get message => 'API error $statusCode: $body';
  @override String toString() => message;
}
