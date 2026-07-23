import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:file_picker/file_picker.dart';
import 'services/isle_api_service.dart';
import 'services/server_addresses.dart';
import 'services/tor_aware_client.dart' show createHttpClient;
import 'services/tor_manager.dart';
import 'services/relay_manager.dart';
import 'services/simplex_loader.dart';
import 'widgets/relay_selector_dialog.dart';
import 'screens/dashboard_screen.dart';
import 'screens/wallet_screen.dart';
import 'screens/vault_screen.dart';
import 'screens/market_screen.dart';
import 'screens/pos_screen.dart';
import 'screens/radio_screen.dart';
import 'screens/royal_screen.dart';
import 'screens/simplex_chat_screen.dart';
import 'screens/welcome_screen.dart';
import 'services/radio_player.dart';
import 'widgets/isle_emblem.dart';
import 'widgets/top_bar.dart';
import 'services/vpn_manager.dart';

/// Entry point for The Island application.
///
/// Parses optional CLI args `--server` and `--pubkey`, then launches [TheIsleApp].
void main(List<String> args) {
  String? serverUrl;
  String? pubkey;
  for (int i = 0; i < args.length; i++) {
    if (args[i] == '--server' && i + 1 < args.length) {
      serverUrl = args[i + 1];
    } else if (args[i] == '--pubkey' && i + 1 < args.length) {
      pubkey = args[i + 1];
    }
  }
  runApp(TheIsleApp(initialServerUrl: serverUrl, initialPubkey: pubkey));
}

/// Root MaterialApp widget for The Island.
///
/// Configures light/dark themes using Material 3 with a green seed color
/// and delegates to [Bootstrapper] for the initial screen.
class TheIsleApp extends StatelessWidget {
  /// Optional server URL passed via CLI arguments.
  final String? initialServerUrl;
  /// Optional public key passed via CLI arguments.
  final String? initialPubkey;
  const TheIsleApp({super.key, this.initialServerUrl, this.initialPubkey});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'The Island',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF2E7D32),
          brightness: Brightness.light,
        ),
        useMaterial3: true,
      ),
      darkTheme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF2E7D32),
          brightness: Brightness.dark,
        ),
        useMaterial3: true,
      ),
      themeMode: ThemeMode.system,
      home: Bootstrapper(initialServerUrl: initialServerUrl, initialPubkey: initialPubkey),
    );
  }
}

/// Bootstrapper manages initial application bootstrap and dependency setup.
class Bootstrapper extends StatefulWidget {
  final String? initialServerUrl;
  final String? initialPubkey;
  const Bootstrapper({super.key, this.initialServerUrl, this.initialPubkey});
  @override
  State<Bootstrapper> createState() => _BootstrapperState();
}

