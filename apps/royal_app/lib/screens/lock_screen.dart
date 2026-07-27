import 'package:flutter/material.dart';
import 'package:models/models.dart';
import 'package:provider/provider.dart';
import '../main.dart';
import '../theme.dart';

/// Lock screen for Royal App - requires PIN or biometric to unlock.
/// Uses the shared Identity model from models package.
class RoyalLockScreen extends StatefulWidget {
  final VoidCallback onUnlocked;
  final String? initialError;

  const RoyalLockScreen({
    super.key,
    required this.onUnlocked,
    this.initialError,
  });

  @override
  State<RoyalLockScreen> createState() => _RoyalLockScreenState();
}

class _RoyalLockScreenState extends State<RoyalLockScreen> with TickerProviderStateMixin {
  final _pinCtrl = TextEditingController();
  late AnimationController _shakeCtrl;
  late Animation<double> _shakeAnim;
  bool _obscurePin = true;
  int _pinAttempts = 0;
  bool _lockedOut = false;
  bool _biometricAvailable = false;

  @override
  void initState() {
    super.initState();
    _shakeCtrl = AnimationController(
      duration: const Duration(milliseconds: 400),
      vsync: this,
    );
    _shakeAnim = Tween<double>(begin: 0, end: 10).chain(CurveTween(curve: Curves.elasticIn)).animate(_shakeCtrl);
    _checkBiometric();
    if (widget.initialError != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(widget.initialError!), backgroundColor: RoyalTheme.red),
        );
      });
    }
  }

  Future<void> _checkBiometric() async {
    // TODO: Add local_auth dependency for biometric support
    // For now, just PIN
    setState(() => _biometricAvailable = false);
  }

  @override
  void dispose() {
    _pinCtrl.dispose();
    _shakeCtrl.dispose();
    super.dispose();
  }

  Future<void> _onUnlock() async {
    if (_lockedOut) return;

    final pin = _pinCtrl.text;
    if (pin.length != 6) {
      _shake();
      return;
    }

    final appState = context.read<AppState>();
    final prefs = await SecurePrefs.instance;

    // Check if we have stored identity
    final hasIdentity = await prefs.containsKey('identity_encrypted');
    final hasPin = await prefs.containsKey('pin_hash');

    if (!hasIdentity || !hasPin) {
      _shake();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: const Text('No identity configured. Run Isle app first.'), backgroundColor: RoyalTheme.red),
        );
      }
      return;
    }

    final encrypted = await prefs.getString('identity_encrypted');
    final pinHash = await prefs.getString('pin_hash');

    if (encrypted == null || pinHash == null) {
      _shake();
      return;
    }

    if (Identity.verifyPin(pin, pinHash)) {
      // Correct PIN - decrypt identity and initialize API
      try {
        final identity = Identity.decryptWithPin(encrypted, pin);
        
        // Initialize API with identity
        await appState.api.initializeWithIdentity(identity);
        
        _pinCtrl.clear();
        _pinAttempts = 0;
        widget.onUnlocked();
      } catch (e) {
        _shake();
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Decryption failed: $e'), backgroundColor: RoyalTheme.red),
          );
        }
      }
    } else {
      _pinAttempts++;
      _pinCtrl.clear();
      _shake();

      if (_pinAttempts >= 3) {
        _lockedOut = true;
        Future.delayed(const Duration(seconds: 30), () {
          if (mounted) setState(() { _lockedOut = false; _pinAttempts = 0; });
        });
      }

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(_lockedOut 
                ? 'Too many attempts. Locked for 30 seconds.' 
                : 'Wrong PIN. Attempts left: ${3 - _pinAttempts}'),
            backgroundColor: RoyalTheme.red,
          ),
        );
      }
    }
  }

  void _shake() {
    _shakeCtrl.forward(from: 0);
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    
    return Scaffold(
      backgroundColor: RoyalTheme.darkBg,
      body: AnimatedBuilder(
        animation: _shakeAnim,
        builder: (context, child) {
          return Transform.translate(
            offset: Offset(_shakeAnim.value, 0),
            child: child,
          );
        },
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 380),
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Crown icon
                  Container(
                    width: 80,
                    height: 80,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      border: Border.all(color: RoyalTheme.gold, width: 2),
                      boxShadow: [BoxShadow(color: RoyalTheme.gold.withAlpha(40), blurRadius: 16)],
                    ),
                    child: const Center(
                      child: Text('♚', style: TextStyle(fontSize: 40, color: RoyalTheme.gold)),
                    ),
                  ),
                  const SizedBox(height: 24),
                  Text(
                    'Isle Royal',
                    style: theme.textTheme.headlineMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    'Sovereign Admin Console',
                    style: theme.textTheme.bodyMedium?.copyWith(color: Colors.grey[400]),
                  ),
                  const SizedBox(height: 32),
                  
                  // PIN field
                  Container(
                    decoration: BoxDecoration(
                      color: RoyalTheme.darkCard,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: RoyalTheme.gold.withAlpha(60)),
                    ),
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                    child: Row(
                      children: [
                        Icon(Icons.lock, color: RoyalTheme.gold, size: 20),
                        const SizedBox(width: 12),
                        Expanded(
                          child: TextField(
                            controller: _pinCtrl,
                            obscureText: _obscurePin,
                            keyboardType: TextInputType.number,
                            maxLength: 6,
                            textAlign: TextAlign.center,
                            style: const TextStyle(
                              fontSize: 22,
                              letterSpacing: 8,
                              fontWeight: FontWeight.w500,
                              color: Colors.white,
                            ),
                            decoration: const InputDecoration(
                              border: InputBorder.none,
                              counterText: '',
                              hintText: '• • • • • •',
                              hintStyle: TextStyle(color: Colors.grey, letterSpacing: 8),
                              isDense: true,
                            ),
                            inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                            onSubmitted: (_) => _onUnlock(),
                            enabled: !_lockedOut,
                          ),
                        ),
                        IconButton(
                          icon: Icon(
                            _obscurePin ? Icons.visibility_off : Icons.visibility,
                            color: Colors.grey[500],
                            size: 20,
                          ),
                          onPressed: () => setState(() => _obscurePin = !_obscurePin),
                          padding: EdgeInsets.zero,
                          constraints: const BoxConstraints(),
                        ),
                      ],
                    ),
                  ),
                  
                  if (_lockedOut) ...[
                    const SizedBox(height: 12),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                      decoration: BoxDecoration(
                        color: RoyalTheme.red.withAlpha(30),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: RoyalTheme.red.withAlpha(80)),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.timer, color: RoyalTheme.red, size: 18),
                          const SizedBox(width: 8),
                          Text(
                            'Locked for 30 seconds',
                            style: TextStyle(color: RoyalTheme.red, fontWeight: FontWeight.w600),
                          ),
                        ],
                      ),
                    ),
                  ],
                  
                  const SizedBox(height: 24),
                  
                  // Unlock button
                  SizedBox(
                    width: double.infinity,
                    height: 52,
                    child: FilledButton(
                      style: FilledButton.styleFrom(
                        backgroundColor: RoyalTheme.gold,
                        foregroundColor: Colors.black,
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                        elevation: 4,
                        shadowColor: RoyalTheme.gold.withAlpha(60),
                      ),
                      onPressed: _lockedOut ? null : _onUnlock,
                      child: const Text(
                        'UNLOCK',
                        style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, letterSpacing: 1.5),
                      ),
                    ),
                  ),
                  
                  const SizedBox(height: 16),
                  
                  // Biometric placeholder
                  if (_biometricAvailable) ...[
                    TextButton.icon(
                      onPressed: () {}, // TODO: local_auth
                      icon: const Icon(Icons.fingerprint, size: 18),
                      label: const Text('Use Biometric', style: TextStyle(fontSize: 13)),
                      style: TextButton.styleFrom(foregroundColor: Colors.grey[400]),
                    ),
                    const SizedBox(height: 8),
                  ],
                  
                  // Hint
                  Text(
                    'PIN protects your sovereign identity.\nRun Isle app to create/import identity first.',
                    style: theme.textTheme.bodySmall?.copyWith(color: Colors.grey[600]),
                    textAlign: TextAlign.center,
                  ),
                  
                  const SizedBox(height: 24),
                  
                  // Version info
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    decoration: BoxDecoration(
                      color: Colors.black.withAlpha(80),
                      borderRadius: BorderRadius.circular(6),
                      border: Border.all(color: Colors.grey[800]!),
                    ),
                    child: Text(
                      'Royal App v2.0.0 • Saint Mary Liberty Island',
                      style: TextStyle(fontSize: 10, color: Colors.grey[500], fontFamily: 'monospace'),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}