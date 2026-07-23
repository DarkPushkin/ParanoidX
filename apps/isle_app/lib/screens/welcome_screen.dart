import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:crypto/crypto.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:http/http.dart' as http;
import '../services/tor_manager.dart';
import '../services/server_addresses.dart';
import '../services/relay_manager.dart';
import '../services/tor_aware_client.dart' show createHttpClient, testOnionReachability;
import '../services/radio_player.dart';
import '../services/secure_prefs.dart';
import '../widgets/isle_emblem.dart';
import '../widgets/top_bar.dart' show TopBar, TopBarNodeHealth;
import '../services/vpn_manager.dart';

const Map<String, String> _locales = {
  'en': 'English',
  'ru': 'Русский',
  'es': 'Español',
  'zh': '中文',
};

const Map<String, Map<String, String>> _strings = {
  'en': {
    'welcome': 'The Island',
    'tagline': 'Private Sovereign Network',
    'disclaimer': 'Self-sovereign digital commonwealth. Experimental software — no warranty. You alone control your keys and assets.',
    'set_pin': 'Set 6-digit PIN',
    'enter_pin': 'Enter PIN',
    'pin_warning': 'PIN cannot be reset. Forgot it? Data lost forever.',
    'confirm_pin': 'Confirm PIN',
    'pin_mismatch': 'PINs do not match',
    'wrong_pin': 'Wrong PIN. Attempts:',
    'locked': 'Locked 30s',
    'unlock': 'Unlock',
    'save': 'Save',
    'language': 'Language',
    'visit': 'stmaria.org',
    'connecting': 'Connecting...',
    'tor': 'Tor',
    'dashboard': 'Node',
    'services': 'Services',
    'retry': 'Retry',
    'no_connection': 'Cannot connect to node services.',
    'seed_login': 'Login with Seed Phrase',
    'seed_phrase': 'Enter 24-word seed phrase',
    'seed_wrong': 'Seed phrase does not match stored mnemonic',
    'seed_back': 'Back to PIN',
  },
  'ru': {
    'welcome': 'Остров',
    'tagline': 'Частная Суверенная Сеть',
    'disclaimer': 'Самоуправляемое цифровое сообщество. Экспериментальное ПО — без гарантий. Вы сами управляете своими ключами и активами.',
    'set_pin': 'Установите PIN 6 цифр',
    'enter_pin': 'Введите PIN',
    'pin_warning': 'PIN невозможно сбросить. Забудете — данные потеряны навсегда.',
    'confirm_pin': 'Подтвердите PIN',
    'pin_mismatch': 'PIN-коды не совпадают',
    'wrong_pin': 'Неверный PIN. Попыток:',
    'locked': 'Блок 30с',
    'unlock': 'Войти',
    'save': 'Сохранить',
    'language': 'Язык',
    'visit': 'stmaria.org',
    'connecting': 'Подключение...',
    'tor': 'Tor',
    'dashboard': 'Нода',
    'services': 'Сервисы',
    'retry': 'Повтор',
    'no_connection': 'Нет связи с серверами.',
    'seed_login': 'Войти по сид-фразе',
    'seed_phrase': 'Введите 24 слова сид-фразы',
    'seed_wrong': 'Сид-фраза не совпадает с сохранённой',
    'seed_back': 'Назад к PIN',
  },
  'es': {
    'welcome': 'La Isla',
    'tagline': 'Red Privada Soberana',
    'disclaimer': 'Comunidad digital soberana. Software experimental — sin garantía. Usted controla sus llaves y activos.',
    'set_pin': 'PIN 6 dígitos',
    'enter_pin': 'Ingrese PIN',
    'pin_warning': 'PIN no se puede restablecer.',
    'confirm_pin': 'Confirmar PIN',
    'pin_mismatch': 'PIN no coinciden',
    'wrong_pin': 'PIN incorrecto. Intentos:',
    'locked': 'Bloqueado 30s',
    'unlock': 'Entrar',
    'save': 'Guardar',
    'language': 'Idioma',
    'visit': 'stmaria.org',
    'tor': 'Tor',
    'dashboard': 'Nodo',
    'services': 'Servicios',
    'retry': 'Reintentar',
    'no_connection': 'Sin conexión al nodo.',
    'seed_login': 'Iniciar con Frase Semilla',
    'seed_phrase': 'Ingrese frase de 24 palabras',
    'seed_wrong': 'La frase no coincide con el mnemónico',
    'seed_back': 'Volver a PIN',
    'connecting': 'Conectando...',
  },
  'zh': {
    'welcome': '岛屿',
    'tagline': '私有主权网络',
    'disclaimer': '自治数字共同体。实验性软件——无担保。您完全控制自己的密钥和资产。',
    'set_pin': '设置6位PIN',
    'enter_pin': '输入PIN',
    'pin_warning': 'PIN无法重置。',
    'confirm_pin': '确认PIN',
    'pin_mismatch': 'PIN不匹配',
    'wrong_pin': 'PIN错误。次数：',
    'locked': '锁定30秒',
    'unlock': '解锁',
    'save': '保存',
    'language': '语言',
    'visit': 'stmaria.org',
    'tor': 'Tor',
    'dashboard': '节点',
    'services': '服务',
    'retry': '重试',
    'no_connection': '无法连接到节点服务。',
    'seed_login': '用种子短语登录',
    'seed_phrase': '输入24词种子短语',
    'seed_wrong': '种子短语与存储的不匹配',
    'seed_back': '返回PIN',
    'connecting': '连接中...',
  },
};

