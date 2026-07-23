import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../theme.dart';
import '../main.dart';

/// GovernanceScreen manages governancescreen functionality.
class GovernanceScreen extends StatefulWidget {
  const GovernanceScreen({super.key});
  @override
  State<GovernanceScreen> createState() => _GovernanceScreenState();
}

class _GovernanceScreenState extends State<GovernanceScreen> {
  Map<String, dynamic>? _constitution;
  List<dynamic> _proposals = [];
  Timer? _timer;

  final _titleCtrl = TextEditingController();
  final _descCtrl = TextEditingController();

  @override
  void initState() { super.initState(); _refresh(); _timer = Timer.periodic(const Duration(seconds: 30), (_) => _refresh()); }
  @override
  void dispose() { _timer?.cancel(); _titleCtrl.dispose(); _descCtrl.dispose(); super.dispose(); }

  Future<void> _refresh() async {
    final api = context.read<AppState>().api;
    try {
      final r = await Future.wait([api.getConstitution(), api.getProposals()], eagerError: false);
      if (mounted) setState(() {
        _constitution = r[0] as Map<String, dynamic>?;
        final p = r[1] as Map<String, dynamic>?;
        _proposals = p?['proposals'] as List? ?? [];
      });
    } catch (_) {}
  }

  void _showProposalDialog() {
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: RoyalTheme.darkCard,
      title: const Text('New Proposal'),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        TextField(controller: _titleCtrl, decoration: const InputDecoration(labelText: 'Title')),
        const SizedBox(height: 12),
        TextField(controller: _descCtrl, maxLines: 3, decoration: const InputDecoration(labelText: 'Description')),
      ]),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')),
        ElevatedButton(onPressed: () async {
          if (_titleCtrl.text.trim().isEmpty) return;
          try {
            await context.read<AppState>().api.createProposal({
              'title': _titleCtrl.text, 'description': _descCtrl.text,
            });
            Navigator.pop(ctx);
            _titleCtrl.clear(); _descCtrl.clear();
            _refresh();
          } catch (e) {
            if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e')));
          }
        }, child: const Text('Create')),
      ],
    ));
  }

  @override
  Widget build(BuildContext context) {
    final articles = (_constitution?['articles'] as List?) ?? (_constitution?['constitution'] as List?) ?? [];
    return Scaffold(
      appBar: AppBar(
        title: Row(children: [
          Container(padding: const EdgeInsets.all(6), decoration: BoxDecoration(color: RoyalTheme.purple.withAlpha(30), borderRadius: BorderRadius.circular(8)),
            child: const Icon(Icons.gavel, color: RoyalTheme.purple, size: 20)),
          const SizedBox(width: 12),
          const Text('Governance'),
        ]),
      ),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        Container(
          padding: const EdgeInsets.all(16),
          decoration: RoyalTheme.glowBorder(color: RoyalTheme.purple),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text('Constitution', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 4),
            Text('Articles of the Sovereign Network', style: Theme.of(context).textTheme.bodyMedium),
            const SizedBox(height: 12),
            if (articles.isEmpty)
              Text('No articles defined', style: Theme.of(context).textTheme.bodyMedium)
            else
              ...articles.map((a) {
                final m = a is Map<String, dynamic> ? a : {'title': '$a'};
                return Padding(
                  padding: const EdgeInsets.symmetric(vertical: 6),
                  child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Container(width: 6, height: 6, margin: const EdgeInsets.only(top: 6),
                      decoration: const BoxDecoration(shape: BoxShape.circle, color: RoyalTheme.purple)),
                    const SizedBox(width: 10),
                    Expanded(child: Text('${m['title'] ?? m['article'] ?? '—'}',
                      style: const TextStyle(fontSize: 14, color: Colors.white))),
                  ]),
                );
              }),
          ]),
        ),
        const SizedBox(height: 16),
        Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
          Text('Proposals (${_proposals.length})', style: Theme.of(context).textTheme.titleMedium),
          ElevatedButton.icon(
            icon: const Icon(Icons.add, size: 18),
            label: const Text('New'),
            onPressed: _showProposalDialog,
          ),
        ]),
        const SizedBox(height: 12),
        if (_proposals.isEmpty)
          Container(
            padding: const EdgeInsets.all(32),
            decoration: RoyalTheme.gradientCard(),
            child: Center(
              child: Column(children: [
                Icon(Icons.inbox, color: RoyalTheme.gold.withAlpha(80), size: 40),
                const SizedBox(height: 8),
                Text('No proposals yet', style: Theme.of(context).textTheme.bodyMedium),
              ]),
            ),
          )
        else
          ..._proposals.map((p) {
            final m = p is Map<String, dynamic> ? p : {'title': '$p'};
            return Container(
              margin: const EdgeInsets.only(bottom: 8),
              padding: const EdgeInsets.all(14),
              decoration: RoyalTheme.gradientCard(),
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Text('${m['title'] ?? '—'}', style: const TextStyle(fontWeight: FontWeight.w600)),
                if (m['description'] != null) ...[
                  const SizedBox(height: 4),
                  Text('${m['description']}', style: Theme.of(context).textTheme.bodyMedium),
                ],
              ]),
            );
          }),
      ]),
    );
  }
}
