import 'dart:async';
import 'package:flutter/material.dart';
import 'package:widgets/widgets.dart';
import '../services/isle_api_service.dart';
import '../widgets/nt_display.dart';

/// PosScreen manages the point-of-sale terminal for creating and paying invoices.
class PosScreen extends StatefulWidget {
  final IsleApiService api;
  const PosScreen({super.key, required this.api});

  @override
  State<PosScreen> createState() => _PosScreenState();
}

class _PosScreenState extends State<PosScreen> {
  final _amountCtrl = TextEditingController();
  final _descCtrl = TextEditingController();
  Map<String, dynamic>? _lastInvoice;
  Map<String, dynamic>? _stats;
  List<dynamic> _invoices = [];
  bool _loading = false;
  Timer? _pollTimer;

  @override
  void initState() {
    super.initState();
    _loadStats();
  }

  Future<void> _loadStats() async {
    try {
      final s = await widget.api.pos.stats();
      if (mounted) setState(() => _stats = s);
    } catch (_) {}
  }

  Future<void> _createInvoice() async {
    final amount = int.tryParse(_amountCtrl.text.trim());
    if (amount == null || amount <= 0) return;
    setState(() => _loading = true);
    try {
      final inv = await widget.api.createPosInvoice(
        amount,
        description: _descCtrl.text.trim().isEmpty ? null : _descCtrl.text.trim(),
      );
      setState(() {
        _lastInvoice = inv;
        _loading = false;
      });
      _startPolling(inv['id']);
      _amountCtrl.clear();
      _descCtrl.clear();
      _loadStats();
    } catch (e) {
      setState(() => _loading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e')),
        );
      }
    }
  }

  void _startPolling(String invoiceId) {
    _pollTimer?.cancel();
    _pollTimer = Timer.periodic(const Duration(seconds: 3), (_) async {
      try {
        final inv = await widget.api.checkPosInvoice(invoiceId);
        if (inv['status'] == 'paid') {
          _pollTimer?.cancel();
          if (mounted) {
            setState(() => _lastInvoice = inv);
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('Invoice paid!'), backgroundColor: Colors.green),
            );
            _loadStats();
          }
        }
      } catch (_) {}
    });
  }

  @override
  void dispose() {
    _pollTimer?.cancel();
    _amountCtrl.dispose();
    _descCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('POS Terminal')),
      body: ListView(
        padding: const EdgeInsets.all(12),
        children: [
          if (_stats != null) _buildStatsCard(),
          const SizedBox(height: 10),
          _buildCreateCard(),
          if (_lastInvoice != null) ...[
            const SizedBox(height: 10),
            _buildInvoiceCard(_lastInvoice!),
          ],
          const SizedBox(height: 10),
          _buildHistoryCard(),
        ],
      ),
    );
  }

  Widget _buildStatsCard() {
    return SectionCard(
      title: 'POS Stats',
      icon: Icons.point_of_sale,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Total: ${_stats!['total_invoices'] ?? 0} invoices'),
          Text('Paid: ${_stats!['paid'] ?? 0}'),
          Text('Pending: ${_stats!['pending'] ?? 0}'),
          Text('Volume: ${_stats!['total_volume_ng'] ?? 0} NT'),
          Text('Fee: ${_stats!['fee_bps'] ?? 100} bps'),
        ],
      ),
    );
  }

  Widget _buildCreateCard() {
    return SectionCard(
      title: 'New Invoice',
      icon: Icons.add_shopping_cart,
      child: Column(
        children: [
          TextField(
            controller: _amountCtrl,
            decoration: const InputDecoration(
              labelText: 'Amount (NT)',
              border: OutlineInputBorder(),
            ),
            keyboardType: TextInputType.number,
          ),
          const SizedBox(height: 8),
          TextField(
            controller: _descCtrl,
            decoration: const InputDecoration(
              labelText: 'Description (optional)',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: _loading ? null : _createInvoice,
            icon: _loading
                ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.qr_code),
            label: Text(_loading ? 'Creating...' : 'Create Invoice'),
          ),
        ],
      ),
    );
  }

  Widget _buildInvoiceCard(Map<String, dynamic> inv) {
    final status = inv['status'] ?? 'unknown';
    final isPaid = status == 'paid';
    final isExpired = status == 'expired' || status != 'paid' && _isExpired(inv['expires_at']);
    final qrUrl = widget.api.getQrUrl(inv['id']);

    return SectionCard(
      title: 'Invoice ${inv['id'].toString().substring(0, 8)}...',
      icon: isPaid ? Icons.check_circle : (isExpired ? Icons.timer_off : Icons.hourglass_empty),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          NtDisplay(amountNt: inv['amount_ng'] ?? 0),
          if (inv['description'] != null && (inv['description'] as String).isNotEmpty)
            Text('${inv['description']}', style: Theme.of(context).textTheme.bodySmall),
          const SizedBox(height: 4),
          Text('Status: $status',
              style: TextStyle(
                  color: isPaid ? Colors.green : (isExpired ? Colors.red : Colors.orange),
                  fontWeight: FontWeight.bold)),
          Text('Created: ${inv['created_at'] ?? '?'}'),
          Text('Expires: ${inv['expires_at'] ?? '?'}'),
          if (inv['commission_ng'] != null)
            Text('Fee: ${inv['commission_ng']} NT (${inv['net_amount_ng']} NT net)'),
          if (isPaid) ...[
            const SizedBox(height: 8),
            Text('Paid by: ${inv['payer'] ?? '?'}', style: const TextStyle(color: Colors.green)),
          ],
          if (!isPaid && !isExpired) ...[
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton.icon(
                onPressed: () => _payInvoice(inv['id']),
                icon: const Icon(Icons.payment),
                label: const Text('Pay this invoice'),
              ),
            ),
          ],
        ],
      ),
    );
  }

  bool _isExpired(String? expiresAt) {
    if (expiresAt == null) return false;
    try {
      final exp = DateTime.parse(expiresAt);
      return exp.isBefore(DateTime.now());
    } catch (_) {
      return false;
    }
  }

  Future<void> _payInvoice(String invoiceId) async {
    try {
      final result = await widget.api.payPosInvoice(invoiceId);
      if (result['status'] == 'paid') {
        _pollTimer?.cancel();
        setState(() => _lastInvoice = result);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Payment successful!'), backgroundColor: Colors.green),
          );
        }
        _loadStats();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Payment error: $e')),
        );
      }
    }
  }

  Widget _buildHistoryCard() {
    return SectionCard(
      title: 'My Invoices',
      icon: Icons.history,
      child: FutureBuilder<Map<String, dynamic>>(
        future: widget.api.listPosInvoices(limit: 10),
        builder: (ctx, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snapshot.hasError) {
            return Text('Error: ${snapshot.error}');
          }
          final invoices = snapshot.data?['invoices'] as List<dynamic>? ?? [];
          if (invoices.isEmpty) return const Text('No invoices yet');
          return Column(
            children: invoices.take(5).map((inv) {
              final isPaid = inv['status'] == 'paid';
              return ListTile(
                dense: true,
                leading: Icon(isPaid ? Icons.check_circle : Icons.hourglass_empty,
                    color: isPaid ? Colors.green : Colors.orange, size: 20),
                title: Text('${inv['amount_ng']} NT — ${inv['status']}',
                    style: const TextStyle(fontSize: 13)),
                subtitle: Text(inv['id'] ?? '', style: const TextStyle(fontSize: 10)),
              );
            }).toList(),
          );
        },
      ),
    );
  }
}