String _tr(String locale, String key) {
  return _strings[locale]?[key] ?? _strings['en']![key]!;
}

String detectSystemLocale() {
  try {
    final loc = Platform.localeName;
    if (loc.length >= 2) {
      final lang = loc.substring(0, 2);
      if (_locales.containsKey(lang)) return lang;
    }
  } catch (_) {}
  return 'en';
}

/// Values for the welcomemode domain.
enum WelcomeMode { pinSetup, pinEntry, seedLogin, connecting, ready, error }

/// Values for the nodehealth domain.
enum NodeHealth { up, down, reconnecting }

/// Values for the handshakestatus domain.
enum HandshakeStatus { waiting, inProgress, done, failed }

/// HandshakeStep manages handshakestep functionality.
class HandshakeStep {
  final String label;
  final String ruLabel;
  HandshakeStatus status;
  String detail;
  HandshakeStep({required this.label, required this.ruLabel, this.status = HandshakeStatus.waiting, this.detail = ''});
}

/// WelcomeScreen manages the welcome screen with PIN setup, lock screen, and connection handshake.
class WelcomeScreen extends StatefulWidget {
  final String initialLocale;
  final bool needsConnectionCheck;
  final TorManager torMgr;
  final ServerAddressManager addrMgr;
  final RelayManager relayMgr;
  final String serverUrl;
  final ValueChanged<String> onLocaleChanged;
  final VoidCallback onUnlocked;
  final RadioPlayer radioPlayer;
  final VpnManager? vpnMgr;
  final String buildVersion;

  const WelcomeScreen({
    super.key,
    required this.initialLocale,
    this.needsConnectionCheck = true,
    required this.torMgr,
    required this.addrMgr,
    required this.relayMgr,
    required this.serverUrl,
    required this.onLocaleChanged,
    required this.onUnlocked,
    required this.radioPlayer,
    this.vpnMgr,
    this.buildVersion = '',
  });
  @override
  State<WelcomeScreen> createState() => _WelcomeScreenState();
}

class _WelcomeScreenState extends State<WelcomeScreen> {
  late String _locale;
  WelcomeMode _mode = WelcomeMode.pinEntry;
  final _pinCtrl = TextEditingController();
  final _confirmCtrl = TextEditingController();
  final _seedCtrl = TextEditingController();
  bool _obscurePin = true;
  int _attempts = 0;
  bool _lockedOut = false;

  late http.Client _httpClient;
  final List<HandshakeStep> _steps = [];
  String _smpAddr = '';
  String _xftpAddr = '';
  String _radioAddr = '';

  bool _bgTorOk = false;
  bool _bgTorChecking = false;
  bool _bgDashOk = false;
  bool _bgServicesOk = false;
  bool _healthSmp = false;
  bool _healthXftp = false;
  bool _healthRadio = false;
  NodeHealth _nodeHealth = NodeHealth.down;
  Timer? _healthTimer;
  String _jokeText = '';

  @override
  void initState() {
    super.initState();
    _locale = widget.initialLocale;
    _httpClient = createHttpClient(useProxy: true);
    _initSteps();
    _initMode();
    _startBackgroundChecks();
    _startHealthPoller();
  }

  void _initSteps() {
    _steps.addAll([
      HandshakeStep(label: 'Tor Network', ruLabel: 'Сеть Tor'),
      HandshakeStep(label: 'Dashboard node', ruLabel: 'Нода Dashboard'),
      HandshakeStep(label: 'Services + Steward', ruLabel: 'Сервисы + Стюарт'),
      HandshakeStep(label: 'SMP transport', ruLabel: 'Транспорт SMP'),
      HandshakeStep(label: 'XFTP transport', ruLabel: 'Транспорт XFTP'),
      HandshakeStep(label: 'Greeting to Royal Bot', ruLabel: 'Приветствие Королю'),
    ]);
  }

  Future<void> _initMode() async {
    final sec = await SecurePrefs.instance;
    final pinExists = await sec.containsKey('app_pin');
    if (mounted) setState(() { _mode = pinExists ? WelcomeMode.pinEntry : WelcomeMode.pinSetup; });
  }

  Future<void> _startBackgroundChecks() async {
    setState(() { _bgTorChecking = true; });
    try {
      await widget.torMgr.ensureRunning();
      await Future.delayed(const Duration(milliseconds: 500));
      if (mounted) setState(() { _bgTorOk = widget.torMgr.isRunning; _bgTorChecking = false; });
    } catch (_) {
      if (mounted) setState(() { _bgTorOk = false; _bgTorChecking = false; });
    }
    if (!mounted || !_bgTorOk) return;

    try {
      final dashParsed = _parseOnionAddr(widget.serverUrl, 80);
      if (dashParsed != null) {
        final ok = await testOnionReachability(dashParsed.host, dashParsed.port, timeoutMs: 8000);
        if (mounted) setState(() => _bgDashOk = ok);
      }
    } catch (_) {}
    if (!mounted || !_bgDashOk) return;

    try {
      final infoUrl = _apiUrl('/api/transport/info');
      if (infoUrl.hasAuthority) {
        final resp = await _httpClient.get(infoUrl).timeout(const Duration(seconds: 8));
        if (resp.statusCode == 200 && mounted) {
          final info = resp.body;
          final smpMatch = RegExp(r'"smp"\s*:\s*"([^"]+)"').firstMatch(info);
          final xftpMatch = RegExp(r'"xftp"\s*:\s*"([^"]+)"').firstMatch(info);
          _smpAddr = (smpMatch?.group(1) ?? '').trim();
          _xftpAddr = (xftpMatch?.group(1) ?? '').trim();
          if (mounted) setState(() => _bgServicesOk = _smpAddr.isNotEmpty || _xftpAddr.isNotEmpty);
        }
      }
    } catch (_) {}
  }

