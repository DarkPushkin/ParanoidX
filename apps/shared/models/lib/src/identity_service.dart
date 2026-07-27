/// SecureIdentityService provides high-level identity management operations.
/// 
/// Handles:
/// - Generating new BIP39 identities (24-word mnemonic)
/// - Importing existing mnemonics with optional passphrase
/// - Encrypting/decrypting mnemonics with device-bound keys
/// - Deriving Ed25519 and X25519 keys for signing/encryption
/// - Identity verification and export
library models.src.identity_service;

import 'dart:convert';
import 'dart:io';
import 'dart:math' show Random;
import 'dart:typed_data';

import 'package:bip39/bip39.dart' as bip39;
import 'package:bip32/bip32.dart' as bip32;
import 'package:crypto/crypto.dart';
import 'package:convert/convert.dart';
import 'package:ed25519_hd_key/ed25519_hd_key.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:pointycastle/pointycastle.dart';
import 'package:pointycastle/aead/chacha20_poly1305.dart';
import 'package:pointycastle/aead/aead.dart';

import 'identity.dart';

/// Service for managing sovereign identities.
class SecureIdentityService {
  static SecureIdentityService? _instance;
  SharedPreferences? _prefs;
  String _deviceKey = '';
  bool _initialized = false;

  SecureIdentityService._();

  /// Get singleton instance (initializes on first call)
  static Future<SecureIdentityService> get instance async {
    _instance ??= SecureIdentityService._();
    if (!_instance!._initialized) {
      await _instance!._init();
    }
    return _instance!;
  }

  /// Initialize the service (derive device key, open preferences)
  Future<void> _init() async {
    if (_initialized) return;
    
    _prefs = await SharedPreferences.getInstance();
    _deviceKey = await _deriveDeviceKey();
    _initialized = true;
  }

  /// Derive a device-bound encryption key from hardware identifiers
  Future<String> _deriveDeviceKey() async {
    final buffer = StringBuffer('simplex-identity-salt-v1');
    
    // Machine ID (Linux)
    try {
      final machineId = await File('/etc/machine-id').readAsString();
      buffer.write(machineId.trim());
    } catch (_) {}
    
    // Hostname
    try {
      final hostname = await Process.run('hostname', []);
      buffer.write(hostname.stdout.toString().trim());
    } catch (_) {}
    
    // User ID
    try {
      final uid = await Process.run('id', ['-u']);
      buffer.write(uid.stdout.toString().trim());
    } catch (_) {}
    
    // If all fails, use a constant (less secure but functional)
    if (buffer.length < 50) {
      buffer.write('fallback-device-id');
    }
    
    final hash = sha256.convert(utf8.encode(buffer.toString()));
    return hash.toString();
  }

  /// Generate a new BIP39 identity with 24-word mnemonic
  Future<Identity> generateIdentity({
    MnemonicWordCount wordCount = MnemonicWordCount.words24,
    String? label,
    String passphrase = '',
  }) async {
    await _init();
    
    // Generate entropy (32 bytes for 24 words = 256 bits)
    final entropyBytes = _generateEntropy(wordCount.entropyBits ~/ 8);
    final mnemonic = bip39.entropyToMnemonic(entropyBytes);
    
    // Derive seed from mnemonic + passphrase
    final seed = bip39.mnemonicToSeed(mnemonic, passphrase: passphrase);
    
    // Derive keys
    final keys = await _deriveKeys(seed, mnemonic, passphrase);
    
    // Encrypt mnemonic for storage
    final encryptedMnemonic = await _encryptMnemonic(mnemonic);
    
    // Generate identity ID from public key
    final id = _generateIdentityId(keys.identityKey.publicKey);
    
    // Create identity object
    final identity = Identity(
      id: id,
      encryptedMnemonic: encryptedMnemonic,
      derivationPath: DerivationPaths.identity,
      ed25519PubKey: keys.identityKey.publicKeyHex,
      x25519PubKey: keys.encryptionKey.publicKeyHex,
      createdAt: DateTime.now().millisecondsSinceEpoch,
      label: label,
      wordCount: wordCount,
    );
    
    // Store identity
    await _storeIdentity(identity);
    
    return identity;
  }