class _BootstrapperState extends State<Bootstrapper> with WidgetsBindingObserver {
  bool _loading = true;
  bool _hasKeys = false;
  bool _locked = true;
  bool _needsConnectionCheck = true;
  String _locale = 'en';
  String _serverUrl = 'http://q273p7coau3uvzeddexvdgv6andorfzvplstztheso2qcsj4yqvfzzad.onion:80';
  String _pubkey = '';
  String _privkey = '';
  late ServerAddressManager _addrMgr;
  late RelayManager _relayMgr;
  late TorManager _torMgr;
  late http.Client _httpClient;
  late RadioPlayer _radioPlayer;
  late VpnManager _vpnMgr;
  String _buildVersion = '';
  SimpleXTransportConfig _simplexConfig = SimpleXTransportConfig.empty();
  static const _focusChannel = MethodChannel('com.theisle/focus');

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _addrMgr = ServerAddressManager();
    _relayMgr = RelayManager();
    _torMgr = TorManager();
    _httpClient = createHttpClient(useProxy: true);
    _radioPlayer = RadioPlayer(httpClient: _httpClient, apiBase: _serverUrl);
    _vpnMgr = VpnManager(httpClient: _httpClient)..start();
    _focusChannel.setMethodCallHandler((call) async {
      if (call.method == 'onFocusChanged' && call.arguments is bool) {
        if (!(call.arguments as bool) && !_locked && mounted) _lockApp();
      }
    });
    _loadKeys();
  }

  @override
  void dispose() {
    _radioPlayer.dispose();
    _vpnMgr.dispose();
    _httpClient.close();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.inactive || state == AppLifecycleState.paused) {
      if (!_locked && mounted) _lockApp();
    }
  }

  Future<void> _loadKeys() async {
    await _addrMgr.load();
    await _relayMgr.load();
    _updateSimplexFromRelay();
    final prefs = await SharedPreferences.getInstance();
    var savedUrl = widget.initialServerUrl ?? prefs.getString('server_url') ?? _addrMgr.firstUrl;
    if (savedUrl.contains('127.0.0.1') || savedUrl.contains('localhost')) {
      for (final a in _addrMgr.addresses) {
        if (a.url.contains('.onion')) {
          savedUrl = a.url;
          break;
        }
      }
    }
    final savedPubkey = widget.initialPubkey ?? prefs.getString('pubkey') ?? '';
    final savedPrivkey = prefs.getString('privkey') ?? '';
    final savedLocale = prefs.getString('locale') ?? detectSystemLocale();

    final hasKeys = savedPubkey.isNotEmpty && savedPrivkey.isNotEmpty;
    setState(() {
      _serverUrl = savedUrl;
      _pubkey = savedPubkey;
      _privkey = savedPrivkey;
      _hasKeys = hasKeys;
      _locale = savedLocale;
      _loading = false;
      _locked = true;
      _needsConnectionCheck = !hasKeys;
    });
    _fetchVersion();
  }

  void _onLocaleChanged(String loc) async {
    setState(() => _locale = loc);
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('locale', loc);
  }

  void _onUnlocked() {
    setState(() {
      _locked = false;
      _needsConnectionCheck = false;
    });
    _updateSimplexFromRelay();
  }

  void _lockApp() {
    if (mounted) setState(() { _locked = true; _needsConnectionCheck = false; });
  }

  void _fullLockApp() async {
    await _torMgr.stopTor();
    if (mounted) setState(() { _locked = true; _needsConnectionCheck = true; });
  }

  Future<void> _switchAccount() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('pubkey');
    await prefs.remove('privkey');
    await prefs.remove('server_url');
    await prefs.remove('pin_setup_done');
    await _torMgr.stopTor();
    if (mounted) setState(() {
      _hasKeys = false;
      _pubkey = '';
      _privkey = '';
      _locked = false;
      _needsConnectionCheck = true;
    });
  }

  Future<void> _fetchVersion() async {
    try {
      final url = _serverUrl.contains('onion') ? 'http://127.0.0.1:8080/api/version' : '$_serverUrl/api/version';
      final client = http.Client();
      final resp = await client
          .get(Uri.parse(url))
          .timeout(const Duration(seconds: 3));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        setState(() => _buildVersion = data['version'] as String? ?? '');
      }
      client.close();
    } catch (_) {}
  }

  void _updateSimplexFromRelay() {
    final relay = _relayMgr.activeRelay;
    if (relay != null) {
      setState(() {
        _simplexConfig = SimpleXTransportConfig(
          smp: relay.smp,
          xftp: relay.xftp,
          ice: relay.ice,
          onion: _serverUrl,
          contact: '',
          label: relay.label,
        );
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    Widget child;
    if (_loading) {
      child = const Scaffold(body: Center(child: CircularProgressIndicator()));
    } else if (_locked) {
      child = WelcomeScreen(
        initialLocale: _locale,
        needsConnectionCheck: _needsConnectionCheck,
        torMgr: _torMgr,
        addrMgr: _addrMgr,
        relayMgr: _relayMgr,
        serverUrl: _serverUrl,
        onLocaleChanged: _onLocaleChanged,
        onUnlocked: _onUnlocked,
        radioPlayer: _radioPlayer,
        vpnMgr: _vpnMgr,
        buildVersion: _buildVersion,
      );
    } else if (!_hasKeys) {
      child = OnboardingFlow(
        serverUrl: _serverUrl,
        addrMgr: _addrMgr,
        torMgr: _torMgr,
        simplexConfig: _simplexConfig,
        radioPlayer: _radioPlayer,
        vpnMgr: _vpnMgr,
        buildVersion: _buildVersion,
        onComplete: (pubkey, privkey, url) {
          setState(() {
            _pubkey = pubkey;
            _privkey = privkey;
            _serverUrl = url;
            _radioPlayer.updateApiBase(_serverUrl);
            _radioPlayer.autoPlay();
          });
        },
      );
    } else {
      child = AppShell(
        serverUrl: _serverUrl,
        pubkey: _pubkey,
        privkey: _privkey,
        torMgr: _torMgr,
        relayMgr: _relayMgr,
        simplexConfig: _simplexConfig,
        onLock: _lockApp,
        onFullLock: _fullLockApp,
        onSwitchAccount: _switchAccount,
        radioPlayer: _radioPlayer,
        vpnMgr: _vpnMgr,
      );
    }
    return child;
  }
}

/// OnboardingFlow manages new user account creation and restoration flow.
class OnboardingFlow extends StatefulWidget {
  final String serverUrl;
  final ServerAddressManager addrMgr;
  final TorManager torMgr;
  final SimpleXTransportConfig simplexConfig;
  final void Function(String pubkey, String privkey, String url) onComplete;
  final RadioPlayer radioPlayer;
  final VpnManager? vpnMgr;
  final String buildVersion;
  const OnboardingFlow({super.key, required this.serverUrl, required this.addrMgr, required this.torMgr, required this.simplexConfig, required this.onComplete, required this.radioPlayer, this.vpnMgr, this.buildVersion = ''});
  @override
  State<OnboardingFlow> createState() => _OnboardingFlowState();
}

class _OnboardingFlowState extends State<OnboardingFlow> {
  late String _serverUrl;
  String _pubkey = '';
  String _privkey = '';
  String _mnemonic = '';
  late http.Client _httpClient;
  bool _transportReady = false;
  bool _transportChecking = true;
  String _transportError = '';

  @override
  void initState() {
    super.initState();
    _serverUrl = widget.serverUrl;
    _httpClient = createHttpClient(useProxy: true);
    _doTransportHandshake();
  }

  @override
  void dispose() {
    _httpClient.close();
    super.dispose();
  }

  Future<void> _doTransportHandshake() async {
    setState(() {
      _transportChecking = true;
      _transportReady = false;
      _transportError = '';
    });
    try {
      await widget.torMgr.ensureRunning();
      if (!mounted) return;

      const handshake = {
        'type': 'ping',
        'payload': {},
        'id': 'handshake-1',
      };
      final url = Uri.tryParse(_serverUrl)?.resolve('/api/transport/send') ?? Uri.parse('/api/transport/send');
      final resp = await _httpClient
          .post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(handshake))
          .timeout(const Duration(seconds: 10));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['type'] == 'pong') {
          if (mounted) setState(() { _transportReady = true; _transportChecking = false; });
          return;
        }
      }
      if (mounted) setState(() { _transportReady = false; _transportChecking = false; _transportError = 'Handshake failed'; });
    } catch (e) {
      if (mounted) setState(() { _transportReady = false; _transportChecking = false; _transportError = '$e'; });
    }
  }

  Future<Map<String, dynamic>> _apiPost(String endpoint, Map<String, dynamic> body) async {
    final envelope = {
      'type': 'api.request',
      'payload': {'method': 'POST', 'path': endpoint, 'body': body},
      'id': 'req-${DateTime.now().millisecondsSinceEpoch}',
    };
    final url = Uri.tryParse(_serverUrl)?.resolve('/api/transport/send') ?? Uri.parse('/api/transport/send');
    final resp = await _httpClient.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode(envelope));
    final env = jsonDecode(resp.body) as Map<String, dynamic>;
    if (env['type'] == 'error') throw Exception(env['payload'].toString());
    final apiResp = env['payload'] as Map<String, dynamic>;
    final status = apiResp['status'] as int;
    final bodyRaw = apiResp['body'];
    final result = bodyRaw is Map ? bodyRaw as Map<String, dynamic> : jsonDecode(bodyRaw as String) as Map<String, dynamic>;
    if (status >= 400) throw Exception(result['error']?.toString() ?? 'API error $status');
    return result;
  }

  void _onCreateAccount() async {
    if (!_transportReady) return;
    try {
      final data = await _apiPost('/api/account/create', {});
      setState(() {
        _pubkey = data['pubkey'] as String;
        _privkey = data['privkey'] as String;
        _mnemonic = data['mnemonic'] as String;
      });
      if (!mounted) return;
      Navigator.of(context).push(MaterialPageRoute(
        builder: (_) => _MnemonicDisplayScreen(
          mnemonic: _mnemonic,
          pubkey: _pubkey,
          onVerified: () => _saveAndContinue(),
          onRetry: () { Navigator.of(context).pop(); _onCreateAccount(); },
        ),
      ));
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red));
      }
    }
  }

  void _onRestoreAccount() {
    if (!_transportReady) return;
    Navigator.of(context).push(MaterialPageRoute(
      builder: (_) => _RestoreScreen(
        serverUrl: _serverUrl,
        httpClient: _httpClient,
        onRestored: (pubkey, privkey) { Navigator.of(context).pop(); _saveAndContinue(pubkey: pubkey, privkey: privkey); },
      ),
    ));
  }

  void _saveAndContinue({String? pubkey, String? privkey}) async {
    final pk = pubkey ?? _pubkey;
    final sk = privkey ?? _privkey;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('pubkey', pk);
    await prefs.setString('privkey', sk);
    await prefs.setString('server_url', _serverUrl);
    if (_mnemonic.isNotEmpty) await prefs.setString('mnemonic', _mnemonic);
    widget.onComplete(pk, sk, _serverUrl);
  }

  void _openAddressManager() {
    showDialog(
      context: context,
      builder: (ctx) => _AddressManagerDialog(addrMgr: widget.addrMgr, httpClient: _httpClient),
    ).then((_) {
      if (widget.addrMgr.addresses.isNotEmpty) {
        final newUrl = widget.addrMgr.firstUrl;
        if (newUrl != _serverUrl) {
          setState(() => _serverUrl = newUrl);
          _doTransportHandshake();
        }
      }
    });
  }

  Widget _torStatusIcon(TorStatus status) {
    switch (status) {
      case TorStatus.running:
        return const Icon(Icons.check_circle, color: Colors.green, size: 16);
      case TorStatus.stopped:
        return const Icon(Icons.cancel, color: Colors.red, size: 16);
      case TorStatus.error:
        return const Icon(Icons.error, color: Colors.orange, size: 16);
      case TorStatus.unknown:
        return const Icon(Icons.help, color: Colors.grey, size: 16);
    }
  }

  Future<void> _testAddresses() async {
    final mgr = widget.addrMgr;
    await mgr.testAll(client: _httpClient);
    if (!mounted) return;

    final reachable = mgr.addresses.where((a) => a.isReachable).toList();
    if (reachable.isNotEmpty) {
      setState(() => _serverUrl = reachable.first.url);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Доступны: ${reachable.map((a) => a.label.isEmpty ? a.url : a.label).join(', ')}'),
          backgroundColor: Colors.green,
        ),
      );
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Нет доступных адресов'), backgroundColor: Colors.red),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final torOk = widget.torMgr.isRunning;
    return Scaffold(
      body: Column(
        children: [
          TopBar(
            player: widget.radioPlayer,
            torMgr: widget.torMgr,
            vpnMgr: widget.vpnMgr,
            buildVersion: widget.buildVersion,
          ),
          Expanded(
            child: Stack(
              children: [
                Positioned.fill(
                  child: Opacity(
                    opacity: 0.04,
                    child: Image.asset('assets/images/emblem.png', fit: BoxFit.contain, alignment: Alignment.center),
                  ),
                ),
                Center(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 12),
                    child: ConstrainedBox(
                      constraints: const BoxConstraints(maxWidth: 400),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          // Header
                          Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              GestureDetector(
                                onTap: () => showIsleDeclaration(context),
                                child: ClipOval(
                                  child: Image.asset('assets/images/emblem.png', width: 56, height: 56, fit: BoxFit.cover),
                                ),
                              ),
                              const SizedBox(width: 12),
                              Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text('The Isle', style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                                  const SizedBox(height: 2),
                                  Text('Saint Mary Liberty Island', style: TextStyle(fontSize: 11, color: Colors.grey[400])),
                                ],
                              ),
                            ],
                          ),
                          const SizedBox(height: 16),

                          // Status card
                          Card(
                            margin: EdgeInsets.zero,
                            child: Padding(
                              padding: const EdgeInsets.all(12),
                              child: Column(
                                children: [
                                  _statusRow(Icons.network_check, 'Tor', widget.torMgr.statusLabel, torOk,
                                      () async {
                                        if (torOk) await widget.torMgr.stopTor();
                                        else await widget.torMgr.startTor();
                                        if (mounted) setState(() {});
                                      }, torOk ? 'Disconnect' : 'Connect'),
                                  const Divider(height: 8),
                                  _statusRow(Icons.swap_horiz, 'Transport',
                                      _transportChecking ? 'Connecting...'
                                          : _transportReady ? 'Connected'
                                          : _transportError.isNotEmpty ? _transportError : 'Off',
                                      _transportReady, _doTransportHandshake, 'Retry'),
                                  const Divider(height: 8),
                                  _statusRow(Icons.person, 'Account',
                                      _transportReady ? 'Ready to create' : 'Waiting...', _transportReady,
                                      null, null),
                                ],
                              ),
                            ),
                          ),
                          const SizedBox(height: 14),

                          // Account buttons
                          SizedBox(
                            width: 280,
                            child: Row(
                              children: [
                                Expanded(
                                  child: FilledButton(
                                    style: FilledButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 12)),
                                    onPressed: _transportReady ? _onCreateAccount : null,
                                    child: const Text('Create Account', style: TextStyle(fontSize: 13)),
                                  ),
                                ),
                                const SizedBox(width: 10),
                                Expanded(
                                  child: OutlinedButton(
                                    style: OutlinedButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 12)),
                                    onPressed: _transportReady ? _onRestoreAccount : null,
                                    child: const Text('Restore', style: TextStyle(fontSize: 13)),
                                  ),
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(height: 14),

                          // Server
                          SizedBox(
                            width: 280,
                            child: TextField(
                              decoration: const InputDecoration(
                                labelText: 'Server Address',
                                hintText: 'http://onion:80',
                                border: OutlineInputBorder(),
                                isDense: true,
                                prefixIcon: Icon(Icons.link, size: 16),
                              ),
                              controller: TextEditingController(text: _serverUrl),
                              style: const TextStyle(fontSize: 12, fontFamily: 'monospace'),
                              onChanged: (v) { _serverUrl = v.trim(); _doTransportHandshake(); },
                            ),
                          ),
                          const SizedBox(height: 6),
                          Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              TextButton.icon(
                                icon: const Icon(Icons.cloud_sync, size: 14),
                                label: const Text('Addresses', style: TextStyle(fontSize: 11)),
                                onPressed: _openAddressManager,
                              ),
                              const SizedBox(width: 8),
                              TextButton.icon(
                                icon: const Icon(Icons.network_check, size: 14),
                                label: const Text('Test', style: TextStyle(fontSize: 11)),
                                onPressed: _testAddresses,
                              ),
                            ],
                          ),

                          // SMP info
                          if (_transportReady && widget.simplexConfig.hasSMP) ...[
                            const SizedBox(height: 8),
                            Container(
                              padding: const EdgeInsets.all(8),
                              decoration: BoxDecoration(
                                color: Theme.of(context).colorScheme.surfaceContainerHighest,
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: SelectableText(
                                widget.simplexConfig.smp,
                                style: const TextStyle(fontSize: 9, fontFamily: 'monospace', color: Colors.grey),
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _statusRow(IconData icon, String title, String subtitle, bool? ok, VoidCallback? onAction, String? actionLabel) {
    return Row(
      children: [
        Container(
          width: 10, height: 10,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: ok == null ? Colors.grey : ok ? Colors.green : Colors.red,
          ),
        ),
        const SizedBox(width: 8),
        Icon(icon, size: 16, color: Colors.grey[500]),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(title, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500)),
              Text(subtitle, style: TextStyle(fontSize: 11, color: Colors.grey[500])),
            ],
          ),
        ),
        if (onAction != null && actionLabel != null)
          SizedBox(
            height: 28,
            child: TextButton(
              onPressed: onAction,
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(horizontal: 10),
                minimumSize: Size.zero,
                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
              ),
              child: Text(actionLabel, style: const TextStyle(fontSize: 11)),
            ),
          ),
      ],
    );
  }
}

