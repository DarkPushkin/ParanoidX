import 'dart:convert';
import 'dart:io';
import 'package:crypto/crypto.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Encrypted preferences using SHA-256 hashing for sensitive values.
/// PINs are stored as SHA-256(PIN + device_key) instead of plaintext.
/// SecurePrefs manages encrypted preferences using SHA-256 hashing for sensitive values.
class SecurePrefs {
  static SecurePrefs? _instance;
  String _deviceKey = '';
  bool _ready = false;

  SecurePrefs._();

  static Future<SecurePrefs> get instance async {
    if (_instance != null) return _instance!;
    _instance = SecurePrefs._();
    await _instance!._init();
    return _instance!;
  }

  Future<void> _init() async {
    final seed = StringBuffer('simplex-prefs-salt');
    try {
      seed.write(await File('/etc/machine-id').readAsString());
    } catch (_) {
      seed.write((await _runCmd('hostname')).trim());
    }
    try {
      seed.write((await _runCmd('id', ['-u'])).trim());
    } catch (_) {}
    _deviceKey = sha256.convert(utf8.encode(seed.toString())).toString();
    _ready = true;
  }

  Future<String> _runCmd(String cmd, [List<String>? args]) async {
    final proc = await Process.run(cmd, args ?? []);
    return proc.stdout as String;
  }

  /// Store a hashed value (one-way, for PIN verification)
  Future<void> setHashed(String key, String value) async {
    if (!_ready) await _init();
    final prefs = await SharedPreferences.getInstance();
    final hash = _hash(value);
    await prefs.setString('_sec_$key', hash);
  }

  /// Verify a value against stored hash
  Future<bool> verify(String key, String value) async {
    if (!_ready) await _init();
    final prefs = await SharedPreferences.getInstance();
    final stored = prefs.getString('_sec_$key');
    if (stored == null) return false;
    return stored == _hash(value);
  }

  /// Check if a key exists in secure storage
  Future<bool> containsKey(String key) async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.containsKey('_sec_$key');
  }

  /// Remove a key
  Future<void> remove(String key) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('_sec_$key');
  }

  /// Store an encrypted (reversible) value
  Future<void> setString(String key, String value) async {
    if (!_ready) await _init();
    final prefs = await SharedPreferences.getInstance();
    final encrypted = _xorEncrypt(value);
    await prefs.setString('_sec_str_$key', encrypted);
  }

  /// Read a reversibly encrypted value
  Future<String?> getString(String key) async {
    if (!_ready) await _init();
    final prefs = await SharedPreferences.getInstance();
    final encrypted = prefs.getString('_sec_str_$key');
    if (encrypted == null) return null;
    return _xorDecrypt(encrypted);
  }

  /// Clear all secure entries
  Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    final keys = prefs.getKeys().where((k) => k.startsWith('_sec_'));
    for (final k in keys) {
      await prefs.remove(k);
    }
  }

  String _hash(String value) {
    return sha256.convert(utf8.encode('$value:$_deviceKey')).toString();
  }

  String _xorEncrypt(String value) {
    final bytes = utf8.encode(value);
    final key = sha256.convert(utf8.encode(_deviceKey)).bytes;
    final result = <int>[];
    for (var i = 0; i < bytes.length; i++) {
      result.add(bytes[i] ^ key[i % key.length]);
    }
    return base64UrlEncode(result);
  }

  String _xorDecrypt(String encoded) {
    final bytes = base64Decode(encoded);
    final key = sha256.convert(utf8.encode(_deviceKey)).bytes;
    final result = <int>[];
    for (var i = 0; i < bytes.length; i++) {
      result.add(bytes[i] ^ key[i % key.length]);
    }
    return utf8.decode(result);
  }
}