  /// Import identity from mnemonic phrase
  Future<Identity> importIdentity({
    required String mnemonic,
    String passphrase = '',
    String? label,
  }) async {
    await _init();
    
    // Validate mnemonic
    if (!bip39.validateMnemonic(mnemonic)) {
      throw ArgumentError('Invalid BIP39 mnemonic');
    }
    
    final wordCount = mnemonic.split(' ').length;
    final mnemonicWordCount = MnemonicWordCount.values.firstWhere(
      (e) => e.value == wordCount,
      orElse: () => MnemonicWordCount.words24,
    );
    
    // Derive seed
    final seed = bip39.mnemonicToSeed(mnemonic, passphrase: passphrase);
    
    // Derive keys
    final keys = await _deriveKeys(seed, mnemonic, passphrase);
    
    // Encrypt mnemonic for storage
    final encryptedMnemonic = await _encryptMnemonic(mnemonic);
    
    // Generate identity ID
    final id = _generateIdentityId(keys.identityKey.publicKey);
    
    // Create identity
    final identity = Identity(
      id: id,
      encryptedMnemonic: encryptedMnemonic,
      derivationPath: DerivationPaths.identity,
      ed25519PubKey: keys.identityKey.publicKeyHex,
      x25519PubKey: keys.encryptionKey.publicKeyHex,
      createdAt: DateTime.now().millisecondsSinceEpoch,
      label: label,
      wordCount: mnemonicWordCount,
    );
    
    // Store
    await _storeIdentity(identity);
    
    return identity;
  }

  /// Export mnemonic (requires identity verification)
  Future<String?> exportMnemonic({
    required String identityId,
    required String passphrase,
  }) async {
    await _init();
    
    final identity = await getIdentity(identityId);
    if (identity == null) return null;
    
    // Decrypt mnemonic
    final mnemonic = await _decryptMnemonic(identity.encryptedMnemonic);
    if (mnemonic == null) return null;
    
    // Verify with passphrase
    final seed = bip39.mnemonicToSeed(mnemonic, passphrase: passphrase);
    final ed25519Key = Ed25519HdKey.fromSeed(seed);
    final derived = ed25519Key.derive(DerivationPaths.identity);
    final pubKeyHex = hex.encode(derived.key);
    
    if (pubKeyHex != identity.ed25519PubKey) {
      throw ArgumentError('Invalid passphrase');
    }
    
    return mnemonic;
  }

  /// Get all stored identities
  Future<List<Identity>> getAllIdentities() async {
    await _init();
    
    final keys = _prefs!.getKeys().where((k) => k.startsWith('identity_')).toList();
    final identities = <Identity>[];
    
    for (final key in keys) {
      final jsonStr = _prefs!.getString(key);
      if (jsonStr != null) {
        try {
          final json = jsonDecode(jsonStr) as Map<String, dynamic>;
          identities.add(Identity.fromJson(json));
        } catch (_) {}
      }
    }
    
    // Sort by creation date (newest first)
    identities.sort((a, b) => b.createdAt.compareTo(a.createdAt));
    return identities;
  }

  /// Get identity by ID
  Future<Identity?> getIdentity(String identityId) async {
    await _init();
    
    final jsonStr = _prefs!.getString('identity_$identityId');
    if (jsonStr == null) return null;
    
    try {
      final json = jsonDecode(jsonStr) as Map<String, dynamic>;
      return Identity.fromJson(json);
    } catch (_) {
      return null;
    }
  }

  /// Delete identity
  Future<void> deleteIdentity(String identityId) async {
    await _init();
    await _prefs!.remove('identity_$identityId');
  }

  /// Set identity label
  Future<void> setIdentityLabel(String identityId, String label) async {
    await _init();
    
    final identity = await getIdentity(identityId);
    if (identity == null) throw ArgumentError('Identity not found');
    
    final updated = Identity(
      id: identity.id,
      encryptedMnemonic: identity.encryptedMnemonic,
      derivationPath: identity.derivationPath,
      ed25519PubKey: identity.ed25519PubKey,
      x25519PubKey: identity.x25519PubKey,
      createdAt: identity.createdAt,
      label: label,
      wordCount: identity.wordCount,
    );
    
    await _storeIdentity(updated);
  }

  /// Verify identity ownership (check mnemonic matches public key)
  Future<bool> verifyIdentity({
    required String identityId,
    required String mnemonic,
    String passphrase = '',
  }) async {
    await _init();
    
    final identity = await getIdentity(identityId);
    if (identity == null) return false;
    
    return identity.verifyMnemonic(mnemonic, passphrase: passphrase);
  }

  /// Derive signing key (Ed25519) for an identity
  Future<Ed25519HdKey> getSigningKey({
    required String identityId,
    required String mnemonic,
    String passphrase = '',
    String derivationPath = DerivationPaths.identity,
  }) async {
    final seed = bip39.mnemonicToSeed(mnemonic, passphrase: passphrase);
    return Ed25519HdKey.fromSeed(seed).derive(derivationPath);
  }

  /// Derive encryption key (X25519) for an identity
  Future<Bip32Key> getEncryptionKey({
    required String identityId,
    required String mnemonic,
    String passphrase = '',
    String derivationPath = DerivationPaths.encryption,
  }) async {
    // Note: BIP32 doesn't natively support X25519
    // This is a placeholder - use a proper X25519 derivation library
    // For now, derive Ed25519 and convert, or use a different approach
    final seed = bip39.mnemonicToSeed(mnemonic, passphrase: passphrase);
    final edKey = Ed25519HdKey.fromSeed(seed).derive(derivationPath);
    // Convert or derive X25519 from Ed25519 seed
    // This is a simplification - in production use x25519_dart or similar
    throw UnimplementedError('X25519 derivation needs dedicated library');
  }