class _MnemonicDisplayScreen extends StatefulWidget {
  final String mnemonic;
  final String pubkey;
  final VoidCallback onVerified;
  final VoidCallback onRetry;
  const _MnemonicDisplayScreen({required this.mnemonic, required this.pubkey, required this.onVerified, required this.onRetry});
  @override
  State<_MnemonicDisplayScreen> createState() => _MnemonicDisplayScreenState();
}

class _MnemonicDisplayScreenState extends State<_MnemonicDisplayScreen> {
  bool _showVerify = false;
  final _verifyCtrl = TextEditingController();

  @override
  void dispose() {
    _verifyCtrl.dispose();
    super.dispose();
  }

  void _onContinue() {
    setState(() => _showVerify = true);
  }

  Future<void> _onVerify() async {
    final input = _verifyCtrl.text.trim().toLowerCase();
    if (input == widget.mnemonic) {
      widget.onVerified();
    } else {
      showDialog(
        context: context,
        builder: (ctx) => AlertDialog(
          title: const Text('Mnemonic Mismatch'),
          content: const Text('The phrase you entered does not match. Would you like to try again or generate a new account?'),
          actions: [
            TextButton(onPressed: () { Navigator.of(ctx).pop(); _verifyCtrl.clear(); }, child: const Text('Try Again')),
            FilledButton(onPressed: () { Navigator.of(ctx).pop(); widget.onRetry(); }, child: const Text('New Account')),
          ],
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!_showVerify) {
      return Scaffold(
        appBar: AppBar(title: const Text('Your Secret Phrase')),
        body: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Card(
                color: Theme.of(context).colorScheme.primaryContainer,
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      Icon(Icons.warning_amber, color: Theme.of(context).colorScheme.error),
                      const SizedBox(width: 12),
                      const Expanded(child: Text('Write down these 24 words in order. Never share them with anyone. This is the only way to restore your account.')),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
              Expanded(
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    border: Border.all(color: Theme.of(context).dividerColor),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: SelectableText(widget.mnemonic, style: const TextStyle(fontSize: 16, fontFamily: 'monospace', height: 1.6)),
                ),
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton.icon(
                      icon: const Icon(Icons.copy),
                      label: const Text('Copy'),
                      onPressed: () {
                        Clipboard.setData(ClipboardData(text: widget.mnemonic));
                        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Copied to clipboard')));
                      },
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: FilledButton.icon(
                      icon: const Icon(Icons.check),
                      label: const Text('I Saved It'),
                      onPressed: _onContinue,
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      );
    }
    return Scaffold(
      appBar: AppBar(title: const Text('Verify Your Phrase')),
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Type your 24-word secret phrase to confirm you saved it correctly.',
                style: Theme.of(context).textTheme.bodyLarge),
            const SizedBox(height: 16),
            TextField(
              controller: _verifyCtrl,
              maxLines: 6,
              decoration: const InputDecoration(
                hintText: 'Paste or type your 24 words here...',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 16),
            FilledButton.icon(
              icon: const Icon(Icons.verified),
              label: const Text('Verify'),
              onPressed: _onVerify,
            ),
          ],
        ),
      ),
    );
  }
}

class _RestoreScreen extends StatefulWidget {
  final String serverUrl;
  final http.Client httpClient;
  final void Function(String pubkey, String privkey) onRestored;
  const _RestoreScreen({required this.serverUrl, required this.httpClient, required this.onRestored});
  @override
  State<_RestoreScreen> createState() => _RestoreScreenState();
}

class _RestoreScreenState extends State<_RestoreScreen> {
  final _mnemonicCtrl = TextEditingController();
  bool _loading = false;

  @override
  void dispose() {
    _mnemonicCtrl.dispose();
    super.dispose();
  }

  Future<void> _onRestore() async {
    final mnemonic = _mnemonicCtrl.text.trim().toLowerCase();
    if (mnemonic.isEmpty) return;
    setState(() => _loading = true);
    try {
      final url = Uri.tryParse(widget.serverUrl)?.resolve('/api/account/restore') ?? Uri.parse('/api/account/restore');
      final resp = await widget.httpClient.post(url,
          headers: {'Content-Type': 'application/json'},
          body: jsonEncode({'mnemonic': mnemonic}));
      final data = jsonDecode(resp.body) as Map<String, dynamic>;
      if (data.containsKey('pubkey')) {
        widget.onRestored(data['pubkey'] as String, data['privkey'] as String);
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Error: ${data['error'] ?? 'unknown'}'), backgroundColor: Colors.red),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Connection error: $e'), backgroundColor: Colors.red));
      }
    }
    if (mounted) setState(() => _loading = false);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Restore Account')),
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Enter your 24-word secret phrase to restore your account.',
                style: Theme.of(context).textTheme.bodyLarge),
            const SizedBox(height: 16),
            TextField(
              controller: _mnemonicCtrl,
              maxLines: 6,
              decoration: const InputDecoration(
                hintText: 'Paste or type your 24 words here...',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 16),
            FilledButton.icon(
              icon: _loading
                  ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                  : const Icon(Icons.login),
              label: Text(_loading ? 'Restoring...' : 'Restore'),
              onPressed: _loading ? null : _onRestore,
            ),
          ],
        ),
      ),
    );
  }
}