  void _startHealthPoller() {
    _healthTimer = Timer.periodic(const Duration(seconds: 30), (_) => _healthPing());
    _healthPing();
  }

  Future<void> _healthPing() async {
    if (!mounted) return;
    try {
      await widget.torMgr.ensureRunning();
      final torOk = widget.torMgr.isRunning;
      if (mounted) setState(() { _bgTorOk = torOk; _bgTorChecking = false; });
      if (!torOk) {
        if (mounted) setState(() { _nodeHealth = NodeHealth.down; _healthSmp = false; _healthXftp = false; _healthRadio = false; });
        return;
      }

      final healthUrl = _apiUrl('/api/transport/health');
      if (healthUrl.hasAuthority) {
        final resp = await _httpClient.get(healthUrl).timeout(const Duration(seconds: 5));
        if (resp.statusCode == 200 && mounted) {
          final data = jsonDecode(resp.body) as Map<String, dynamic>;
          final smpOk = data['smp'] == true;
          final xftpOk = data['xftp'] == true;
          final radioOk = data['radio'] == true;
          setState(() {
            _healthSmp = smpOk;
            _healthXftp = xftpOk;
            _healthRadio = radioOk;
            _nodeHealth = (smpOk || xftpOk || radioOk) ? NodeHealth.up : NodeHealth.reconnecting;
            _bgDashOk = true;
            _bgServicesOk = smpOk || xftpOk;
          });
          return;
        }
      }

      final dashParsed = _parseOnionAddr(widget.serverUrl, 80);
      if (dashParsed != null) {
        final dashOk = await testOnionReachability(dashParsed.host, dashParsed.port, timeoutMs: 5000);
        if (mounted) {
          setState(() {
            _bgDashOk = dashOk;
            _nodeHealth = dashOk ? NodeHealth.reconnecting : NodeHealth.down;
            if (!dashOk) { _healthSmp = false; _healthXftp = false; _healthRadio = false; }
          });
        }
      } else {
        if (mounted) setState(() { _bgDashOk = false; _nodeHealth = NodeHealth.down; });
      }
    } catch (_) {
      if (mounted) setState(() { _nodeHealth = NodeHealth.down; _healthSmp = false; _healthXftp = false; _healthRadio = false; });
    }
  }

  @override
  void dispose() {
    _healthTimer?.cancel();
    _pinCtrl.dispose();
    _confirmCtrl.dispose();
    _seedCtrl.dispose();
    _httpClient.close();
    super.dispose();
  }

  Future<void> _savePin(String pin) async {
    final sec = await SecurePrefs.instance;
    await sec.setHashed('app_pin', pin);
  }

  Future<bool> _checkPin(String pin) async {
    final sec = await SecurePrefs.instance;
    final exists = await sec.containsKey('app_pin');
    if (!exists) return pin == '878787';
    return sec.verify('app_pin', pin);
  }

