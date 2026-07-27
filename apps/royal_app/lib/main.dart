import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'theme.dart';
import 'services/royal_api_service.dart';
import 'screens/dashboard_screen.dart';
import 'screens/ai_office_screen.dart';
import 'screens/treasury_screen.dart';
import 'screens/communications_screen.dart';
import 'screens/dc_cloud_screen.dart';
import 'screens/governance_screen.dart';
import 'screens/system_screen.dart';
import 'screens/settings_screen.dart';
import 'screens/lock_screen.dart';

void main() {
  runApp(const RoyalApp());
}

/// RoyalApp manages the root MaterialApp widget for the Isle Royal client.
class RoyalApp extends StatelessWidget {
  const RoyalApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => AppState(),
      child: MaterialApp(
        title: 'Isle Royal',
        theme: RoyalTheme.dark,
        debugShowCheckedModeBanner: false,
        home: const AuthGate(),
      ),
    );
  }
}

/// AuthGate decides whether to show lock screen or main shell based on unlock state.
class AuthGate extends StatelessWidget {
  const AuthGate({super.key});

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    if (!state.unlocked) {
      return LockScreen(onUnlocked: () => state.setUnlocked(true));
    }
    return const RoyalShell();
  }
}

/// AppState manages application-wide state management for the Royal client.
class AppState extends ChangeNotifier {
  final RoyalApiService api = RoyalApiService();
  int selectedIndex = 0;
  bool isOffline = false;
  bool unlocked = false;
  Map<String, dynamic>? liveTelemetry;
  StreamSubscription? _sseSub;

  AppState() {
    _startSSE();
  }

  void _startSSE() {
    try {
      _sseSub?.cancel();
      _sseSub = api.sseEvents().listen((data) {
        liveTelemetry = data;
        notifyListeners();
      }, onError: (_) {});
    } catch (_) {}
  }

  void setIndex(int i) {
    selectedIndex = i;
    notifyListeners();
  }

  void setOffline(bool v) {
    isOffline = v;
    notifyListeners();
  }

  void setUnlocked(bool v) {
    unlocked = v;
    notifyListeners();
  }

  @override
  void dispose() {
    _sseSub?.cancel();
    api.dispose();
    super.dispose();
  }
}

/// RoyalShell manages the main shell with NavigationRail-based screen routing.
class RoyalShell extends StatefulWidget {
  const RoyalShell({super.key});

  @override
  State<RoyalShell> createState() => _RoyalShellState();
}

class _RoyalShellState extends State<RoyalShell> {
  final List<Widget> _screens = [
    const DashboardScreen(),
    const AIOfficeScreen(),
    const TreasuryScreen(),
    const CommunicationsScreen(),
    const DCCloudScreen(),
    const GovernanceScreen(),
    const SystemScreen(),
    const SettingsScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    final state = context.watch<AppState>();
    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth > 900;
        return Scaffold(
          body: Row(
            children: [
              if (isWide)
                NavigationRail(
                  selectedIndex: state.selectedIndex,
                  onDestinationSelected: state.setIndex,
                  labelType: NavigationRailLabelType.all,
                  backgroundColor: RoyalTheme.darkCard,
                  indicatorColor: RoyalTheme.gold.withAlpha(40),
                  leading: Padding(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    child: Center(
                      child: Container(
                        width: 48,
                        height: 48,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          border: Border.all(color: RoyalTheme.gold, width: 2),
                          boxShadow: [BoxShadow(color: RoyalTheme.gold.withAlpha(40), blurRadius: 8)],
                        ),
                        child: const Center(
                          child: Text('♚', style: TextStyle(fontSize: 24, color: RoyalTheme.gold)),
                        ),
                      ),
                    ),
                  ),
                  destinations: const [
                    NavigationRailDestination(icon: Icon(Icons.dashboard), label: Text('Dashboard')),
                    NavigationRailDestination(icon: Icon(Icons.auto_awesome), label: Text('AI Office')),
                    NavigationRailDestination(icon: Icon(Icons.account_balance), label: Text('Treasury')),
                    NavigationRailDestination(icon: Icon(Icons.chat), label: Text('Comms')),
                    NavigationRailDestination(icon: Icon(Icons.cloud), label: Text('DC Cloud')),
                    NavigationRailDestination(icon: Icon(Icons.gavel), label: Text('Governance')),
                    NavigationRailDestination(icon: Icon(Icons.monitor_heart), label: Text('System')),
                    NavigationRailDestination(icon: Icon(Icons.settings), label: Text('Settings')),
                  ],
                ),
              if (!isWide)
                NavigationBar(
                  selectedIndex: state.selectedIndex,
                  onDestinationSelected: state.setIndex,
                  backgroundColor: RoyalTheme.darkCard,
                  indicatorColor: RoyalTheme.gold.withAlpha(40),
                  destinations: const [
                    NavigationDestination(icon: Icon(Icons.dashboard), label: 'Dashboard'),
                    NavigationDestination(icon: Icon(Icons.auto_awesome), label: 'AI'),
                    NavigationDestination(icon: Icon(Icons.account_balance), label: 'Treasury'),
                    NavigationDestination(icon: Icon(Icons.chat), label: 'Chat'),
                    NavigationDestination(icon: Icon(Icons.cloud), label: 'DC'),
                    NavigationDestination(icon: Icon(Icons.gavel), label: 'Gov'),
                    NavigationDestination(icon: Icon(Icons.monitor_heart), label: 'System'),
                    NavigationDestination(icon: Icon(Icons.settings), label: 'Settings'),
                  ],
                ),
              const VerticalDivider(width: 1, thickness: 1),
              Expanded(
                child: AnimatedSwitcher(
                  duration: const Duration(milliseconds: 300),
                  child: _screens[state.selectedIndex],
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}