class _AddressManagerDialog extends StatefulWidget {
  final ServerAddressManager addrMgr;
  final http.Client httpClient;
  const _AddressManagerDialog({required this.addrMgr, required this.httpClient});
  @override
  State<_AddressManagerDialog> createState() => _AddressManagerDialogState();
}

class _AddressManagerDialogState extends State<_AddressManagerDialog> {
  final _manualCtrl = TextEditingController();

  @override
  void dispose() {
    _manualCtrl.dispose();
    super.dispose();
  }

  void _addManual() {
    final url = _manualCtrl.text.trim();
    if (url.isEmpty) return;
    setState(() {
      widget.addrMgr.add(url);
      _manualCtrl.clear();
    });
    widget.addrMgr.save();
  }

  Future<void> _addFromFile() async {
    try {
      final result = await FilePicker.platform.pickFiles(
        type: FileType.any,
        allowMultiple: false,
      );
      if (result == null || result.files.isEmpty) return;
      final file = File(result.files.single.path!);
      final content = await file.readAsString();
      final lines = content.split('\n').map((l) => l.trim()).where((l) => l.isNotEmpty && !l.startsWith('#'));
      for (final line in lines) {
        final parts = line.split('|');
        widget.addrMgr.add(parts[0].trim(), label: parts.length > 1 ? parts[1].trim() : null);
      }
      await widget.addrMgr.save();
      setState(() {});
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Loaded ${lines.length} address(es) from file')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _testSingle(int index) async {
    final ok = await widget.addrMgr.testAddress(index, client: widget.httpClient);
    if (!mounted) return;
    setState(() {});
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(ok ? '${widget.addrMgr.addresses[index].url} — OK' : '${widget.addrMgr.addresses[index].url} — недоступен'),
        backgroundColor: ok ? Colors.green : Colors.red,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final addresses = widget.addrMgr.addresses;
    return AlertDialog(
      title: const Text('Адреса серверов'),
      content: SizedBox(
        width: 400,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (addresses.isEmpty)
              const Padding(
                padding: EdgeInsets.all(16),
                child: Text('Нет сохранённых адресов'),
              )
            else
              Flexible(
                child: ListView.builder(
                  shrinkWrap: true,
                  itemCount: addresses.length,
                  itemBuilder: (_, i) {
                    final addr = addresses[i];
                    return Card(
                      child: ListTile(
                        dense: true,
                        title: Text(addr.label.isNotEmpty ? addr.label : addr.url, style: const TextStyle(fontSize: 13)),
                        subtitle: addr.label.isNotEmpty ? Text(addr.url, style: const TextStyle(fontSize: 11, fontFamily: 'monospace')) : null,
                        trailing: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(addr.isReachable ? Icons.check_circle : Icons.help_outline,
                                size: 18, color: addr.isReachable ? Colors.green : Colors.grey),
                            IconButton(
                              icon: const Icon(Icons.network_check, size: 18),
                              onPressed: () => _testSingle(i),
                              tooltip: 'Test',
                            ),
                            IconButton(
                              icon: const Icon(Icons.delete, size: 18, color: Colors.red),
                              onPressed: () {
                                setState(() => widget.addrMgr.removeAt(i));
                                widget.addrMgr.save();
                              },
                              tooltip: 'Remove',
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
              ),
            const Divider(),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _manualCtrl,
                    decoration: const InputDecoration(
                      hintText: 'http://адрес:порт',
                      border: OutlineInputBorder(),
                      isDense: true,
                      contentPadding: EdgeInsets.symmetric(horizontal: 8, vertical: 8),
                    ),
                    onSubmitted: (_) => _addManual(),
                  ),
                ),
                const SizedBox(width: 8),
                FilledButton.tonal(onPressed: _addManual, child: const Text('Add')),
              ],
            ),
            const SizedBox(height: 8),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton.icon(
                icon: const Icon(Icons.file_open, size: 18),
                label: const Text('Загрузить из файла'),
                onPressed: _addFromFile,
              ),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.of(context).pop(), child: const Text('Закрыть')),
      ],
    );
  }
}

/// AppShell manages the main application shell with navigation and settings.
class AppShell extends StatefulWidget {
  final String serverUrl;
  final String pubkey;
  final String privkey;
  final TorManager torMgr;
  final RelayManager relayMgr;
  final SimpleXTransportConfig simplexConfig;
  final VoidCallback? onLock;
  final VoidCallback? onFullLock;
  final VoidCallback? onSwitchAccount;
  final RadioPlayer radioPlayer;
  final VpnManager vpnMgr;
  const AppShell({super.key, required this.serverUrl, required this.pubkey, required this.privkey, required this.torMgr, required this.relayMgr, required this.simplexConfig, this.onLock, this.onFullLock, this.onSwitchAccount, required this.radioPlayer, required this.vpnMgr});
  @override
  State<AppShell> createState() => _AppShellState();
}

class _AppShellState extends State<AppShell> {
  int _selectedIndex = 0;
  late IsleApiService _api;
  Map<String, dynamic>? _balance;
  Map<String, dynamic>? _economy;
  late String _serverUrl;
  String _pubkey = '';
  bool _showSettings = false;
  late http.Client _httpClient;
  late TorManager _torMgr;
  late RelayManager _relayMgr;
  SimpleXTransportConfig _simplexConfig = SimpleXTransportConfig.empty();
  bool _torAutoConnect = true;
  Timer? _inactivityTimer;
  Timer? _healthTimer;
  bool _healthSmp = false;
  bool _healthXftp = false;
  bool _healthRadio = false;
  NodeHealth _nodeHealth = NodeHealth.up;
  int _healthFailCount = 0;
  String _royalPubkey = '';
  String _contactLink = '';
  String _mnemonic = '';
  bool _showMnemonic = false;
  TextEditingController? _urlCtrl;
  TextEditingController? _keyCtrl;