  /// Sign a message with identity's Ed25519 key
  Future<Uint8List> signMessage({
    required String identityId,
    required String mnemonic,
    required Uint8List message,
    String passphrase = '',
    String derivationPath = DerivationPaths.identity,
  }) async {
    final key = await getSigningKey(
      identityId: identityId,
      mnemonic: mnemonic,
      passphrase: passphrase,
      derivationPath: derivationPath,
    );
    
    // Ed25519 signing
    return key.sign(message);
  }

  /// Verify a signature
  bool verifySignature({
    required String publicKeyHex,
    required Uint8List message,
    required Uint8List signature,
  }) {
    // Ed25519 verification
    final publicKey = hex.decode(publicKeyHex);
    return Ed25519HdKey.verify(message, signature, publicKey);
  }

  // ============ Private helpers ============

  /// Generate cryptographically secure entropy
  Uint8List _generateEntropy(int bytes) {
    final random = Random.secure();
    return Uint8List.fromList(List.generate(bytes, (_) => random.nextInt(256)));
  }

  /// Derive all keys from seed
  Future<DerivedKeys> _deriveKeys(Uint8List seed, String mnemonic, String passphrase) async {
    // Ed25519 identity key (BIP44 path: m/44'/1337'/0'/0/0)
    final ed25519Key = Ed25519HdKey.fromSeed(seed);
    final identityKey = ed25519Key.derive(DerivationPaths.identity);
    
    // For X25519, we use a separate path
    // Note: ed25519_hd_key derives Ed25519, for X25519 we need different handling
    // This is a placeholder - in production use x25519_dart with SLIP-10
    final encryptionKey = ed25519Key.derive(DerivationPaths.encryption);
    
    return DerivedKeys(
      seed: seed,
      identityKey: KeyPair(
        privateKey: identityKey.key,
        publicKey: identityKey.publicKey,
      ),
      encryptionKey: EncryptionKeyPair(
        privateKey: encryptionKey.key,
        publicKey: encryptionKey.publicKey,
      ),
      mnemonic: mnemonic,
      passphrase: passphrase,
    );
  }

  /// Generate identity ID from Ed25519 public key
  String _generateIdentityId(Uint8List publicKey) {
    final hash = sha256.convert(publicKey);
    return hash.toString();
  }

  /// Encrypt mnemonic with device-bound key (ChaCha20-Poly1305)
  Future<String> _encryptMnemonic(String mnemonic) async {
    final key = _getEncryptionKey();
    final nonce = _generateNonce();
    
    final cipher = ChaCha20Poly1305Engine();
    final params = AeadParameters(KeyParameter(key), 128, nonce);
    cipher.init(true, params);
    
    final plaintext = utf8.encode(mnemonic);
    final ciphertext = Uint8List(cipher.getOutputSize(plaintext.length));
    final len = cipher.processBytes(plaintext, 0, plaintext.length, ciphertext, 0);
    cipher.doFinal(ciphertext, len);
    
    // Combine nonce + ciphertext for storage
    final combined = Uint8List(nonce.length + ciphertext.length);
    combined.setAll(0, nonce);
    combined.setAll(nonce.length, ciphertext);
    
    return base64.encode(combined);
  }

  /// Decrypt mnemonic
  Future<String?> _decryptMnemonic(String encryptedB64) async {
    try {
      final combined = base64.decode(encryptedB64);
      if (combined.length < 12) return null; // nonce is 12 bytes
      
      final nonce = combined.sublist(0, 12);
      final ciphertext = combined.sublist(12);
      
      final key = _getEncryptionKey();
      final cipher = ChaCha20Poly1305Engine();
      final params = AeadParameters(KeyParameter(key), 128, nonce);
      cipher.init(false, params);
      
      final plaintext = Uint8List(cipher.getOutputSize(ciphertext.length));
      final len = cipher.processBytes(ciphertext, 0, ciphertext.length, plaintext, 0);
      cipher.doFinal(plaintext, len);
      
      return utf8.decode(plaintext.sublist(0, len));
    } catch (_) {
      return null;
    }
  }

  /// Get encryption key (32 bytes from device key)
  Uint8List _getEncryptionKey() {
    final hash = sha256.convert(utf8.encode(_deviceKey));
    return Uint8List.fromList(hash.bytes);
  }

  /// Generate 12-byte nonce for ChaCha20-Poly1305
  Uint8List _generateNonce() {
    final random = Random.secure();
    return Uint8List.fromList(List.generate(12, (_) => random.nextInt(256)));
  }

  /// Store identity in SharedPreferences
  Future<void> _storeIdentity(Identity identity) async {
    await _prefs!.setString('identity_${identity.id}', jsonEncode(identity.toJson()));
  }
}

/// Convenience function to get the service instance
Future<SecureIdentityService> getIdentityService() => SecureIdentityService.instance;