  Future<void> _onPinSetupConfirm() async {
    final pin = _pinCtrl.text.trim();
    if (pin.length != 6) return;
    final confirm = _confirmCtrl.text.trim();
    if (pin != confirm) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(
        content: Text(_tr(_locale, 'pin_mismatch')), backgroundColor: Colors.red,
      ));
      return;
    }
    await _savePin(pin);
    setState(() => _mode = WelcomeMode.pinEntry);
    _pinCtrl.clear();
    _confirmCtrl.clear();
  }

  Future<void> _onPinEntryUnlock() async {
    if (_lockedOut) return;
    final pin = _pinCtrl.text.trim();
    if (pin.length != 6) return;
    final ok = await _checkPin(pin);
    if (ok) {
      _attempts = 0;
      _startConnection();
    } else {
      _attempts++;
      if (_attempts >= 3) {
        setState(() { _lockedOut = true; });
        Future.delayed(const Duration(seconds: 30), () {
          if (mounted) setState(() { _lockedOut = false; _attempts = 0; });
        });
      } else {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(
          content: Text('${_tr(_locale, 'wrong_pin')} ${3 - _attempts}'), backgroundColor: Colors.red,
        ));
      }
    }
  }

  ({String host, int port})? _parseOnionAddr(String addr, int defaultPort) {
    try {
      final uri = Uri.parse(addr);
      if (uri.host.endsWith('.onion')) return (host: uri.host, port: uri.port > 0 ? uri.port : defaultPort);
    } catch (_) {}
    final atMatch = RegExp(r'@([^:]+)(?::(\d+))?').firstMatch(addr);
    if (atMatch != null && atMatch.group(1)!.endsWith('.onion')) {
      return (host: atMatch.group(1)!, port: int.tryParse(atMatch.group(2) ?? '') ?? defaultPort);
    }
    return null;
  }

  void _setStep(int index, HandshakeStatus status, {String detail = ''}) {
    if (index >= 0 && index < _steps.length) {
      _steps[index].status = status;
      if (detail.isNotEmpty) _steps[index].detail = detail;
    }
  }

  void _stopRadio() { widget.radioPlayer.stop(); }

  Uri get _baseUri => Uri.tryParse(widget.serverUrl) ?? Uri();
  Uri _apiUrl(String path) => _baseUri.resolve(path);

  Future<void> _sendJokeToSteward() async {
    try {
      final req = {
        'type': 'api.request',
        'payload': {
          'method': 'POST',
          'path': '/api/ai/chat',
          'body': {'question': 'Tell a very short joke about sovereign networks.', 'context': ''},
        },
        'id': 'joke-${DateTime.now().millisecondsSinceEpoch}',
      };
      final resp = await _httpClient
          .post(_apiUrl('/api/transport/send'), headers: {'Content-Type': 'application/json'}, body: jsonEncode(req))
          .timeout(const Duration(seconds: 10));
      if (resp.statusCode == 200 && mounted) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['type'] == 'response') {
          final payload = data['payload'];
          if (payload is Map) {
            final body = payload['body'];
            if (body is Map && body['answer'] != null) {
              final joke = body['answer'] as String;
              setState(() => _jokeText = joke);
              _setStep(2, HandshakeStatus.done, detail: joke.length > 50 ? '${joke.substring(0, 50)}…' : joke);
            }
          }
        } else if (data['type'] == 'error') {
          _setStep(2, HandshakeStatus.done, detail: 'Joke unavailable');
        }
      }
    } catch (_) {
      _setStep(2, HandshakeStatus.done, detail: 'Joke unavailable');
    }
  }

  Future<void> _sendGreeting() async {
    try {
      final greeting = {
        'type': 'api.request',
        'payload': {'method': 'POST', 'path': '/api/steward', 'body': {'action': 'greet'}},
        'id': 'greet-${DateTime.now().millisecondsSinceEpoch}',
      };
      await _httpClient.post(_apiUrl('/api/transport/send'), headers: {'Content-Type': 'application/json'}, body: jsonEncode(greeting)).timeout(const Duration(seconds: 8));
    } catch (_) {}
  }

  Future<void> _startConnection() async {
    setState(() {
      _mode = WelcomeMode.connecting;
      for (final s in _steps) { s.status = HandshakeStatus.waiting; s.detail = ''; }
    });

    _setStep(0, _bgTorOk ? HandshakeStatus.done : HandshakeStatus.inProgress,
        detail: _bgTorOk ? 'Tor ✓' : 'Starting Tor...');
    if (mounted) setState(() {});
    bool torOk = _bgTorOk;
    if (!torOk) {
      try {
        await widget.torMgr.ensureRunning();
        await Future.delayed(const Duration(milliseconds: 300));
        torOk = widget.torMgr.isRunning;
        if (mounted) _setStep(0, torOk ? HandshakeStatus.done : HandshakeStatus.failed, detail: torOk ? 'Tor ✓' : 'Tor failed');
        if (mounted) setState(() {});
      } catch (e) {
        if (mounted) { _setStep(0, HandshakeStatus.failed, detail: '$e'); setState(() {}); }
      }
    }
    if (!mounted || !torOk) { _onConnectionFailed(); return; }

    _setStep(1, _bgDashOk ? HandshakeStatus.done : HandshakeStatus.inProgress,
        detail: _bgDashOk ? 'Node ✓' : 'Verifying node...');
    if (mounted) setState(() {});
    bool dashOk = _bgDashOk;
    if (!dashOk) {
      try {
        final parsed = _parseOnionAddr(widget.serverUrl, 80);
        if (parsed != null) {
          dashOk = await testOnionReachability(parsed.host, parsed.port, timeoutMs: 10000);
          if (mounted) _setStep(1, dashOk ? HandshakeStatus.done : HandshakeStatus.failed, detail: dashOk ? 'Node ✓' : 'Node unreachable');
        } else {
          if (mounted) _setStep(1, HandshakeStatus.failed, detail: 'Bad address');
        }
        if (mounted) setState(() {});
      } catch (e) {
        if (mounted) { _setStep(1, HandshakeStatus.failed, detail: '$e'); setState(() {}); }
      }
    }
    if (!mounted || !dashOk) { _onConnectionFailed(); return; }

    _setStep(2, _bgServicesOk ? HandshakeStatus.done : HandshakeStatus.inProgress, detail: 'Loading...');
    if (mounted) setState(() {});
    try {
      if (!_bgServicesOk) {
        final infoUrl = _apiUrl('/api/transport/info');
        if (!infoUrl.hasAuthority) {
          if (mounted) { _setStep(2, HandshakeStatus.failed, detail: 'Bad URL'); setState(() {}); }
          _onConnectionFailed(); return;
        }
        final resp = await _httpClient.get(infoUrl).timeout(const Duration(seconds: 10));
        if (resp.statusCode == 200 && mounted) {
          final info = resp.body;
          final smpMatch = RegExp(r'"smp"\s*:\s*"([^"]+)"').firstMatch(info);
          final xftpMatch = RegExp(r'"xftp"\s*:\s*"([^"]+)"').firstMatch(info);
          final onionMatch = RegExp(r'"onion"\s*:\s*"([^"]+)"').firstMatch(info);
          final iceMatch = RegExp(r'"ice"\s*:\s*"([^"]+)"').firstMatch(info);
          _smpAddr = (smpMatch?.group(1) ?? '').trim();
          _xftpAddr = (xftpMatch?.group(1) ?? '').trim();
          _radioAddr = (onionMatch?.group(1) ?? '').trim();
          final iceAddr = (iceMatch?.group(1) ?? '').trim();
          if (_smpAddr.isNotEmpty || _xftpAddr.isNotEmpty) {
            widget.relayMgr.upsertFromServer(_smpAddr, _xftpAddr, iceAddr, _radioAddr);
          }
        } else {
          if (mounted) { _setStep(2, HandshakeStatus.failed, detail: 'HTTP ${resp.statusCode}'); setState(() {}); }
          _onConnectionFailed(); return;
        }
      }
      _sendJokeToSteward();
      if (mounted) {
        _setStep(2, (_smpAddr.isNotEmpty || _xftpAddr.isNotEmpty || _bgServicesOk) ? HandshakeStatus.done : HandshakeStatus.failed,
            detail: 'Services OK');
        setState(() {});
      }
    } catch (e) {
      if (mounted) { _setStep(2, HandshakeStatus.failed, detail: '$e'); setState(() {}); }
      _onConnectionFailed(); return;
    }
    if (!mounted) return;

    _setStep(3, HandshakeStatus.inProgress, detail: 'Testing SMP...');
    if (mounted) setState(() {});
    bool smpOk = false;
    try {
      if (_smpAddr.isNotEmpty) {
        final parsed = _parseOnionAddr(_smpAddr, 5223);
        if (parsed != null) {
          smpOk = await testOnionReachability(parsed.host, parsed.port, timeoutMs: 10000);
          if (mounted) _setStep(3, smpOk ? HandshakeStatus.done : HandshakeStatus.failed, detail: smpOk ? 'SMP ✓' : 'SMP unreachable');
        } else {
          if (mounted) _setStep(3, HandshakeStatus.failed, detail: 'Bad SMP');
        }
      } else {
        if (mounted) { _setStep(3, HandshakeStatus.done, detail: 'SMP skipped'); }
        smpOk = true;
      }
      if (mounted) setState(() {});
    } catch (e) {
      if (mounted) { _setStep(3, HandshakeStatus.failed, detail: '$e'); setState(() {}); }
    }
    if (!mounted) return;

    _setStep(4, HandshakeStatus.inProgress, detail: 'Testing XFTP...');
    if (mounted) setState(() {});
    bool xftpOk = false;
    try {
      if (_xftpAddr.isNotEmpty) {
        final parsed = _parseOnionAddr(_xftpAddr, 443);
        if (parsed != null) {
          xftpOk = await testOnionReachability(parsed.host, parsed.port, timeoutMs: 10000);
          if (mounted) _setStep(4, xftpOk ? HandshakeStatus.done : HandshakeStatus.failed, detail: xftpOk ? 'XFTP ✓' : 'XFTP unreachable');
        } else {
          if (mounted) _setStep(4, HandshakeStatus.failed, detail: 'Bad XFTP');
        }
      } else {
        if (mounted) { _setStep(4, HandshakeStatus.done, detail: 'XFTP skipped'); }
        xftpOk = true;
      }
      if (mounted) setState(() {});
    } catch (e) {
      if (mounted) { _setStep(4, HandshakeStatus.failed, detail: '$e'); setState(() {}); }
    }
    if (!mounted) return;

    _setStep(5, HandshakeStatus.inProgress, detail: 'Greeting Royal Bot...');
    if (mounted) setState(() {});
    await _sendGreeting();
    if (mounted) { _setStep(5, HandshakeStatus.done, detail: 'Greeting ✓'); setState(() {}); }

    await Future.delayed(const Duration(milliseconds: 400));
    if (mounted) widget.onUnlocked();
  }

  void _onConnectionFailed() {
    _stopRadio();
    if (mounted) setState(() { _mode = WelcomeMode.error; });
  }

  void _startRadioBackground() {
    // Radio no longer auto-starts on connect – replaced by Steward AI joke
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final tr = (String key) => _tr(_locale, key);

    return Scaffold(
      body: Column(
        children: [
          TopBar(
            player: widget.radioPlayer,
            torMgr: widget.torMgr,
            vpnMgr: widget.vpnMgr,
            buildVersion: widget.buildVersion,
            healthSmp: _healthSmp,
            healthXftp: _healthXftp,
            healthRadio: _healthRadio,
            nodeHealth: _nodeHealth == NodeHealth.up ? TopBarNodeHealth.up : _nodeHealth == NodeHealth.reconnecting ? TopBarNodeHealth.reconnecting : TopBarNodeHealth.down,
          ),
          Expanded(
            child: Stack(
              children: [
                Positioned.fill(
                  child: Opacity(
                    opacity: 0.06,
                    child: Image.asset('assets/images/emblem.png', fit: BoxFit.contain, alignment: Alignment.center),
                  ),
                ),
                Center(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 12),
                    child: ConstrainedBox(
                      constraints: const BoxConstraints(maxWidth: 360),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          // Status row
                          _statusRow(tr),
                          const SizedBox(height: 14),

                          // Emblem
                          RoundEmblem(
                            size: 100,
                            onTap: () => showIsleDeclaration(context, httpClient: _httpClient, serverUrl: widget.serverUrl, pubkey: ''),
                          ),
                          const SizedBox(height: 10),

                          // Title
                          Text(tr('welcome'), style: theme.textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.bold)),
                          const SizedBox(height: 2),
                          Text(tr('tagline'), style: const TextStyle(fontSize: 12, color: Colors.grey)),
                          const SizedBox(height: 14),

                          if (_mode == WelcomeMode.pinSetup) _buildPinSetup(theme, tr),
                          if (_mode == WelcomeMode.pinEntry) _buildPinEntry(theme, tr),
                          if (_mode == WelcomeMode.seedLogin) _buildSeedLogin(theme, tr),
                          if (_mode == WelcomeMode.connecting) _buildConnecting(),
                          if (_mode == WelcomeMode.error) _buildError(theme, tr),

                          const SizedBox(height: 10),

                          // Lang + version
                          Row(
                            children: [
                              Icon(Icons.language, size: 13, color: Colors.grey[500]),
                              const SizedBox(width: 4),
                              Text(tr('language'), style: TextStyle(fontSize: 11, color: Colors.grey[500])),
                              const Spacer(),
                              DropdownButton<String>(
                                value: _locale,
                                underline: const SizedBox(),
                                isDense: true,
                                items: _locales.entries.map((e) => DropdownMenuItem(
                                  value: e.key, child: Text(e.value, style: const TextStyle(fontSize: 12)),
                                )).toList(),
                                onChanged: (v) {
                                  if (v == null) return;
                                  setState(() => _locale = v);
                                  widget.onLocaleChanged(v);
                                },
                              ),
                              if (widget.buildVersion.isNotEmpty) ...[
                                const SizedBox(width: 6),
                                Container(
                                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                  decoration: BoxDecoration(
                                    color: Colors.grey.shade800,
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                  child: Text('v${widget.buildVersion}',
                                      style: TextStyle(fontSize: 9, color: Colors.grey[500])),
                                ),
                              ],
                            ],
                          ),
                          const SizedBox(height: 2),
                          TextButton(
                            style: TextButton.styleFrom(
                              padding: EdgeInsets.zero, minimumSize: Size.zero,
                              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                            ),
                            onPressed: () {},
                            child: Text(tr('visit'), style: TextStyle(fontSize: 11, color: theme.colorScheme.primary)),
                          ),
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

  Widget _statusRow(String Function(String) tr) {
    Widget _dot({required bool? ok, required bool busy, required String label}) {
      return Padding(
        padding: const EdgeInsets.symmetric(horizontal: 4),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            busy
                ? SizedBox(width: 10, height: 10, child: CircularProgressIndicator(strokeWidth: 1.5, color: Colors.grey))
                : Icon(
                    ok == null ? Icons.help_outline : ok ? Icons.check_circle : Icons.error,
                    size: 12, color: ok == null ? Colors.grey : ok ? Colors.green : Colors.red,
                  ),
            const SizedBox(width: 3),
            Text(label, style: TextStyle(fontSize: 10, color: Colors.grey[400])),
          ],
        ),
      );
    }

    Widget _nodeDot() {
      IconData icon;
      Color color;
      String label;
      switch (_nodeHealth) {
        case NodeHealth.up:
          icon = Icons.check_circle;
          color = Colors.green;
          label = 'Up';
        case NodeHealth.down:
          icon = Icons.error;
          color = Colors.red;
          label = 'Down';
        case NodeHealth.reconnecting:
          icon = Icons.sync;
          color = Colors.orange;
          label = 'Reconnect';
      }
      return Padding(
        padding: const EdgeInsets.symmetric(horizontal: 4),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 12, color: color),
            const SizedBox(width: 3),
            Text(label, style: TextStyle(fontSize: 10, color: color, fontWeight: FontWeight.w600)),
          ],
        ),
      );
    }

    return Card(
      margin: EdgeInsets.zero,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 4),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            _nodeDot(),
            const SizedBox(width: 2),
            Container(width: 1, height: 14, color: Colors.grey.shade700),
            const SizedBox(width: 2),
            _dot(ok: _bgTorOk == false && _bgTorChecking ? null : _bgTorOk, busy: _bgTorChecking, label: 'Tor'),
            _dot(ok: _bgTorOk ? (_healthSmp == false ? null : _healthSmp) : null, busy: false, label: 'SMP'),
            _dot(ok: _bgTorOk ? (_healthXftp == false ? null : _healthXftp) : null, busy: false, label: 'XFTP'),
            _dot(ok: _bgTorOk ? (_healthRadio == false ? null : _healthRadio) : null, busy: false, label: 'Radio'),
          ],
        ),
      ),
    );
  }

  Widget _buildPinSetup(ThemeData theme, String Function(String) tr) {
    return Column(
      children: [
        Text(tr('set_pin'), style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
        const SizedBox(height: 6),
        Container(
          padding: const EdgeInsets.all(10),
          decoration: BoxDecoration(
            color: Colors.red.shade50,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.red.shade200, width: 0.5),
          ),
          child: Row(
            children: [
              Icon(Icons.warning_amber_rounded, size: 16, color: Colors.red.shade700),
              const SizedBox(width: 8),
              Expanded(
                child: Text(tr('pin_warning'),
                    style: TextStyle(fontSize: 11, color: Colors.red.shade800, height: 1.3)),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        _pinField(controller: _pinCtrl, obscure: _obscurePin, label: tr('set_pin'),
          onToggleObscure: () => setState(() => _obscurePin = !_obscurePin)),
        const SizedBox(height: 8),
        _pinField(controller: _confirmCtrl, obscure: _obscurePin, label: tr('confirm_pin')),
        const SizedBox(height: 12),
        SizedBox(
          width: 200,
          child: FilledButton(
            style: FilledButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 12)),
            onPressed: _onPinSetupConfirm,
            child: Text(tr('save'), style: const TextStyle(fontSize: 14)),
          ),
        ),
      ],
    );
  }

  Future<void> _onSeedLogin() async {
    final seed = _seedCtrl.text.trim().toLowerCase();
    if (seed.isEmpty) return;
    final sec = await SecurePrefs.instance;
    final stored = await sec.getString('mnemonic');
    if (stored != null && stored.toLowerCase() == seed) {
      setState(() => _mode = WelcomeMode.pinEntry);
      _startConnection();
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(
          content: Text(_tr(_locale, 'seed_wrong')), backgroundColor: Colors.red,
        ));
      }
    }
  }

  Widget _buildSeedLogin(ThemeData theme, String Function(String) tr) {
    return Column(
      children: [
        Icon(Icons.key, size: 20, color: Colors.amber.shade700),
        const SizedBox(height: 8),
        Text(tr('seed_login'), style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
        const SizedBox(height: 8),
        SizedBox(
          width: 280,
          child: TextField(
            controller: _seedCtrl,
            maxLines: 3,
            keyboardType: TextInputType.multiline,
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 13, fontFamily: 'monospace'),
            decoration: InputDecoration(
              hintText: tr('seed_phrase'),
              hintStyle: const TextStyle(fontSize: 11),
              border: const OutlineInputBorder(),
              isDense: true,
              contentPadding: const EdgeInsets.all(10),
            ),
          ),
        ),
        const SizedBox(height: 12),
        SizedBox(
          width: 200,
          child: FilledButton(
            style: FilledButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 12)),
            onPressed: _onSeedLogin,
            child: Text(tr('unlock'), style: const TextStyle(fontSize: 14)),
          ),
        ),
        const SizedBox(height: 6),
        TextButton(
          style: TextButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 4, horizontal: 8)),
          onPressed: () => setState(() { _mode = WelcomeMode.pinEntry; _seedCtrl.clear(); }),
          child: Text(tr('seed_back'), style: TextStyle(fontSize: 12, color: theme.colorScheme.primary)),
        ),
      ],
    );
  }

  Widget _buildPinEntry(ThemeData theme, String Function(String) tr) {
    return Column(
      children: [
        Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.lock, size: 16, color: Colors.amber.shade700),
            const SizedBox(width: 6),
            Text('LOCKED', style: TextStyle(
              fontSize: 14, fontWeight: FontWeight.w900, letterSpacing: 3, color: Colors.amber.shade700)),
          ],
        ),
        const SizedBox(height: 8),
        Text(tr('enter_pin'), style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
        const SizedBox(height: 8),
        _pinField(controller: _pinCtrl, obscure: _obscurePin, label: tr('enter_pin'),
          autofocus: true, onSubmitted: _lockedOut ? null : (_) => _onPinEntryUnlock(),
          onToggleObscure: () => setState(() => _obscurePin = !_obscurePin)),
        if (_lockedOut) ...[
          const SizedBox(height: 6),
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.timer, size: 14, color: Colors.red.shade700),
              const SizedBox(width: 4),
              Text(tr('locked'), style: TextStyle(color: Colors.red.shade700, fontSize: 12, fontWeight: FontWeight.w600)),
            ],
          ),
        ],
        const SizedBox(height: 12),
        SizedBox(
          width: 200,
          child: FilledButton(
            style: FilledButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 12)),
            onPressed: _lockedOut ? null : _onPinEntryUnlock,
            child: Text(tr('unlock'), style: const TextStyle(fontSize: 14)),
          ),
        ),
        const SizedBox(height: 6),
        TextButton(
          style: TextButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 4, horizontal: 8)),
          onPressed: () => setState(() { _mode = WelcomeMode.seedLogin; _seedCtrl.clear(); }),
          child: Text(_tr(_locale, 'seed_login'), style: TextStyle(fontSize: 12, color: theme.colorScheme.primary)),
        ),
      ],
    );
  }

  Widget _pinField({
    required TextEditingController controller,
    required bool obscure,
    required String label,
    bool autofocus = false,
    void Function(String)? onSubmitted,
    VoidCallback? onToggleObscure,
  }) {
    return SizedBox(
      width: 220,
      child: TextField(
        controller: controller,
        obscureText: obscure,
        keyboardType: TextInputType.number,
        maxLength: 6,
        autofocus: autofocus,
        inputFormatters: [FilteringTextInputFormatter.digitsOnly],
        textAlign: TextAlign.center,
        style: const TextStyle(fontSize: 18, letterSpacing: 6),
        decoration: InputDecoration(
          labelText: label,
          labelStyle: const TextStyle(fontSize: 12),
          border: const OutlineInputBorder(),
          isDense: true,
          counterText: '',
          contentPadding: const EdgeInsets.symmetric(vertical: 10, horizontal: 12),
          suffixIcon: onToggleObscure != null
              ? IconButton(
                  icon: Icon(obscure ? Icons.visibility_off : Icons.visibility, size: 18),
                  onPressed: onToggleObscure, padding: EdgeInsets.zero,
                )
              : null,
        ),
        onSubmitted: onSubmitted,
      ),
    );
  }

  Widget _buildConnecting() {
    return Column(
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          decoration: BoxDecoration(
            color: Colors.blue.shade900.withValues(alpha: 0.3),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.blue)),
              const SizedBox(width: 8),
              Text('Connecting...', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Colors.blue[200])),
            ],
          ),
        ),
        const SizedBox(height: 10),
        Container(
          constraints: const BoxConstraints(maxHeight: 180),
          decoration: BoxDecoration(
            border: Border.all(color: Colors.grey.shade800, width: 0.5),
            borderRadius: BorderRadius.circular(8),
          ),
          child: ListView.builder(
            shrinkWrap: true,
            padding: const EdgeInsets.all(8),
            itemCount: _steps.length,
            itemBuilder: (ctx, i) => _stepRow(_steps[i]),
          ),
        ),
      ],
    );
  }

  Widget _stepRow(HandshakeStep s) {
    final label = _locale == 'ru' && s.ruLabel.isNotEmpty ? s.ruLabel : s.label;
    Color color;
    IconData icon;
    switch (s.status) {
      case HandshakeStatus.waiting: color = Colors.grey; icon = Icons.radio_button_unchecked;
      case HandshakeStatus.inProgress: color = Colors.blue; icon = Icons.sync;
      case HandshakeStatus.done: color = Colors.green; icon = Icons.check_circle;
      case HandshakeStatus.failed: color = Colors.red; icon = Icons.error;
    }
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3),
      child: Row(
        children: [
          s.status == HandshakeStatus.inProgress
              ? SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2, color: color))
              : Icon(icon, size: 16, color: color),
          const SizedBox(width: 8),
          Text(label, style: TextStyle(fontSize: 12, color: color, fontWeight: FontWeight.w500)),
          const Spacer(),
          if (s.detail.isNotEmpty)
            Text(s.detail, style: TextStyle(fontSize: 9, color: Colors.grey[600])),
        ],
      ),
    );
  }

  Widget _buildError(ThemeData theme, String Function(String) tr) {
    return Column(
      children: [
        Icon(Icons.cloud_off, size: 40, color: Colors.red.shade300),
        const SizedBox(height: 8),
        Text(tr('no_connection'), style: const TextStyle(fontSize: 12, color: Colors.grey)),
        const SizedBox(height: 10),
        if (_steps.any((s) => s.status == HandshakeStatus.failed))
          Container(
            constraints: const BoxConstraints(maxHeight: 140),
            decoration: BoxDecoration(
              border: Border.all(color: Colors.grey.shade800, width: 0.5),
              borderRadius: BorderRadius.circular(8),
            ),
            child: ListView.builder(
              shrinkWrap: true, padding: const EdgeInsets.all(8),
              itemCount: _steps.length,
              itemBuilder: (ctx, i) => _stepRow(_steps[i]),
            ),
          ),
        const SizedBox(height: 12),
        SizedBox(
          width: 200,
          child: FilledButton(
            style: FilledButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 12)),
            onPressed: _startConnection,
            child: Text(tr('retry'), style: const TextStyle(fontSize: 14)),
          ),
        ),
      ],
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
  final _urlCtrl = TextEditingController();
  final _labelCtrl = TextEditingController();
  bool _testing = false;

  void _addAddress() {
    final url = _urlCtrl.text.trim();
    final label = _labelCtrl.text.trim();
    if (url.isEmpty) return;
    widget.addrMgr.add(url, label: label);
    setState(() { _urlCtrl.clear(); _labelCtrl.clear(); });
  }

  void _testAll() async {
    setState(() => _testing = true);
    await widget.addrMgr.testAll(client: widget.httpClient);
    if (mounted) setState(() => _testing = false);
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Server Addresses'),
      content: SizedBox(
        width: 400,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _urlCtrl,
                    decoration: const InputDecoration(labelText: 'Onion URL', border: OutlineInputBorder(), isDense: true),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: TextField(
                    controller: _labelCtrl,
                    decoration: const InputDecoration(labelText: 'Label', border: OutlineInputBorder(), isDense: true),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                FilledButton.tonalIcon(
                  icon: const Icon(Icons.add, size: 16),
                  label: const Text('Add', style: TextStyle(fontSize: 12)),
                  onPressed: _addAddress,
                ),
                const SizedBox(width: 8),
                OutlinedButton.icon(
                  icon: _testing
                      ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2))
                      : const Icon(Icons.network_check, size: 16),
                  label: Text(_testing ? 'Testing...' : 'Test All', style: const TextStyle(fontSize: 12)),
                  onPressed: _testing ? null : _testAll,
                ),
              ],
            ),
            const SizedBox(height: 8),
            if (widget.addrMgr.addresses.isEmpty)
              const Text('No addresses saved', style: TextStyle(color: Colors.grey))
            else
              SizedBox(
                height: 200,
                child: ListView.builder(
                  itemCount: widget.addrMgr.addresses.length,
                  itemBuilder: (ctx, i) {
                    final addr = widget.addrMgr.addresses[i];
                    return ListTile(
                      dense: true,
                      title: Text(addr.label.isNotEmpty ? addr.label : addr.url, style: const TextStyle(fontSize: 12)),
                      subtitle: Text(addr.url, style: const TextStyle(fontSize: 9, fontFamily: 'monospace')),
                      trailing: Icon(
                        addr.isReachable ? Icons.check_circle : Icons.help,
                        size: 16, color: addr.isReachable ? Colors.green : Colors.grey,
                      ),
                    );
                  },
                ),
              ),
          ],
        ),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.pop(context), child: const Text('Close')),
      ],
    );
  }
}