  @override
  void initState() {
    super.initState();
    _serverUrl = widget.serverUrl;
    _pubkey = widget.pubkey;
    _torMgr = widget.torMgr;
    _relayMgr = widget.relayMgr;
    _simplexConfig = widget.simplexConfig;
    _torMgr.autoConnect.then((v) => _torAutoConnect = v);
    _httpClient = createHttpClient(useProxy: true);
    _api = IsleApiService(_serverUrl, _pubkey, httpClient: _httpClient);
    _urlCtrl = TextEditingController(text: _serverUrl);
    _keyCtrl = TextEditingController(text: _pubkey);
    _loadMnemonic();
    _connect();
    _resetInactivityTimer();
    _startHealthCheck();
    _refreshConfigFromServer();
    Future.delayed(const Duration(seconds: 2), () => widget.radioPlayer.autoPlay());
  }

  @override
  void dispose() {
    _inactivityTimer?.cancel();
    _healthTimer?.cancel();
    _urlCtrl?.dispose();
    _keyCtrl?.dispose();
    super.dispose();
  }

  void _startHealthCheck() {
    _healthTimer = Timer.periodic(const Duration(seconds: 15), (_) => _checkHealth());
    _checkHealth();
  }

  Future<void> _checkHealth() async {
    try {
      final healthUrl = Uri.tryParse(_serverUrl)?.resolve('/api/transport/health') ?? Uri.parse('/api/transport/health');
      final resp = await _httpClient
          .get(healthUrl)
          .timeout(const Duration(seconds: 5));
      if (!mounted) return;
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        setState(() {
          _healthSmp = data['smp'] == true;
          _healthXftp = data['xftp'] == true;
          _healthRadio = data['radio'] == true;
          _healthFailCount = 0;
          _nodeHealth = (_healthSmp || _healthXftp || _healthRadio) ? NodeHealth.up : NodeHealth.reconnecting;
        });
        _fetchRoyalPubkey();
      } else {
        _handleHealthFail();
      }
    } catch (_) {
      _handleHealthFail();
    }
  }

  void _handleHealthFail() {
    if (!mounted) return;
    _healthFailCount++;
    setState(() {
      _healthSmp = _healthXftp = _healthRadio = false;
      _nodeHealth = _healthFailCount >= 2 ? NodeHealth.down : NodeHealth.reconnecting;
    });
    if (_healthFailCount >= 3) {
      _healthFailCount = 0;
      _connect();
    }
  }

  void _resetInactivityTimer() {
    _inactivityTimer?.cancel();
    _inactivityTimer = Timer(const Duration(seconds: 90), () {
      widget.onLock?.call();
    });
  }

  Future<void> _savePrefs() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('server_url', _serverUrl);
    await prefs.setString('pubkey', _pubkey);
    await prefs.setString('privkey', widget.privkey);
  }

  Future<void> _connect() async {
    if (_pubkey.isEmpty) return;
    _httpClient.close();
    _httpClient = createHttpClient(useProxy: true);
    _api = IsleApiService(_serverUrl, _pubkey, httpClient: _httpClient);
    await _savePrefs();
    try {
      final b = await _api.getBalance();
      final e = await _api.getEconomyState();
      setState(() { _balance = b; _economy = e; _showSettings = false; });
      _fetchRoyalPubkey();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Connection error: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  void _loadMnemonic() {
    SharedPreferences.getInstance().then((prefs) {
      final m = prefs.getString('mnemonic');
      if (m != null && mounted) setState(() => _mnemonic = m);
    });
  }

  Future<void> _fetchRoyalPubkey() async {
    try {
      final data = await _api.royal.nodes();
      if (mounted) {
        setState(() {
          _royalPubkey = data['royal_pubkey'] as String? ?? '';
          _contactLink = data['contact_link'] as String? ?? '';
        });
      }
    } catch (_) {}
  }

  void _setServerUrl(String url) {
    _serverUrl = url;
    _urlCtrl?.text = url;
    setState(() {});
  }
  void _setPubkey(String pk) {
    _pubkey = pk;
    _keyCtrl?.text = pk;
    setState(() {});
  }

  Future<void> _refreshConfigFromServer() async {
    try {
      final infoUrl = Uri.tryParse(_serverUrl)?.resolve('/api/transport/info') ?? Uri.parse('/api/transport/info');
      final resp = await _httpClient
          .get(infoUrl)
          .timeout(const Duration(seconds: 5));
      if (!mounted || resp.statusCode != 200) return;
      final data = jsonDecode(resp.body) as Map<String, dynamic>;
      final smp = (data['smp'] as String? ?? '').trim();
      final xftp = (data['xftp'] as String? ?? '').trim();
      final ice = (data['ice'] as String? ?? '').trim();
      final onion = (data['onion'] as String? ?? '').trim();
      if (smp.isNotEmpty || xftp.isNotEmpty) {
        _relayMgr.upsertFromServer(smp, xftp, ice, onion);
        _updateSimplexFromRelay();
      }
    } catch (_) {}
  }

  void _updateSimplexFromRelay() {
    final relay = _relayMgr.activeRelay;
    if (relay != null) {
      setState(() {
        _simplexConfig = SimpleXTransportConfig(
          smp: relay.smp,
          xftp: relay.xftp,
          ice: relay.ice,
          onion: _serverUrl,
          contact: '',
          label: relay.label,
        );
      });
    }
  }

  Widget _buildDesktopNav() {
    return NavigationRail(
      selectedIndex: _selectedIndex,
      onDestinationSelected: (i) => setState(() => _selectedIndex = i),
      labelType: NavigationRailLabelType.all,
      leading: Padding(
        padding: const EdgeInsets.symmetric(vertical: 8),
        child: GestureDetector(
          onTap: () => showIsleDeclaration(context),
          child: Column(
            children: [
              ClipOval(child: Image.asset('assets/images/emblem.png', width: 32, height: 32)),
              Text('The Island', style: Theme.of(context).textTheme.labelSmall),
            ],
          ),
        ),
      ),
      destinations: const [
        NavigationRailDestination(icon: Icon(Icons.dashboard), label: Text('Dashboard')),
        NavigationRailDestination(icon: Icon(Icons.account_balance_wallet), label: Text('Wallet')),
        NavigationRailDestination(icon: Icon(Icons.folder), label: Text('Vault')),
        NavigationRailDestination(icon: Icon(Icons.sell), label: Text('Market')),
        NavigationRailDestination(icon: Icon(Icons.point_of_sale), label: Text('POS')),
        NavigationRailDestination(icon: Icon(Icons.radio), label: Text('Radio')),
        NavigationRailDestination(icon: Icon(Icons.account_tree), label: Text('Royal')),
        NavigationRailDestination(icon: Icon(Icons.chat), label: Text('Chat')),
      ],
      trailing: Expanded(
        child: Align(
          alignment: Alignment.bottomCenter,
          child: Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: _torNavIcon(),
          ),
        ),
      ),
    );
  }

  Widget _buildMobileNav() {
    return NavigationBar(
      selectedIndex: _selectedIndex,
      onDestinationSelected: (i) => setState(() => _selectedIndex = i),
      destinations: const [
        NavigationDestination(icon: Icon(Icons.dashboard), label: 'Dashboard'),
        NavigationDestination(icon: Icon(Icons.account_balance_wallet), label: 'Wallet'),
        NavigationDestination(icon: Icon(Icons.folder), label: 'Vault'),
        NavigationDestination(icon: Icon(Icons.sell), label: 'Market'),
        NavigationDestination(icon: Icon(Icons.point_of_sale), label: 'POS'),
        NavigationDestination(icon: Icon(Icons.radio), label: 'Radio'),
        NavigationDestination(icon: Icon(Icons.account_tree), label: 'Royal'),
        NavigationDestination(icon: Icon(Icons.chat), label: 'Chat'),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final isDesktop = MediaQuery.of(context).size.width > 720;
    return CallbackShortcuts(
      bindings: {
        SingleActivator(LogicalKeyboardKey.keyQ, control: true): () {
          widget.radioPlayer.stop();
          Future.delayed(const Duration(milliseconds: 200), () => exit(0));
        },
        SingleActivator(LogicalKeyboardKey.keyR, control: true): () => _connect(),
        SingleActivator(LogicalKeyboardKey.digit1, control: true): () => setState(() => _selectedIndex = 0),
        SingleActivator(LogicalKeyboardKey.digit2, control: true): () => setState(() => _selectedIndex = 1),
        SingleActivator(LogicalKeyboardKey.digit3, control: true): () => setState(() => _selectedIndex = 2),
        SingleActivator(LogicalKeyboardKey.digit4, control: true): () => setState(() => _selectedIndex = 3),
        SingleActivator(LogicalKeyboardKey.digit5, control: true): () => setState(() => _selectedIndex = 4),
        SingleActivator(LogicalKeyboardKey.digit6, control: true): () => setState(() => _selectedIndex = 5),
        SingleActivator(LogicalKeyboardKey.digit7, control: true): () => setState(() => _selectedIndex = 6),
        SingleActivator(LogicalKeyboardKey.digit8, control: true): () => setState(() => _selectedIndex = 7),
        SingleActivator(LogicalKeyboardKey.comma, control: true): () => setState(() => _showSettings = !_showSettings),
        SingleActivator(LogicalKeyboardKey.keyL, control: true): () { _resetInactivityTimer(); widget.onLock?.call(); },
      },
      child: Focus(
        autofocus: true,
        onKeyEvent: (_, __) { _resetInactivityTimer(); return KeyEventResult.ignored; },
        child: Listener(
          onPointerDown: (_) => _resetInactivityTimer(),
          onPointerMove: (_) => _resetInactivityTimer(),
          onPointerUp: (_) => _resetInactivityTimer(),
          child: Scaffold(
            body: Column(
              children: [
                TopBar(
                  player: widget.radioPlayer,
                  torMgr: _torMgr,
                  vpnMgr: widget.vpnMgr,
                  healthSmp: _healthSmp,
                  healthXftp: _healthXftp,
                  healthRadio: _healthRadio,
                  nodeHealth: _nodeHealth == NodeHealth.up ? TopBarNodeHealth.up : _nodeHealth == NodeHealth.reconnecting ? TopBarNodeHealth.reconnecting : TopBarNodeHealth.down,
                  onLock: widget.onLock,
                  onSwitchAccount: () => _confirmSwitchAccount(context),
                  onOpenSettings: () => setState(() => _showSettings = !_showSettings),
                ),
                Expanded(
                  child: Row(
                    children: [
                      if (isDesktop) _buildDesktopNav(),
                      if (isDesktop) const VerticalDivider(width: 1),
                      Expanded(
                        child: _showSettings ? _buildSettingsPanel() : _buildBody(),
                      ),
                    ],
                  ),
                ),
                if (!isDesktop) _buildMobileNav(),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void _confirmSwitchAccount(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Switch Account'),
        content: const Text('This will log out of the current account and clear saved keys. '
            'You will need your seed phrase to restore it later. Continue?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: Colors.red),
            onPressed: () {
              Navigator.pop(ctx);
              widget.onSwitchAccount?.call();
            },
            child: const Text('Log Out & Switch'),
          ),
        ],
      ),
    );
  }

  Widget _torNavIcon() {
    return Tooltip(
      message: _torMgr.statusLabel,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 2),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 8,
              height: 8,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: _torMgr.isRunning ? Colors.green : (_torMgr.status == TorStatus.error ? Colors.orange : Colors.red),
              ),
            ),
            const SizedBox(width: 4),
            Text('TOR', style: TextStyle(fontSize: 9, color: _torMgr.isRunning ? Colors.green : Colors.red)),
          ],
        ),
      ),
    );
  }

  Widget _healthDot(String label, bool ok) {
    return Tooltip(
      message: '$label ${ok ? "connected" : "offline"}',
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 1),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 6,
              height: 6,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: ok ? Colors.green : Colors.grey,
              ),
            ),
            const SizedBox(width: 3),
            Text(label, style: TextStyle(fontSize: 7, color: ok ? Colors.green : Colors.grey)),
          ],
        ),
      ),
    );
  }

  Widget _torStatusIcon(TorStatus status) {
    switch (status) {
      case TorStatus.running:
        return const Icon(Icons.check_circle, color: Colors.green, size: 16);
      case TorStatus.stopped:
        return const Icon(Icons.cancel, color: Colors.red, size: 16);
      case TorStatus.error:
        return const Icon(Icons.error, color: Colors.orange, size: 16);
      case TorStatus.unknown:
        return const Icon(Icons.help, color: Colors.grey, size: 16);
    }
  }

  Widget _buildSettingsPanel() {
    final urlCtrl = _urlCtrl!;
    final keyCtrl = _keyCtrl!;
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text('Settings', style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 24),

        // Account card
        Card(
          child: Padding(
            padding: const EdgeInsets.all(14),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    ClipOval(child: Image.asset('assets/images/emblem.png', width: 32, height: 32)),
                    const SizedBox(width: 10),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('Current Account', style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold)),
                        Text(widget.pubkey.length > 20 ? '${widget.pubkey.substring(0, 20)}...' : widget.pubkey,
                            style: TextStyle(fontSize: 10, fontFamily: 'monospace', color: Colors.grey[500])),
                      ],
                    ),
                  ],
                ),
                const Divider(height: 16),
                Text('Public Key', style: Theme.of(context).textTheme.labelSmall),
                const SizedBox(height: 4),
                SelectableText(widget.pubkey,
                    style: const TextStyle(fontSize: 11, fontFamily: 'monospace', color: Colors.blueGrey)),
                const SizedBox(height: 10),
                Text('Private Key (local only)', style: Theme.of(context).textTheme.labelSmall),
                const SizedBox(height: 4),
                SelectableText(widget.privkey,
                    style: const TextStyle(fontSize: 11, fontFamily: 'monospace', color: Colors.grey)),
                if (_mnemonic.isNotEmpty) ...[
                  const SizedBox(height: 10),
                  Row(
                    children: [
                      Icon(Icons.warning_amber, size: 14, color: Colors.orange.shade700),
                      const SizedBox(width: 6),
                      Text('Seed Phrase', style: Theme.of(context).textTheme.labelSmall),
                      const Spacer(),
                      TextButton.icon(
                        icon: Icon(_showMnemonic ? Icons.visibility_off : Icons.visibility, size: 14),
                        label: Text(_showMnemonic ? 'Hide' : 'Show', style: const TextStyle(fontSize: 10)),
                        onPressed: () => setState(() => _showMnemonic = !_showMnemonic),
                        style: TextButton.styleFrom(
                          padding: const EdgeInsets.symmetric(horizontal: 8),
                          minimumSize: Size.zero,
                          tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                        ),
                      ),
                    ],
                  ),
                  if (_showMnemonic)
                    Container(
                      width: double.infinity,
                      margin: const EdgeInsets.only(top: 4),
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: Colors.orange.shade900.withValues(alpha: 0.15),
                        borderRadius: BorderRadius.circular(6),
                        border: Border.all(color: Colors.orange.shade700.withValues(alpha: 0.3)),
                      ),
                      child: SelectableText(_mnemonic,
                          style: const TextStyle(fontSize: 12, fontFamily: 'monospace', height: 1.5, color: Colors.orange)),
                    ),
                ],
                const SizedBox(height: 10),
                Row(
                  children: [
                    Expanded(
                      child: OutlinedButton.icon(
                        icon: const Icon(Icons.copy, size: 14),
                        label: const Text('Copy Pubkey', style: TextStyle(fontSize: 11)),
                        onPressed: () {
                          Clipboard.setData(ClipboardData(text: widget.pubkey));
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('Pubkey copied'), duration: Duration(seconds: 1)),
                          );
                        },
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: OutlinedButton.icon(
                        icon: const Icon(Icons.logout, size: 14, color: Colors.red),
                        label: const Text('Log Out', style: TextStyle(fontSize: 11, color: Colors.red)),
                        onPressed: () => _confirmSwitchAccount(context),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 16),

        // Server config
        Text('Server', style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: 8),
        TextField(
          controller: urlCtrl,
          decoration: const InputDecoration(
            labelText: 'Server URL',
            hintText: 'http://onion:80',
            border: OutlineInputBorder(),
            prefixIcon: Icon(Icons.link),
            isDense: true,
          ),
        ),
        const SizedBox(height: 10),
        TextField(
          controller: keyCtrl,
          decoration: const InputDecoration(
            labelText: 'Ed25519 Public Key',
            hintText: 'Enter your public key...',
            border: OutlineInputBorder(),
            prefixIcon: Icon(Icons.vpn_key),
            isDense: true,
          ),
        ),
        const SizedBox(height: 12),
        FilledButton.icon(
          onPressed: () {
            _setServerUrl(urlCtrl.text.trim());
            _setPubkey(keyCtrl.text.trim());
            if (mounted) setState(() {});
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('Config updated'), backgroundColor: Colors.green),
            );
          },
          icon: const Icon(Icons.save),
          label: const Text('Save'),
        ),
        const SizedBox(height: 16),
        Text('Network', style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: 8),
        Card(
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    _torStatusIcon(_torMgr.status),
                    const SizedBox(width: 8),
                    Text(_torMgr.statusLabel, style: Theme.of(context).textTheme.bodyMedium),
                    const Spacer(),
                    FilledButton.tonal(
                      onPressed: () async {
                        if (_torMgr.isRunning) {
                          await _torMgr.stopTor();
                        } else {
                          await _torMgr.startTor();
                        }
                        if (mounted) setState(() {});
                      },
                      child: Text(_torMgr.isRunning ? 'Disconnect' : 'Connect'),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                CheckboxListTile(
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  title: const Text('Auto-connect TOR at startup', style: TextStyle(fontSize: 13)),
                  value: _torAutoConnect,
                  onChanged: (v) async {
                    if (v != null) {
                      _torAutoConnect = v;
                      await _torMgr.setAutoConnect(v);
                      if (mounted) setState(() {});
                    }
                  },
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 8),
        OutlinedButton.icon(
          icon: const Icon(Icons.settings_ethernet, size: 18),
          label: const Text('Select Relay', style: TextStyle(fontSize: 13)),
          onPressed: () async {
            final result = await showDialog<String>(
              context: context,
              builder: (_) => RelaySelectorDialog(manager: _relayMgr, httpClient: _httpClient),
            );
            if (result != null && mounted) {
              _updateSimplexFromRelay();
            }
          },
        ),
        if (_simplexConfig.hasSMP) ...[
          const SizedBox(height: 16),
          Text('SimpleX Transport', style: Theme.of(context).textTheme.titleSmall),
          const SizedBox(height: 8),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('SMP:', style: Theme.of(context).textTheme.labelSmall),
                  SelectableText(_simplexConfig.smp, style: const TextStyle(fontSize: 10, fontFamily: 'monospace')),
                  if (_simplexConfig.xftp.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    Text('XFTP:', style: Theme.of(context).textTheme.labelSmall),
                    SelectableText(_simplexConfig.xftp, style: const TextStyle(fontSize: 10, fontFamily: 'monospace')),
                  ],
                  if (_simplexConfig.onion.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    Text('Onion:', style: Theme.of(context).textTheme.labelSmall),
                    SelectableText(_simplexConfig.onion, style: const TextStyle(fontSize: 10, fontFamily: 'monospace')),
                  ],
                ],
              ),
            ),
          ),
        ],
        const SizedBox(height: 16),
        Text('Keyboard shortcuts:', style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: 8),
        _shortcut('Ctrl+1-6', 'Switch tabs'),
        _shortcut('Ctrl+R', 'Reconnect'),
        _shortcut('Ctrl+Q', 'Quit'),
        _shortcut('Ctrl+,', 'Settings'),
      ],
    );
  }

  Widget _shortcut(String key, String desc) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            decoration: BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              borderRadius: BorderRadius.circular(4),
            ),
            child: Text(key, style: const TextStyle(fontFamily: 'monospace', fontSize: 12)),
          ),
          const SizedBox(width: 12),
          Text(desc),
        ],
      ),
    );
  }

  Widget _buildBody() {
    Widget content;
    switch (_selectedIndex) {
      case 0:
        content = DashboardScreen(
          serverUrl: _serverUrl,
          pubkey: _pubkey,
        );
        break;
      case 1:
        content = _balance != null
            ? WalletScreen(api: _api, serverUrl: _serverUrl, httpClient: _httpClient)
            : _notConnected('Connect on Dashboard first');
        break;
      case 2:
        content = _balance != null
            ? VaultScreen(api: _api)
            : _notConnected('Connect on Dashboard first');
        break;
      case 3:
        content = _balance != null
            ? MarketScreen(serverUrl: _serverUrl, httpClient: _httpClient)
            : _notConnected('Connect on Dashboard first');
        break;
      case 4:
        content = _balance != null
            ? PosScreen(api: _api)
            : _notConnected('Connect on Dashboard first');
        break;
      case 5:
        content = RadioScreen(
          apiBase: _serverUrl,
          httpClient: _httpClient,
          radioPlayer: widget.radioPlayer,
        );
        break;
      case 6:
        content = RoyalScreen(royal: _api.royal, pubkey: _pubkey, httpClient: _httpClient);
        break;
      case 7:
        content = SimplexChatScreen(serverUrl: _serverUrl, httpClient: _httpClient);
        break;
      default:
        content = const Center(child: Text('Unknown tab'));
    }
    return Stack(
      children: [
        Positioned.fill(
          child: Opacity(
            opacity: 0.05,
            child: Image.asset('assets/images/emblem.png', fit: BoxFit.contain, alignment: Alignment.center),
          ),
        ),
        content,
      ],
    );
  }

  Widget _notConnected(String msg) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.link_off, size: 64, color: Colors.grey),
          const SizedBox(height: 16),
          Text(msg, style: Theme.of(context).textTheme.bodyLarge),
          const SizedBox(height: 16),
          FilledButton.tonal(
            onPressed: () { setState(() => _selectedIndex = 0); },
            child: const Text('Go to Dashboard'),
          ),
        ],
      ),
    );
  }
}
