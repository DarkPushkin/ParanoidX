import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:http/http.dart' as http;
import 'package:file_picker/file_picker.dart';

import 'paranoidx_screen.dart';
import '../models/chat_message.dart';
import '../widgets/chat_message_bubble.dart';
import '../widgets/chat_input_bar.dart';
import '../widgets/chat_contact_list.dart';
import '../widgets/chat_encryption_indicator.dart';
import '../widgets/pulse_dot.dart';

final _timeFmt = DateFormat('HH:mm');
final _dateFmt = DateFormat('yyyy-MM-dd HH:mm');

class _InvoiceData {
  final String id;
  final double amount;
  final String currency;
  final String description;
  final String status;
  _InvoiceData({required this.id, required this.amount, required this.currency, required this.description, required this.status});
  factory _InvoiceData.fromJson(Map<String, dynamic> json) {
    return _InvoiceData(
      id: json['id'] as String? ?? '',
      amount: (json['amount'] as num?)?.toDouble() ?? 0,
      currency: json['currency'] as String? ?? 'XAG',
      description: json['description'] as String? ?? '',
      status: json['status'] as String? ?? '',
    );
  }
}

class _TemplateData {
  final String name; final String text;
  _TemplateData({required this.name, required this.text});
  factory _TemplateData.fromJson(Map<String, dynamic> json) {
    return _TemplateData(name: json['name'] as String? ?? '', text: json['text'] as String? ?? '');
  }
}

class _AutoReplyRule {
  final String keyword;
  final String response;
  _AutoReplyRule({required this.keyword, required this.response});
  factory _AutoReplyRule.fromJson(Map<String, dynamic> json) {
    return _AutoReplyRule(keyword: json['keyword'] as String? ?? '', response: json['response'] as String? ?? '');
  }
  Map<String, dynamic> toJson() => {'keyword': keyword, 'response': response};
}

/// SimplexChatScreen manages the full SimpleX chat interface with contacts, messages, invoices and settings.
class SimplexChatScreen extends StatefulWidget {
  final String serverUrl;
  final http.Client httpClient;
  const SimplexChatScreen({super.key, required this.serverUrl, required this.httpClient});
  @override
  State<SimplexChatScreen> createState() => _SimplexChatScreenState();
}

class _SimplexChatScreenState extends State<SimplexChatScreen> {
  List<ChatMessage> _messages = [];
  List<ContactInfo> _contacts = [];
  final _inputCtrl = TextEditingController();
  final _connectCtrl = TextEditingController();
  final _searchCtrl = TextEditingController();
  final _invAmountCtrl = TextEditingController();
  final _invDescCtrl = TextEditingController();
  final _scrollCtrl = ScrollController();
  List<_TemplateData> _templates = [];
  bool _loading = true;
  bool _setupDone = false;
  String _selectedContactId = '';
  String _contactLink = '';
  Uint8List? _qrImage;
  http.Client? _sseClient;
  http.StreamedResponse? _sseResponse;
  Timer? _pollTimer;
  int _selectedTab = 0;
  int _unreadCount = 0;
  final Map<String, int> _perContactUnread = {};
  bool _searching = false;
  String _bridgeStatus = 'unknown';
  int _totalMessages = 0;
  String _serverStatusText = '';
  bool _showTemplatePicker = false;
  int _sseReconnectAttempt = 0;
  bool _reconnecting = false;
  // ParanoidX multi-layer proxy chain status
  bool _pxV2Ray = false;
  bool _pxVPN = false;
  bool _pxTor = false;
  bool _pxSimplex = false;
  DateTime? _scheduledTime;
  String? _replyingTo;
  String? _replyingText;
  String? _typingContact;
  Timer? _typingTimer;
  List<ContactGroup> _contactGroups = [];
  List<_AutoReplyRule> _autoReplyRules = [];
  String _selectedGroup = '';

  Timer? _pxPollTimer;
  String _chatThemeColor = 'default';
  String _stewardThinking = '';
  double _walletBalance = 0;
  String _activeLanguage = 'en';

  // Cycle 21: Contact trust
  Map<String, String> _contactTrust = {}; // chatID → trust level
  // Cycle 23: Media gallery
  bool _mediaGalleryOpen = false;
  // Cycle 25: Link preview
  final Set<String> _urlsPreviewed = {};
  // Cycle 26: Notification sounds
  Map<String, String> _contactSounds = {};
  // Cycle 27: AI suggestions
  List<String> _suggestions = [];
  bool _showSuggestions = false;
  // Cycle 28: Bulk operations
  bool _bulkMode = false;
  final Set<String> _selectedBulkIds = {};
  // Cycle 29: Contact statuses
  Map<String, String> _contactStatuses = {}; // chatID → status
  Map<String, String> _contactLastSeen = {};

  @override
  void initState() {
    super.initState();
    _checkSetup();
    _loadEverything();
    _pxPollTimer = Timer.periodic(const Duration(seconds: 15), (_) => _loadParanoidxStatus());
  }

  @override
  void dispose() {
    _inputCtrl.dispose();
    _connectCtrl.dispose();
    _searchCtrl.dispose();
    _invAmountCtrl.dispose();
    _invDescCtrl.dispose();
    _scrollCtrl.dispose();
    _pollTimer?.cancel();
    _sseResponse?.stream.listen((_) {});
    _sseResponse?.stream.drain();
    _sseClient?.close();
    super.dispose();
  }

/// Returns the current  apiBase value.
  String get _apiBase => widget.serverUrl;

  Future<void> _checkSetup() async {
    try {
      final resp = await widget.httpClient
          .get(Uri.parse('$_apiBase/api/transport/info'))
          .timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200 && mounted) setState(() => _setupDone = true);
    } catch (_) {
      if (mounted) setState(() => _setupDone = false);
    }
  }

  Future<void> _loadEverything() async {
    _loadHistory();
    _loadAddress();
    _loadContacts();
    _loadInvoices();
    _loadInvoiceStats();
    _loadServerStatus();
    _loadTemplates();
    _loadContactGroups();
    _loadAutoReplyRules();
    _loadParanoidxStatus();
    _loadTrust();
    _loadSounds();
    _loadContactStatuses();
    _loadSuggestions();
    _connectSSE();
  }

  Future<void> _loadParanoidxStatus() async {
    try {
      final resp = await widget.httpClient
          .get(Uri.parse('$_apiBase/api/paranoidx/status'))
          .timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200 && mounted) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        final layers = data['layers'] as List<dynamic>? ?? [];
        for (final l in layers) {
          final layer = (l as Map<String, dynamic>)['layer'] as String?;
          final healthy = (l as Map<String, dynamic>)['healthy'] as bool? ?? false;
          if (mounted) setState(() {
            if (layer == 'v2ray') _pxV2Ray = healthy;
            if (layer == 'vpn') _pxVPN = healthy;
            if (layer == 'tor') _pxTor = healthy;
            if (layer == 'simplex') _pxSimplex = healthy;
          });
        }
      }
    } catch (_) {}
  }

  Future<void> _loadContactGroups() async {
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/chat/groups')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['groups'] is List) {
          final list = (data['groups'] as List).map((e) => ContactGroup.fromJson(e as Map<String, dynamic>)).toList();
          if (mounted) setState(() => _contactGroups = list);
        }
      }
    } catch (_) {}
  }

  Future<void> _loadAutoReplyRules() async {
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/chat/auto-reply')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['rules'] is List) {
          final list = (data['rules'] as List).map((e) => _AutoReplyRule.fromJson(e as Map<String, dynamic>)).toList();
          if (mounted) setState(() => _autoReplyRules = list);
        }
      }
    } catch (_) {}
  }

  // ============ Cycles 21-30 new features ============

  Future<void> _loadTrust() async {
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/chat/trust')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['trust'] is Map) {
          if (mounted) setState(() => _contactTrust = (data['trust'] as Map).map((k, v) => MapEntry(k as String, v as String)));
        }
      }
    } catch (_) {}
  }

  Future<void> _updateTrust(String chatId, String level) async {
    try {
      await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/trust'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'chat_id': chatId, 'level': level}),
      ).timeout(const Duration(seconds: 5));
      await _loadTrust();
    } catch (_) {}
  }

  Future<void> _loadSounds() async {
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/chat/sound')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['sounds'] is Map) {
          if (mounted) setState(() => _contactSounds = (data['sounds'] as Map).map((k, v) => MapEntry(k as String, v as String)));
        }
      }
    } catch (_) {}
  }

  Future<void> _updateSound(String chatId, String sound) async {
    try {
      await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/sound'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'chat_id': chatId, 'sound': sound}),
      ).timeout(const Duration(seconds: 5));
      await _loadSounds();
    } catch (_) {}
  }

  Future<void> _loadContactStatuses() async {
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/chat/contact/status')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true) {
          if (data['statuses'] is Map) {
            if (mounted) setState(() => _contactStatuses = (data['statuses'] as Map).map((k, v) => MapEntry(k as String, v as String)));
          }
        }
      }
    } catch (_) {}
  }

  Future<void> _loadSuggestions() async {
    if (_selectedContactId.isEmpty) return;
    try {
      final cid = _selectedContactNumber;
      final resp = await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/suggest'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'chat_id': _selectedContactId, 'contact_id': cid}),
      ).timeout(const Duration(seconds: 15));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['suggestions'] is List) {
          if (mounted) setState(() {
            _suggestions = (data['suggestions'] as List).map((e) => e as String).toList();
            _showSuggestions = _suggestions.isNotEmpty;
          });
        }
      }
    } catch (_) {}
  }

  void _sendSuggestion(String text) {
    _inputCtrl.text = text;
    _showSuggestions = false;
    _sendMessage();
  }

  Future<void> _translateText(String text) async {
    final lang = {'en': 'Russian', 'ru': 'English', 'es': 'English'}[_activeLanguage] ?? 'English';
    final targetLang = _activeLanguage == 'en' ? 'ru' : (_activeLanguage == 'ru' ? 'en' : 'en');
    try {
      final resp = await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/translate'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'text': text, 'target_lang': targetLang, 'source_lang': _activeLanguage}),
      ).timeout(const Duration(seconds: 15));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['translated'] != null) {
          if (mounted) ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Translation: ${data['translated']}'), backgroundColor: Colors.blue));
        }
      }
    } catch (_) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Translation failed'), backgroundColor: Colors.red));
    }
  }

  void _showMediaGallery() {
    showDialog(context: context, builder: (ctx) => StatefulBuilder(builder: (ctx, setDlgState) {
      List<ChatMessage> mediaItems = [];
      var loading = true;
      _fetchMedia(ctx).then((items) { mediaItems = items; loading = false; setDlgState(() {}); });
      return AlertDialog(
        backgroundColor: Colors.grey[900],
        title: Row(children: [
          const Text('Media Gallery', style: TextStyle(fontSize: 20)),
          const Spacer(),
          if (loading) const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)),
        ]),
        content: SizedBox(width: 400, height: 400, child: loading
          ? const Center(child: CircularProgressIndicator())
          : mediaItems.isEmpty
            ? Center(child: Text('No media', style: TextStyle(fontSize: 16, color: Colors.grey[500])))
            : GridView.builder(
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(crossAxisCount: 3, crossAxisSpacing: 4, mainAxisSpacing: 4),
                itemCount: mediaItems.length,
                itemBuilder: (_, i) {
                  final m = mediaItems[i];
                  return GestureDetector(
                    onTap: () => _showMediaDetail(m),
                    child: Container(
                      decoration: BoxDecoration(
                        color: Colors.grey.shade800,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
                        Icon(m.voiceUrl != null ? Icons.mic : Icons.attach_file, size: 32, color: Colors.grey[400]),
                        Text(m.voiceUrl != null ? '🎤 ${m.voiceDuration ?? 0}s' : (m.fileName ?? 'file'), style: const TextStyle(fontSize: 12), maxLines: 1, overflow: TextOverflow.ellipsis),
                      ]),
                    ),
                  );
                },
              ),
        ),
        actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
      );
    }));
  }

  Future<List<ChatMessage>> _fetchMedia(BuildContext ctx) async {
    try {
      var url = '$_apiBase/api/chat/media';
      if (_selectedContactId.isNotEmpty) url += '?chat_id=$_selectedContactId';
      final resp = await widget.httpClient.get(Uri.parse(url)).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['media'] is List) {
          return (data['media'] as List).map((m) => ChatMessage.fromJson(m as Map<String, dynamic>)).toList();
        }
      }
    } catch (_) {}
    return [];
  }

  void _showMediaDetail(ChatMessage msg) {
    final isImage = msg.fileName != null && RegExp(r'\.(jpg|jpeg|png|gif|webp|bmp|svg)$', caseSensitive: false).hasMatch(msg.fileName!);
    if (isImage && msg.fileUrl != null) {
      // Fullscreen image viewer with pinch-to-zoom
      Navigator.push(context, MaterialPageRoute(builder: (_) => Scaffold(
        backgroundColor: Colors.black,
        appBar: AppBar(backgroundColor: Colors.black87, title: Text(msg.fileName ?? 'Image', style: const TextStyle(fontSize: 18)),
          actions: [IconButton(icon: const Icon(Icons.close, size: 28), onPressed: () => Navigator.pop(context))]),
        body: Center(
          child: InteractiveViewer(
            minScale: 0.5, maxScale: 4.0,
            child: msg.fileUrl!.startsWith('http')
                ? Image.network(msg.fileUrl!, fit: BoxFit.contain, loadingBuilder: (_, child, progress) =>
                    progress == null ? child : const Center(child: CircularProgressIndicator()),
                    errorBuilder: (_, __, ___) => Icon(Icons.broken_image, size: 64, color: Colors.grey[500]))
                : Image.file(File(msg.fileUrl!), fit: BoxFit.contain, errorBuilder: (_, __, ___) =>
                    Icon(Icons.broken_image, size: 64, color: Colors.grey[500])),
          ),
        ),
      )));
      return;
    }
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: Text(msg.voiceUrl != null ? 'Voice Message' : msg.fileName ?? 'Media', style: const TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        Icon(msg.voiceUrl != null ? Icons.play_circle : isImage ? Icons.image : Icons.attach_file, size: 64, color: Colors.grey[400]),
        const SizedBox(height: 12),
        Text(msg.voiceUrl != null ? 'Duration: ${msg.voiceDuration ?? 0}s' : 'File: ${msg.fileName ?? "unknown"}', style: const TextStyle(fontSize: 18)),
        if (msg.fileSize != null && msg.fileSize! > 0) Text('Size: ${(msg.fileSize! / 1024).toStringAsFixed(1)} KB', style: TextStyle(fontSize: 16, color: Colors.grey[500])),
        Text(msg.timestamp, style: TextStyle(fontSize: 14, color: Colors.grey[600])),
      ]),
      actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
    ));
  }

  void _showSoundDialog() {
    final sounds = ['default', 'chime', 'bell', 'alert', 'soft', 'none'];
    final current = _contactSounds[_selectedContactId] ?? 'default';
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: const Text('Notification Sound', style: TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, children: sounds.map((s) => ListTile(
        leading: Icon(Icons.check, size: 22, color: current == s ? Colors.cyan : Colors.transparent),
        title: Text(s.capitalizeFirst(), style: TextStyle(fontSize: 18, color: Colors.white)),
        onTap: () { Navigator.pop(ctx); _updateSound(_selectedContactId, s); },
      )).toList()),
    ));
  }

  void _showBulkModeToggle() {
    setState(() {
      _bulkMode = !_bulkMode;
      if (!_bulkMode) _selectedBulkIds.clear();
    });
    if (_bulkMode && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: const Text('Select messages to bulk delete or forward'),
          backgroundColor: Colors.amber,
          action: SnackBarAction(label: 'Delete', textColor: Colors.white, onPressed: () => _showBulkDeleteConfirm()),
        ),
      );
    }
  }

  void _showBulkDeleteConfirm() {
    if (_selectedBulkIds.isEmpty) { if (mounted) ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('No messages selected'), backgroundColor: Colors.red)); return; }
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: const Text('Bulk Delete', style: TextStyle(fontSize: 20)),
      content: Text('Delete ${_selectedBulkIds.length} selected messages?', style: const TextStyle(fontSize: 18)),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
        ElevatedButton(onPressed: () async {
          Navigator.pop(ctx);
          try {
            await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/bulk-delete'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'ids': _selectedBulkIds.toList()}),
            ).timeout(const Duration(seconds: 10));
            setState(() { _bulkMode = false; _selectedBulkIds.clear(); });
            _loadHistory();
          } catch (_) {}
        }, child: const Text('Delete', style: TextStyle(fontSize: 18, color: Colors.red))),
      ],
    ));
  }

  void _showAuditLog() async {
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/admin/audit-log/enhanced')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200 && mounted) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['events'] is List) {
          final events = (data['events'] as List).map((e) => e as Map<String, dynamic>).toList();
          showDialog(context: context, builder: (ctx) => AlertDialog(
            backgroundColor: Colors.grey[900],
            title: Row(children: [
              Icon(Icons.security, size: 22, color: Colors.cyan),
              const SizedBox(width: 8),
              const Text('Security Audit Log', style: TextStyle(fontSize: 20)),
            ]),
            content: SizedBox(width: 500, height: 400, child: events.isEmpty
              ? Center(child: Text('No events', style: TextStyle(fontSize: 16, color: Colors.grey[500])))
              : ListView.builder(
                  itemCount: events.length,
                  itemBuilder: (_, i) {
                    final e = events[i];
                    final sev = (e['severity'] as String?) ?? 'info';
                    final sevColor = sev == 'critical' ? Colors.red : (sev == 'warning' ? Colors.orange : Colors.grey);
                    return ListTile(
                      dense: true,
                      leading: Icon(Icons.circle, size: 12, color: sevColor),
                      title: Text('${e['event'] ?? ""}', style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
                      subtitle: Text('${e['detail'] ?? ""}  •  ${e['timestamp'] ?? ""}', style: TextStyle(fontSize: 12, color: Colors.grey[500])),
                    );
                  },
                ),
            ),
            actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
          ));
        }
      }
    } catch (_) {}
  }

  Widget _trustBadge(String? trustLevel) {
    if (trustLevel == null || trustLevel == 'none') return const SizedBox.shrink();
    final color = trustLevel == 'verified' ? Colors.blue : (trustLevel == 'trusted' ? Colors.green : Colors.red);
    final icon = trustLevel == 'verified' ? Icons.verified : (trustLevel == 'trusted' ? Icons.shield : Icons.block);
    return Padding(padding: const EdgeInsets.only(right: 4),
      child: Tooltip(message: trustLevel, child: Icon(icon, size: 16, color: color)));
  }

  Widget _statusIndicator(String? status) {
    if (status == null || status == 'offline') return const SizedBox.shrink();
    final color = status == 'online' ? Colors.green : Colors.orange;
    return Padding(padding: const EdgeInsets.only(right: 4),
      child: Container(width: 10, height: 10, decoration: BoxDecoration(shape: BoxShape.circle, color: color)));
  }

  void _sendTypingIndicator() {
    if (_selectedContactId.isEmpty) return;
    widget.httpClient.post(Uri.parse('$_apiBase/api/chat/typing'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'chat_id': _selectedContactId, 'typing': true}),
    ).timeout(const Duration(seconds: 3)).catchError((_) {});
  }

  Future<void> _loadTemplates() async {
    try {
      final resp = await widget.httpClient
          .get(Uri.parse('$_apiBase/api/chat/templates'))
          .timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['templates'] != null) {
          final list = data['templates'] as List;
          if (mounted) setState(() {
            _templates = list.map((e) => _TemplateData.fromJson(e as Map<String, dynamic>)).toList();
          });
        }
      }
    } catch (_) {}
  }

  void _saveDraft() {
    final text = _inputCtrl.text.trim();
    widget.httpClient.post(
      Uri.parse('$_apiBase/api/chat/drafts'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'chat_id': _selectedContactId, 'text': text}),
    ).timeout(const Duration(seconds: 3)).catchError((_) {});
  }

  void _showScheduleDialog() {
    showDatePicker(
      context: context, initialDate: DateTime.now().add(const Duration(hours: 1)),
      firstDate: DateTime.now(), lastDate: DateTime.now().add(const Duration(days: 30)),
      builder: (ctx, child) => Theme(data: Theme.of(context).copyWith(dialogBackgroundColor: Colors.grey[900]), child: child!),
    ).then((date) {
      if (date == null) return;
      showTimePicker(
        context: context, initialTime: TimeOfDay.fromDateTime(DateTime.now().add(const Duration(hours: 1))),
        builder: (ctx, child) => Theme(data: Theme.of(context).copyWith(dialogBackgroundColor: Colors.grey[900]), child: child!),
      ).then((time) {
        if (time == null) return;
        setState(() => _scheduledTime = DateTime(date.year, date.month, date.day, time.hour, time.minute));
      });
    });
  }

  Future<void> _loadServerStatus() async {
    try {
      final resp = await widget.httpClient
          .get(Uri.parse('$_apiBase/api/chat/server-status'))
          .timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (mounted) setState(() {
          _bridgeStatus = data['bridge'] as String? ?? 'unknown';
          _serverStatusText = data['status_text'] as String? ?? '';
        });
      }
    } catch (_) {}
    try {
      final hresp = await widget.httpClient
          .get(Uri.parse('$_apiBase/api/health'))
          .timeout(const Duration(seconds: 5));
      if (hresp.statusCode == 200) {
        final data = jsonDecode(hresp.body) as Map<String, dynamic>;
        if (mounted) setState(() { _totalMessages = data['total_messages'] as int? ?? 0; });
      }
    } catch (_) {}
  }

  List<_InvoiceData> _invoices = [];
  int _totalInvoiceCount = 0;
  int _pendingInvoiceCount = 0;
  int _paidInvoiceCount = 0;

  Future<void> _loadInvoices() async {
    try {
      final resp = await widget.httpClient
          .get(Uri.parse('$_apiBase/api/chat/invoice/list'))
          .timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['invoices'] is List) {
          final list = (data['invoices'] as List).map((e) => _InvoiceData.fromJson(e as Map<String, dynamic>)).toList();
          if (mounted) setState(() => _invoices = list);
        }
      }
    } catch (_) {}
  }

  Future<void> _loadInvoiceStats() async {
    try {
      final resp = await widget.httpClient
          .get(Uri.parse('$_apiBase/api/chat/invoice/stats'))
          .timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (mounted) setState(() {
          _totalInvoiceCount = (data['total'] as num?)?.toInt() ?? 0;
          _pendingInvoiceCount = (data['pending'] as num?)?.toInt() ?? 0;
          _paidInvoiceCount = (data['paid'] as num?)?.toInt() ?? 0;
        });
      }
    } catch (_) {}
  }

  Future<void> _loadAddress() async {
    try {
      final resp = await widget.httpClient
          .get(Uri.parse('$_apiBase/api/chat/address'))
          .timeout(const Duration(seconds: 8));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['link'] != null) {
          if (mounted) setState(() => _contactLink = data['link'] as String);
          _loadQR();
          return;
        }
      }
    } catch (_) {}
    try {
      final resp = await widget.httpClient
          .post(Uri.parse('$_apiBase/api/chat/address/create'))
          .timeout(const Duration(seconds: 8));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['link'] != null) {
          if (mounted) setState(() => _contactLink = data['link'] as String);
          _loadQR();
        }
      }
    } catch (_) {}
  }

  Future<void> _loadQR() async {
    try {
      final resp = await widget.httpClient
          .get(Uri.parse('$_apiBase/api/chat/qr'))
          .timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200 && mounted) setState(() => _qrImage = resp.bodyBytes);
    } catch (_) {}
  }

  Future<void> _loadContacts() async {
    try {
      final resp = await widget.httpClient
          .get(Uri.parse('$_apiBase/api/chat/contacts'))
          .timeout(const Duration(seconds: 8));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['contacts'] is List) {
          final list = (data['contacts'] as List).map((c) {
            final m = c as Map<String, dynamic>;
            final prof = m['profile'] as Map<String, dynamic>? ?? {};
            return ContactInfo(
              id: (m['contactId'] as num?)?.toInt() ?? 0,
              displayName: (m['localDisplayName'] as String?) ?? (prof['displayName'] as String?) ?? '?',
              fullName: prof['fullName'] as String?,
              msgCount: (m['msg_count'] as num?)?.toInt() ?? 0,
            );
          }).toList();
          if (mounted) setState(() => _contacts = list);
        }
      }
    } catch (_) {}
  }

  Future<void> _loadHistory({String? chatId}) async {
    try {
      var url = '$_apiBase/api/chat/history';
      if (chatId != null) url += '?chat_id=$chatId';
      final resp = await widget.httpClient.get(Uri.parse(url)).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        final list = (data['messages'] as List).map((m) => ChatMessage.fromJson(m as Map<String, dynamic>)).toList();
        if (mounted) { setState(() { _messages = list; _loading = false; }); _scrollDown(); }
      }
    } catch (_) { if (mounted) setState(() => _loading = false); }
  }

  Future<void> _loadContactInfo(String cid) async {
    try {
      final resp = await widget.httpClient
          .get(Uri.parse('$_apiBase/api/chat/contact/info?chat_id=$cid'))
          .timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['count'] != null && mounted) {
          final idx = _contacts.indexWhere((c) => '@${c.id}' == cid);
          if (idx >= 0) {
            _contacts[idx] = ContactInfo(
              id: _contacts[idx].id, displayName: _contacts[idx].displayName,
              fullName: _contacts[idx].fullName, msgCount: (data['count'] as num).toInt(),
            );
            setState(() {});
          }
        }
      }
    } catch (_) {}
  }

  void _connectSSE() {
    _connectSSEInner();
  }

  Duration _sseBackoff() {
    final seconds = [1, 2, 4, 8, 30];
    final idx = _sseReconnectAttempt < seconds.length ? _sseReconnectAttempt : seconds.length - 1;
    return Duration(seconds: seconds[idx]);
  }

  Future<void> _connectSSEInner() async {
    try {
      final client = http.Client();
      final request = http.Request('GET', Uri.parse('$_apiBase/api/chat/stream'));
      final response = await client.send(request);
      if (response.statusCode != 200) {
        client.close();
        _pollTimer = Timer.periodic(const Duration(seconds: 3), (_) => _loadHistory());
        return;
      }
      if (_sseReconnectAttempt > 0 && mounted) {
        debugPrint('[SSE] reconnected after $_sseReconnectAttempt attempt(s)');
      }
      _sseReconnectAttempt = 0;
      if (mounted) setState(() => _reconnecting = false);
      _sseClient = client;
      _sseResponse = response;
      String buffer = '';
      response.stream.transform(utf8.decoder).listen(
        (chunk) {
          buffer += chunk;
          while (buffer.contains('\n\n')) {
            final idx = buffer.indexOf('\n\n');
            final event = buffer.substring(0, idx);
            buffer = buffer.substring(idx + 2);
            if (event.startsWith('event: message\ndata: ')) {
              final data = event.substring('event: message\ndata: '.length);
              try {
                final msg = ChatMessage.fromJson(jsonDecode(data));
                if (mounted) { setState(() {
                  _messages.add(msg);
                  if (_selectedContactId.isEmpty || msg.chatId != _selectedContactId) {
                    _unreadCount++;
                    _perContactUnread[msg.chatId] = (_perContactUnread[msg.chatId] ?? 0) + 1;
                  }
                }); _scrollDown(); }
              } catch (_) {}
            } else if (event.startsWith('event: typing\ndata: ')) {
              final data = event.substring('event: typing\ndata: '.length);
              try {
                final td = jsonDecode(data) as Map<String, dynamic>;
                final chatId = td['chat_id'] as String?;
                final typing = td['typing'] as bool? ?? false;
                if (chatId != null && chatId == _selectedContactId && mounted) {
                  setState(() => _typingContact = typing ? 'typing...' : null);
                  _typingTimer?.cancel();
                  if (typing) _typingTimer = Timer(const Duration(seconds: 4), () { if (mounted) setState(() => _typingContact = null); });
                }
              } catch (_) {}
            }
          }
        },
        onError: (_) { _cleanupSSE(); _scheduleSSERetry(); },
        onDone: () { _cleanupSSE(); _scheduleSSERetry(); },
        cancelOnError: false,
      );
    } catch (_) {
      _pollTimer = Timer.periodic(const Duration(seconds: 3), (_) => _loadHistory());
    }
  }

  void _cleanupSSE() {
    _sseResponse?.stream.listen((_) {});
    _sseResponse?.stream.drain();
    _sseClient?.close();
    _sseClient = null;
    _sseResponse = null;
  }

  void _scheduleSSERetry() {
    if (!mounted) return;
    final backoff = _sseBackoff();
    debugPrint('[SSE] scheduling retry in ${backoff.inSeconds}s (attempt ${_sseReconnectAttempt + 1})');
    setState(() => _reconnecting = true);
    _sseReconnectAttempt++;
    Future.delayed(backoff, () { if (_sseClient == null && mounted) _connectSSE(); });
  }

  void _scrollDown() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollCtrl.hasClients) {
        _scrollCtrl.animateTo(_scrollCtrl.position.maxScrollExtent, duration: const Duration(milliseconds: 200), curve: Curves.easeOut);
      }
    });
  }

  List<ChatMessage> get _conversations {
    final byChatId = <String, ChatMessage>{};
    for (final msg in _messages) {
      final key = msg.chatId.isEmpty ? 'broadcast' : msg.chatId;
      if (!byChatId.containsKey(key) || msg.time.isAfter(byChatId[key]!.time)) byChatId[key] = msg;
    }
    return byChatId.values.toList()..sort((a, b) => b.time.compareTo(a.time));
  }

  List<ChatMessage> get _currentMessages {
    var filtered = _messages;
    if (_selectedContactId.isNotEmpty) {
      filtered = filtered.where((m) => m.chatId == _selectedContactId).toList();
    }
    if (_searching && _searchCtrl.text.trim().isNotEmpty) {
      final q = _searchCtrl.text.trim().toLowerCase();
      filtered = filtered.where((m) =>
        m.text.toLowerCase().contains(q) ||
        (m.fileName?.toLowerCase().contains(q) ?? false) ||
        (m.from.toLowerCase().contains(q))).toList();
    }
    return filtered;
  }

  int? get _selectedContactNumber {
    if (_selectedContactId.startsWith('@')) return int.tryParse(_selectedContactId.substring(1));
    return null;
  }

  Future<void> _sendMessage() async {
    final text = _inputCtrl.text.trim();
    if (text.isEmpty) return;
    final scheduled = _scheduledTime;
    final replyTo = _replyingTo;
    setState(() { _inputCtrl.clear(); _scheduledTime = null; _replyingTo = null; _replyingText = null; });
    if (scheduled != null && scheduled.isAfter(DateTime.now())) {
      try {
        final body = <String, dynamic>{'text': text, 'send_at': scheduled.toUtc().toIso8601String()};
        if (_selectedContactId.isNotEmpty) { body['chat_id'] = _selectedContactId; final cid = _selectedContactNumber; if (cid != null) body['contact_id'] = cid; }
        if (replyTo != null) body['reply_to_id'] = replyTo;
        await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/schedule'),
          headers: {'Content-Type': 'application/json'}, body: jsonEncode(body)).timeout(const Duration(seconds: 10));
        if (mounted) ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Scheduled for ${_timeFmt.format(scheduled)}'), backgroundColor: Colors.amber));
      } catch (_) {}
      return;
    }
    try {
      final body = <String, dynamic>{'text': text};
      if (_selectedContactId.isNotEmpty) { body['chat_id'] = _selectedContactId; final cid = _selectedContactNumber; if (cid != null) body['contact_id'] = cid; }
      if (replyTo != null) body['reply_to_id'] = replyTo;
      await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/send'),
        headers: {'Content-Type': 'application/json'}, body: jsonEncode(body)).timeout(const Duration(seconds: 10));
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Send error: $e'), backgroundColor: Colors.red));
    }
  }

  Future<void> _connectViaLink() async {
    final link = _connectCtrl.text.trim();
    if (link.isEmpty) return;
    try {
      final resp = await widget.httpClient
          .post(Uri.parse('$_apiBase/api/chat/connect'), headers: {'Content-Type': 'application/json'}, body: jsonEncode({'link': link}))
          .timeout(const Duration(seconds: 15));
      final data = jsonDecode(resp.body) as Map<String, dynamic>;
      if (mounted) {
        _connectCtrl.clear();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(data['ok'] == true ? 'Connection sent!' : 'Error: ${data['error']}'),
            backgroundColor: data['ok'] == true ? Colors.green : Colors.red));
        if (data['ok'] == true) _loadContacts();
      }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Connect error: $e'), backgroundColor: Colors.red));
    }
  }

  void _showBroadcastDialog() {
    final ctrl = TextEditingController();
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Broadcast', style: TextStyle(fontSize: 20)),
      content: TextField(controller: ctrl, maxLines: 3, style: const TextStyle(fontSize: 18, color: Colors.white),
        decoration: InputDecoration(hintText: 'Message to all contacts...', hintStyle: TextStyle(fontSize: 18, color: Colors.grey[600]),
          border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none), filled: true, fillColor: Colors.grey.shade800)),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
        ElevatedButton(onPressed: () async {
          final text = ctrl.text.trim();
          if (text.isEmpty) return;
          Navigator.pop(ctx);
          try {
            await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/broadcast'),
              headers: {'Content-Type': 'application/json'}, body: jsonEncode({'text': text})).timeout(const Duration(seconds: 15));
          } catch (_) {}
        }, child: const Text('Send', style: TextStyle(fontSize: 18))),
      ],
    ));
  }

  void _showParanoidxDialog() {
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: Row(children: [
        Icon(Icons.shield, color: (_pxV2Ray && _pxVPN && _pxTor && _pxSimplex) ? Colors.green : Colors.red, size: 24),
        const SizedBox(width: 8),
        const Text('ParanoidX Bridge', style: TextStyle(fontSize: 20)),
      ]),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        _pxLayerRow('V2Ray', _pxV2Ray, Colors.blue, 'Traffic obfuscation'),
        _pxLayerRow('WireGuard', _pxVPN, Colors.teal, 'VPN tunnel'),
        _pxLayerRow('Tor', _pxTor, Colors.orange, 'Onion routing'),
        _pxLayerRow('SimpleX', _pxSimplex, Colors.green, 'Metadata-free chat'),
        const SizedBox(height: 12),
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: Colors.grey.shade800, borderRadius: BorderRadius.circular(6)),
          child: Text(
            'Proxy chain: V2Ray -> VPN -> Tor -> SimpleX',
            style: TextStyle(fontSize: 14, color: Colors.grey[400], fontFamily: 'monospace'),
          ),
        ),
      ]),
      actions: [
        TextButton(
          onPressed: () {
            Navigator.pop(ctx);
            _showBroadcastDialog();
          },
          child: const Text('Broadcast', style: TextStyle(fontSize: 16)),
        ),
        TextButton(
          onPressed: () {
            Navigator.pop(ctx);
            _openParanoidXScreen();
          },
          child: const Text('Settings', style: TextStyle(fontSize: 16, color: Colors.cyan)),
        ),
        ElevatedButton(
          onPressed: () => Navigator.pop(ctx),
          child: const Text('Close', style: TextStyle(fontSize: 16)),
        ),
      ],
    ));
  }

  Widget _pxLayerRow(String label, bool healthy, Color color, String desc) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3),
      child: Row(children: [
        _chainDot(Theme.of(context), label[0], healthy, color),
        const SizedBox(width: 8),
        Text(label, style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: healthy ? Colors.white : Colors.grey)),
        const SizedBox(width: 6),
        Text(desc, style: TextStyle(fontSize: 13, color: Colors.grey[500])),
        const Spacer(),
        Icon(healthy ? Icons.check_circle : Icons.cancel, size: 18, color: healthy ? color : Colors.red),
      ]),
    );
  }

  void _showCreateInvoiceDialog() {
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: const Text('Create Invoice', style: TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        TextField(controller: _invAmountCtrl, style: const TextStyle(fontSize: 18, color: Colors.white), keyboardType: TextInputType.number,
          decoration: InputDecoration(labelText: 'Amount (XAG)', labelStyle: const TextStyle(fontSize: 18),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none), filled: true, fillColor: Colors.grey.shade800)),
        const SizedBox(height: 8),
        TextField(controller: _invDescCtrl, style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(labelText: 'Description', labelStyle: const TextStyle(fontSize: 18),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none), filled: true, fillColor: Colors.grey.shade800)),
      ]),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
        ElevatedButton(onPressed: () async {
          final amt = double.tryParse(_invAmountCtrl.text);
          if (amt == null || amt <= 0) return;
          Navigator.pop(ctx);
          try {
            final cid = _selectedContactNumber;
            await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/invoice/create'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'chat_id': _selectedContactId, 'contact_id': cid, 'amount': amt, 'description': _invDescCtrl.text.trim()}),
            ).timeout(const Duration(seconds: 10));
            _invAmountCtrl.clear(); _invDescCtrl.clear();
            _loadInvoices(); _loadInvoiceStats();
          } catch (_) {}
        }, child: const Text('Create', style: TextStyle(fontSize: 18))),
      ],
    ));
  }

  void _showForwardDialog(ChatMessage msg) {
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Forward to...', style: TextStyle(fontSize: 20)),
      content: SizedBox(width: double.maxFinite, child: ListView.builder(
        shrinkWrap: true, itemCount: _contacts.length,
        itemBuilder: (_, i) {
          final c = _contacts[i];
          return ListTile(
            dense: true,
            leading: CircleAvatar(radius: 16, backgroundColor: Colors.grey.shade700,
              child: Text(c.displayName[0].toUpperCase(), style: const TextStyle(fontSize: 14))),
            title: Text(c.displayName, style: const TextStyle(fontSize: 18)),
            subtitle: Text('#${c.id}', style: TextStyle(fontSize: 14, color: Colors.grey[600])),
            onTap: () async {
              Navigator.pop(ctx);
              try {
                await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/send'),
                  headers: {'Content-Type': 'application/json'},
                  body: jsonEncode({'chat_id': '@${c.id}', 'contact_id': c.id, 'text': msg.text, 'forwarded': true}),
                ).timeout(const Duration(seconds: 10));
              } catch (_) {}
            },
          );
        },
      )),
    ));
  }

  Future<void> _togglePin(ChatMessage msg) async {
    try {
      await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/pin'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'message_id': msg.id, 'chat_id': msg.chatId}),
      ).timeout(const Duration(seconds: 5));
    } catch (_) {}
  }

  Future<void> _toggleReact(ChatMessage msg, String emoji) async {
    try {
      await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/react'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'message_id': msg.id, 'chat_id': msg.chatId, 'emoji': emoji}),
      ).timeout(const Duration(seconds: 5));
    } catch (_) {}
  }

  void _showReactionPicker(ChatMessage msg) {
    final categories = [
      ('Smileys', ['👍', '❤️', '😂', '😮', '😢', '🙏', '🔥', '🎉', '😍', '😭', '😤', '🤣']),
      ('Gestures', ['👏', '💪', '✌️', '🤝', '👋', '🙌', '🫡', '🤞']),
      ('Objects', ['⭐', '💯', '✅', '❌', '💀', '🎯', '🏆', '🔝']),
    ];
    showDialog(context: context, builder: (ctx) => StatefulBuilder(builder: (ctx, setDlgState) {
      var catIdx = 0;
      return AlertDialog(
        backgroundColor: Colors.grey[900], title: const Text('React', style: TextStyle(fontSize: 20)),
        content: Column(mainAxisSize: MainAxisSize.min, children: [
          Row(mainAxisAlignment: MainAxisAlignment.center, children: List.generate(categories.length, (i) =>
            Padding(padding: const EdgeInsets.symmetric(horizontal: 4),
              child: ChoiceChip(label: Text(categories[i].$1, style: const TextStyle(fontSize: 14)),
                selected: i == catIdx, onSelected: (_) => setDlgState(() => catIdx = i),
                selectedColor: Colors.cyan.shade800, backgroundColor: Colors.grey.shade800)))),
          const SizedBox(height: 8),
          Wrap(spacing: 8, runSpacing: 8,
            children: categories[catIdx].$2.map((e) => GestureDetector(
              onTap: () { Navigator.pop(ctx); _toggleReact(msg, e); },
              child: Container(padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(color: Colors.grey.shade800, borderRadius: BorderRadius.circular(8)),
                child: Text(e, style: const TextStyle(fontSize: 32))),
            )).toList(),
          ),
        ]),
      );
    }));
  }

  Future<void> _editMessage(ChatMessage msg) async {
    final ctrl = TextEditingController(text: msg.text);
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Edit Message', style: TextStyle(fontSize: 20)),
      content: TextField(controller: ctrl, maxLines: 3, style: const TextStyle(fontSize: 18, color: Colors.white),
        decoration: InputDecoration(border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
          filled: true, fillColor: Colors.grey.shade800)),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
        ElevatedButton(onPressed: () async {
          Navigator.pop(ctx);
          try {
            await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/edit'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'message_id': msg.id, 'text': ctrl.text.trim()}),
            ).timeout(const Duration(seconds: 5));
          } catch (_) {}
        }, child: const Text('Save', style: TextStyle(fontSize: 18))),
      ],
    ));
  }

  Future<void> _deleteMessage(ChatMessage msg) async {
    try {
      await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/delete'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'id': msg.id}), // match Go endpoint param key
      ).timeout(const Duration(seconds: 5));
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    if (!_setupDone) return _buildSetupGuide(theme);
    return Column(
      children: [
        _buildServerStatusBar(theme),
        _buildTabBar(theme),
        Expanded(
          child: IndexedStack(
            index: _selectedTab,
            children: [
              _buildChatView(theme),
              _buildQRView(theme),
              _buildContactsView(theme),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildServerStatusBar(ThemeData theme) {
    final isConnected = _bridgeStatus == 'connected';
    return GestureDetector(
      onTap: _showParanoidxDialog,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
        decoration: BoxDecoration(color: Colors.grey.shade900, border: Border(bottom: BorderSide(color: Colors.grey.shade800))),
        child: Row(children: [
          // ParanoidX proxy chain indicators
          _chainDot(theme, 'V', _pxV2Ray, Colors.blue),
          const SizedBox(width: 3),
          _chainDot(theme, 'W', _pxVPN, Colors.teal),
          const SizedBox(width: 3),
          _chainDot(theme, 'T', _pxTor, Colors.orange),
          const SizedBox(width: 3),
          _chainDot(theme, 'S', _pxSimplex, Colors.green),
          const SizedBox(width: 8),
          PulseDot(
            size: 10,
            color: isConnected ? Colors.green : (_bridgeStatus == 'connecting' ? Colors.orange : Colors.red),
            pulsing: isConnected,
          ),
          const SizedBox(width: 6),
          Text(_bridgeStatus, style: TextStyle(fontSize: 16, color: Colors.grey[400])),
          if (_reconnecting) ...[
            const SizedBox(width: 8),
            SizedBox(width: 10, height: 10, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.orange)),
            const SizedBox(width: 4),
            Text('Reconnecting...', style: TextStyle(fontSize: 14, color: Colors.orange)),
          ],
          const SizedBox(width: 12),
          Icon(Icons.message, size: 16, color: Colors.grey[500]),
          const SizedBox(width: 4),
          Text('$_totalMessages', style: TextStyle(fontSize: 16, color: Colors.grey[400])),
          if (_serverStatusText.isNotEmpty) ...[
            const SizedBox(width: 12),
            Expanded(child: Text(_serverStatusText, style: TextStyle(fontSize: 14, color: Colors.cyan[300]), overflow: TextOverflow.ellipsis)),
          ],
          const Spacer(),
          Icon(Icons.campaign, size: 18, color: theme.colorScheme.primary),
          const SizedBox(width: 4),
          Text('Broadcast', style: TextStyle(fontSize: 14, color: theme.colorScheme.primary)),
        ]),
      ),
    );
  }

  Widget _chainDot(ThemeData theme, String label, bool healthy, Color color) {
    return Tooltip(
      message: '$label: ${healthy ? "UP" : "DOWN"}',
      child: Container(
        width: 16, height: 16,
        decoration: BoxDecoration(
          color: healthy ? color : Colors.grey.shade800,
          shape: BoxShape.circle,
          border: Border.all(color: healthy ? color : Colors.grey.shade600, width: 1),
        ),
        child: Center(child: Text(label, style: TextStyle(fontSize: 9, fontWeight: FontWeight.bold, color: healthy ? Colors.white : Colors.grey))),
      ),
    );
  }

  Widget _buildTabBar(ThemeData theme) {
    return Container(
      decoration: BoxDecoration(color: Colors.grey.shade900, border: Border(bottom: BorderSide(color: Colors.grey.shade800))),
      child: Row(children: [
        _tabBtn(theme, 'Chat', 0, Icons.chat, badge: _unreadCount),
        _tabBtn(theme, 'QR', 1, Icons.qr_code),
        _tabBtn(theme, 'Contacts', 2, Icons.people),
      ]),
    );
  }

  Widget _tabBtn(ThemeData theme, String label, int idx, IconData icon, {int badge = 0}) {
    final active = _selectedTab == idx;
    return Expanded(
      child: InkWell(
        onTap: () => setState(() { _selectedTab = idx; if (idx == 0) { _unreadCount = 0; _perContactUnread.clear(); } }),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 10),
          decoration: active ? BoxDecoration(border: Border(bottom: BorderSide(color: theme.colorScheme.primary, width: 2))) : null,
          child: Row(mainAxisSize: MainAxisSize.min, mainAxisAlignment: MainAxisAlignment.center, children: [
            Stack(children: [
              Icon(icon, size: 20, color: active ? theme.colorScheme.primary : Colors.grey),
              if (badge > 0) Positioned(right: -6, top: -2, child: Container(
                padding: const EdgeInsets.all(3),
                decoration: BoxDecoration(color: Colors.red, shape: BoxShape.circle),
                child: Text('$badge', style: const TextStyle(fontSize: 10, color: Colors.white, fontWeight: FontWeight.bold)))),
            ]),
            const SizedBox(width: 6),
            Text(label, style: TextStyle(fontSize: 16, fontWeight: active ? FontWeight.bold : FontWeight.normal, color: active ? Colors.white : Colors.grey)),
          ]),
        ),
      ),
    );
  }

  Widget _buildQRView(ThemeData theme) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(children: [
        Text('My Contact Link', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold, fontSize: 22)),
        const SizedBox(height: 16),
        if (_qrImage != null)
          ClipRRect(borderRadius: BorderRadius.circular(12), child: Image.memory(_qrImage!, width: 320, height: 320))
        else
          Container(width: 320, height: 320, decoration: BoxDecoration(color: Colors.grey.shade800, borderRadius: BorderRadius.circular(12)),
            child: const Center(child: CircularProgressIndicator())),
        const SizedBox(height: 16),
        if (_contactLink.isNotEmpty) ...[
          Container(padding: const EdgeInsets.all(12), decoration: BoxDecoration(color: Colors.grey.shade800, borderRadius: BorderRadius.circular(8)),
            child: SelectableText(_contactLink, style: const TextStyle(fontSize: 14, fontFamily: 'monospace', color: Colors.grey))),
          const SizedBox(height: 8),
          IconButton(icon: const Icon(Icons.copy, size: 24), tooltip: 'Copy link', onPressed: () {
            Clipboard.setData(ClipboardData(text: _contactLink));
            ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Link copied!'), duration: Duration(seconds: 1)));}),
        ],
        const SizedBox(height: 16),
        TextField(controller: _connectCtrl, style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(hintText: 'Paste connection link to connect...', hintStyle: TextStyle(fontSize: 18, color: Colors.grey[600]),
            filled: true, fillColor: Colors.grey.shade800,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
            contentPadding: const EdgeInsets.all(12),
            suffixIcon: IconButton(icon: Icon(Icons.link, size: 22, color: theme.colorScheme.primary), onPressed: _connectViaLink)),
          onSubmitted: (_) => _connectViaLink()),
        const SizedBox(height: 16),
        _buildInvoiceStats(),
      ]),
    );
  }

  Widget _buildInvoiceStats() {
    if (_totalInvoiceCount == 0 && _pendingInvoiceCount == 0 && _paidInvoiceCount == 0) return const SizedBox.shrink();
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(color: Colors.grey.shade900, borderRadius: BorderRadius.circular(8)),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text('Invoices', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.grey[400])),
        const SizedBox(height: 6),
        Row(children: [
          _statChip('Total', Icons.receipt, Colors.blue),
          const SizedBox(width: 6),
          _statChip('Pending', Icons.hourglass_empty, Colors.orange),
          const SizedBox(width: 6),
          _statChip('Paid', Icons.check_circle, Colors.green),
        ]),
      ]),
    );
  }

  Widget _statChip(String label, IconData icon, Color color) {
    int count = 0;
    if (label == 'Total') count = _totalInvoiceCount;
    else if (label == 'Pending') count = _pendingInvoiceCount;
    else if (label == 'Paid') count = _paidInvoiceCount;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(color: color.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(12)),
      child: Row(mainAxisSize: MainAxisSize.min, children: [
        Icon(icon, size: 16, color: color),
        const SizedBox(width: 4),
        Text('$count', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: color)),
        const SizedBox(width: 4),
        Text(label, style: TextStyle(fontSize: 14, color: color)),
      ]),
    );
  }

  Widget _buildContactsView(ThemeData theme) {
    final filteredContacts = _selectedGroup.isEmpty ? _contacts : _contacts.where((c) {
      final grp = _contactGroups.where((g) => g.name == _selectedGroup).firstOrNull;
      return grp != null && grp.memberIds.contains(c.id);
    }).toList();
    return Column(children: [
      Container(padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        child: Row(children: [
          Text('Contacts (${_contacts.length})', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.grey[400])),
          const Spacer(),
          if (_contactGroups.isNotEmpty)
            Padding(padding: const EdgeInsets.only(right: 4),
              child: DropdownButton<String>(
                value: _selectedGroup.isEmpty ? null : _selectedGroup,
                hint: const Text('Group', style: TextStyle(fontSize: 14)),
                items: [const DropdownMenuItem(value: '', child: Text('All', style: TextStyle(fontSize: 14))),
                  ..._contactGroups.map((g) => DropdownMenuItem(value: g.name, child: Text(g.name, style: const TextStyle(fontSize: 14)))),
                ],
                onChanged: (v) => setState(() => _selectedGroup = v ?? ''),
              ),
            ),
          PopupMenuButton<String>(
            icon: const Icon(Icons.settings, size: 20),
            onSelected: (v) {
              if (v == 'analytics') _showAnalytics();
              else if (v == 'webhook') _showWebhookConfig();
              else if (v == 'status') _showStatusPage();
              else if (v == 'templates') _showTemplatesDialog();
              else if (v == 'labels') _showLabelsDialog();
              else if (v == 'search') _showAdvancedSearchDialog();
              else if (v == 'forward') _showBatchForwardDialog();
              else if (v == 'clearold') _showClearOldDialog();
              else if (v == 'autodelete') _showAutoDeleteDialog();
              else if (v == 'export') _showExportDialog();
              else if (v == 'backup') _showBackupRestoreDialog();
              else if (v == 'media') _showMediaGallery();
              else if (v == 'auditlog') _showAuditLog();
              else if (v == 'suggestions') _loadSuggestions();
              else if (v == 'scheduled') _showScheduledMessages();
              else if (v == 'groups') _showGroupManager();
              else if (v == 'qrscan') _showQRScanner();
              else if (v == 'paranoidx') _openParanoidXScreen();
            },
            itemBuilder: (_) => [
              const PopupMenuItem(value: 'analytics', child: Text('Analytics', style: TextStyle(fontSize: 14))),
              const PopupMenuItem(value: 'webhook', child: Text('Webhook Config', style: TextStyle(fontSize: 14))),
              const PopupMenuItem(value: 'status', child: Text('System Status', style: TextStyle(fontSize: 14))),
              const PopupMenuItem(value: 'templates', child: Text('Templates', style: TextStyle(fontSize: 14))),
              const PopupMenuItem(value: 'labels', child: Text('Labels', style: TextStyle(fontSize: 14))),
              const PopupMenuItem(value: 'search', child: Text('Advanced Search', style: TextStyle(fontSize: 14))),
              const PopupMenuItem(value: 'forward', child: Text('Batch Forward', style: TextStyle(fontSize: 14))),
              const PopupMenuItem(value: 'clearold', child: Text('Clear Old', style: TextStyle(fontSize: 14))),
              const PopupMenuItem(value: 'autodelete', child: Text('Auto-Delete', style: TextStyle(fontSize: 14))),
              const PopupMenuItem(value: 'export', child: Text('Export Chat', style: TextStyle(fontSize: 14))),
              const PopupMenuItem(value: 'backup', child: Text('Backup/Restore', style: TextStyle(fontSize: 14))),
              const PopupMenuDivider(),
              PopupMenuItem(value: 'media', child: Row(children: [
                Icon(Icons.photo_library, size: 16, color: Colors.teal),
                const SizedBox(width: 6),
                const Text('Media Gallery', style: TextStyle(fontSize: 14, color: Colors.teal)),
              ])),
              PopupMenuItem(value: 'auditlog', child: Row(children: [
                Icon(Icons.security, size: 16, color: Colors.cyan),
                const SizedBox(width: 6),
                const Text('Security Audit', style: TextStyle(fontSize: 14, color: Colors.cyan)),
              ])),
              PopupMenuItem(value: 'suggestions', child: Row(children: [
                Icon(Icons.lightbulb, size: 16, color: Colors.orange),
                const SizedBox(width: 6),
                const Text('AI Suggestions', style: TextStyle(fontSize: 14, color: Colors.orange)),
              ])),
              const PopupMenuDivider(),
              PopupMenuItem(value: 'scheduled', child: Row(children: [
                Icon(Icons.schedule, size: 16, color: Colors.amber),
                const SizedBox(width: 6),
                const Text('Scheduled Msgs', style: TextStyle(fontSize: 14, color: Colors.amber)),
              ])),
              PopupMenuItem(value: 'groups', child: Row(children: [
                Icon(Icons.group, size: 16, color: Colors.green),
                const SizedBox(width: 6),
                const Text('Groups', style: TextStyle(fontSize: 14, color: Colors.green)),
              ])),
              PopupMenuItem(value: 'qrscan', child: Row(children: [
                Icon(Icons.qr_code_scanner, size: 16, color: Colors.cyan),
                const SizedBox(width: 6),
                const Text('Scan QR', style: TextStyle(fontSize: 14, color: Colors.cyan)),
              ])),
              PopupMenuItem(value: 'paranoidx', child: Row(children: [
                Icon(Icons.shield, size: 16, color: Colors.red),
                const SizedBox(width: 6),
                const Text('ParanoidX', style: TextStyle(fontSize: 14, color: Colors.red)),
              ])),
            ],
          ),
          IconButton(icon: const Icon(Icons.smart_toy, size: 20), tooltip: 'Auto-reply', onPressed: _showAutoReplyDialog),
          IconButton(icon: const Icon(Icons.group_add, size: 20), tooltip: 'Create group', onPressed: _showCreateGroupDialog),
          IconButton(icon: const Icon(Icons.refresh, size: 20), onPressed: _loadContacts),
        ])),
      Expanded(
        child: filteredContacts.isEmpty
            ? Center(child: Column(mainAxisSize: MainAxisSize.min, children: [
                Icon(Icons.people_outline, size: 64, color: Colors.grey[700]),
                const SizedBox(height: 12),
                Text('No contacts yet', style: TextStyle(fontSize: 18, color: Colors.grey[500])),
                const SizedBox(height: 4),
                Text('Share QR code from the QR tab', style: TextStyle(fontSize: 16, color: Colors.grey[600])),
              ]))
            : ListView.builder(
                itemCount: filteredContacts.length,
                itemBuilder: (ctx, i) {
                  final c = filteredContacts[i];
                    final chatId = '@${c.id}';
                    final trustLevel = _contactTrust[chatId];
                    final status = _contactStatuses[chatId];
                    return ListTile(
                    dense: true,
                    leading: Stack(children: [
                      CircleAvatar(radius: 18, backgroundColor: Colors.grey.shade700,
                        child: Text(c.displayName[0].toUpperCase(), style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold))),
                      if (status != null && status == 'online')
                        Positioned(right: 0, bottom: 0, child: Container(width: 10, height: 10,
                          decoration: BoxDecoration(shape: BoxShape.circle, color: Colors.green, border: Border.all(color: Colors.grey.shade900, width: 2)))),
                    ]),
                    title: Row(children: [
                      _trustBadge(trustLevel),
                      Expanded(child: Text(c.displayName, style: const TextStyle(fontSize: 18))),
                    ]),
                    subtitle: Row(children: [
                      if (status != null && status != 'offline')
                        Text(status, style: TextStyle(fontSize: 14, color: status == 'online' ? Colors.green[400] : Colors.orange[400])),
                      if (status != null && status != 'offline') const SizedBox(width: 6),
                      if (c.fullName != null && c.fullName!.isNotEmpty)
                        Text(c.fullName!, style: TextStyle(fontSize: 14, color: Colors.grey[500])),
                    ]),
                    trailing: Row(mainAxisSize: MainAxisSize.min, children: [
                      if (_perContactUnread[chatId] != null && _perContactUnread[chatId]! > 0)
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(color: Colors.red, borderRadius: BorderRadius.circular(10)),
                          child: Text('${_perContactUnread[chatId]}', style: const TextStyle(fontSize: 12, color: Colors.white, fontWeight: FontWeight.bold)),
                        ),
                      if (c.msgCount > 0)
                        Padding(padding: const EdgeInsets.only(left: 6, right: 8),
                          child: Text('${c.msgCount}', style: TextStyle(fontSize: 14, color: Colors.grey[500]))),
                      Text('#${c.id}', style: TextStyle(fontSize: 14, color: Colors.grey[600])),
                    ]),
                    onTap: () { setState(() { _selectedContactId = '@${c.id}'; _selectedTab = 0; }); _loadContactInfo('@${c.id}'); },
                    onLongPress: () => _showContactLongPressActions(c),
                  );
                },
              ),
      ),
    ]);
  }

  void _showAutoReplyDialog() {
    final kwCtrl = TextEditingController();
    final rspCtrl = TextEditingController();
    showDialog(context: context, builder: (ctx) => StatefulBuilder(builder: (ctx, setDlgState) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: const Text('Auto-reply Rules', style: TextStyle(fontSize: 20)),
      content: SizedBox(width: 400, child: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
        if (_autoReplyRules.isNotEmpty) ...[
          Text('Active rules:', style: TextStyle(fontSize: 16, color: Colors.grey[400])),
          const SizedBox(height: 4),
          Container(
            constraints: const BoxConstraints(maxHeight: 200),
            child: ListView.builder(shrinkWrap: true, itemCount: _autoReplyRules.length,
              itemBuilder: (_, i) => ListTile(dense: true, title: Text('"${_autoReplyRules[i].keyword}" → "${_autoReplyRules[i].response}"', style: const TextStyle(fontSize: 14)),
                trailing: IconButton(icon: const Icon(Icons.delete, size: 18, color: Colors.red),
                  onPressed: () async {
                    _autoReplyRules.removeAt(i);
                    await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/auto-reply'),
                      headers: {'Content-Type': 'application/json'}, body: jsonEncode({'rules': _autoReplyRules.map((r) => r.toJson()).toList()}),
                    ).timeout(const Duration(seconds: 5));
                    setDlgState(() {});
                  }),
              )),
          ),
          const Divider(),
        ],
        TextField(controller: kwCtrl, style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(labelText: 'Keyword', labelStyle: const TextStyle(fontSize: 16),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
            filled: true, fillColor: Colors.grey.shade800)),
        const SizedBox(height: 8),
        TextField(controller: rspCtrl, style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(labelText: 'Response', labelStyle: const TextStyle(fontSize: 16),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
            filled: true, fillColor: Colors.grey.shade800)),
        const SizedBox(height: 8),
        ElevatedButton(onPressed: () async {
          if (kwCtrl.text.trim().isEmpty || rspCtrl.text.trim().isEmpty) return;
          _autoReplyRules.add(_AutoReplyRule(keyword: kwCtrl.text.trim(), response: rspCtrl.text.trim()));
          await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/auto-reply'),
            headers: {'Content-Type': 'application/json'}, body: jsonEncode({'rules': _autoReplyRules.map((r) => r.toJson()).toList()}),
          ).timeout(const Duration(seconds: 5));
          kwCtrl.clear(); rspCtrl.clear();
          setDlgState(() {});
        }, child: const Text('Add Rule', style: TextStyle(fontSize: 16))),
      ])),
      actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Done', style: TextStyle(fontSize: 18)))],
    )));
  }

  void _openParanoidXScreen() {
    Navigator.push(context, MaterialPageRoute(builder: (_) =>
      ParanoidXScreen(apiBase: _apiBase, httpClient: widget.httpClient)));
  }

  void _showAnalytics() async {
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/chat/analytics')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200 && mounted) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
          if (data['ok'] == true && data['analytics'] != null) {
            final a = data['analytics'] as Map<String, dynamic>;
            final total = a['total_messages'] ?? 0;
            final today = a['messages_today'] ?? 0;
            final chats = a['unique_chats'] ?? 0;
            showDialog(context: context, builder: (ctx) => AlertDialog(
              backgroundColor: Colors.grey[900], title: const Text('Chat Analytics', style: TextStyle(fontSize: 20)),
              content: Column(mainAxisSize: MainAxisSize.min, children: [
                Row(children: [
                  _statCard('Total', '$total', Icons.message),
                  const SizedBox(width: 8),
                  _statCard('Today', '$today', Icons.today),
                  const SizedBox(width: 8),
                  _statCard('Chats', '$chats', Icons.forum),
                ]),
                const SizedBox(height: 12),
                _anaRow('Pinned', '${a['pinned_messages'] ?? 0}'),
                _anaRow('Reactions', '${a['reactions_count'] ?? 0}'),
                _anaRow('Avg length', '${(a['avg_message_length'] as num?)?.toStringAsFixed(1) ?? '0'} chars'),
              ]),
              actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
            ));
          }
      }
    } catch (_) {}
  }

  Widget _statCard(String label, String value, IconData icon) {
    return Expanded(child: Container(padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(color: Colors.grey.shade800, borderRadius: BorderRadius.circular(8)),
      child: Column(children: [
        Icon(icon, size: 24, color: Colors.cyan[300]),
        const SizedBox(height: 4),
        Text(value, style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: Colors.white)),
        Text(label, style: TextStyle(fontSize: 14, color: Colors.grey[400])),
      ])));
  }

  Widget _anaRow(String label, String value) {
    return Padding(padding: const EdgeInsets.symmetric(vertical: 3),
      child: Row(children: [
        Text('$label: ', style: TextStyle(fontSize: 16, color: Colors.grey[400])),
        Text(value, style: const TextStyle(fontSize: 16, color: Colors.white)),
      ]));
  }

  void _showWebhookConfig() async {
    final urlCtrl = TextEditingController();
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/chat/webhook')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['config'] != null) {
          final cfg = data['config'] as Map<String, dynamic>;
          urlCtrl.text = cfg['url'] as String? ?? '';
        }
      }
    } catch (_) {}
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Webhook Config', style: TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        TextField(controller: urlCtrl, style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(labelText: 'Webhook URL', labelStyle: const TextStyle(fontSize: 16),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
            filled: true, fillColor: Colors.grey.shade800)),
        const SizedBox(height: 8),
        Text('Receives POST with new messages as JSON', style: TextStyle(fontSize: 14, color: Colors.grey[500])),
      ]),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
        ElevatedButton(onPressed: () async {
          await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/webhook'),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({'url': urlCtrl.text.trim(), 'events': ['message']}),
          ).timeout(const Duration(seconds: 5));
          Navigator.pop(ctx);
        }, child: const Text('Save', style: TextStyle(fontSize: 18))),
      ],
    ));
  }

  void _showStatusPage() async {
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/admin/status-page')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200 && mounted) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        showDialog(context: context, builder: (ctx) => AlertDialog(
          backgroundColor: Colors.grey[900], title: const Text('System Status', style: TextStyle(fontSize: 20)),
          content: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
            _anaRow('Status', data['status'] as String? ?? 'unknown'),
            _anaRow('Messages', '${data['messages'] ?? 0}'),
            _anaRow('Bridge', data['bridge'] as String? ?? 'unknown'),
            _anaRow('Uptime', '${data['uptime_hours']?.toStringAsFixed(1) ?? '0'}h'),
          ]),
          actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
        ));
      }
    } catch (_) {}
  }

  void _showTemplatesDialog() async {
    await _loadTemplates();
    final nameCtrl = TextEditingController();
    final textCtrl = TextEditingController();
    showDialog(context: context, builder: (ctx) => StatefulBuilder(builder: (ctx, setDlgState) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Message Templates', style: TextStyle(fontSize: 20)),
      content: SizedBox(width: 400, child: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
        if (_templates.isNotEmpty) ...[
          Text('Saved templates:', style: TextStyle(fontSize: 16, color: Colors.grey[400])),
          const SizedBox(height: 4),
          Container(
            constraints: const BoxConstraints(maxHeight: 200),
            child: ListView.builder(shrinkWrap: true, itemCount: _templates.length,
              itemBuilder: (_, i) => ListTile(dense: true,
                title: Text('${_templates[i].name}: ${_templates[i].text}', style: const TextStyle(fontSize: 14), maxLines: 2, overflow: TextOverflow.ellipsis),
                trailing: IconButton(icon: const Icon(Icons.delete, size: 18, color: Colors.red),
                  onPressed: () async {
                    _templates.removeAt(i);
                    await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/templates'),
                      headers: {'Content-Type': 'application/json'},
                      body: jsonEncode({'templates': _templates.map((t) => {'name': t.name, 'text': t.text}).toList() }),
                    ).timeout(const Duration(seconds: 5));
                    setDlgState(() {});
                  }),
              )),
          ),
          const Divider(),
        ],
        TextField(controller: nameCtrl, style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(labelText: 'Template name', labelStyle: const TextStyle(fontSize: 16),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
            filled: true, fillColor: Colors.grey.shade800)),
        const SizedBox(height: 8),
        TextField(controller: textCtrl, maxLines: 3, style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(labelText: 'Template text', labelStyle: const TextStyle(fontSize: 16),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
            filled: true, fillColor: Colors.grey.shade800)),
        const SizedBox(height: 8),
        ElevatedButton(onPressed: () async {
          if (nameCtrl.text.trim().isEmpty || textCtrl.text.trim().isEmpty) return;
          _templates.add(_TemplateData(name: nameCtrl.text.trim(), text: textCtrl.text.trim()));
          await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/templates'),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({'templates': _templates.map((t) => {'name': t.name, 'text': t.text}).toList() }),
          ).timeout(const Duration(seconds: 5));
          nameCtrl.clear(); textCtrl.clear();
          setDlgState(() {});
        }, child: const Text('Add Template', style: TextStyle(fontSize: 16))),
      ])),
      actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Done', style: TextStyle(fontSize: 18)))],
    )));
  }

  void _showLabelsDialog() {
    final msgId = _currentMessages.lastOrNull?.id ?? '';
    showDialog(context: context, builder: (ctx) {
      final ctrl = TextEditingController();
      final labels = <String>[];
      return StatefulBuilder(builder: (ctx, setDlgState) => AlertDialog(
        backgroundColor: Colors.grey[900], title: const Text('Labels', style: TextStyle(fontSize: 20)),
        content: Column(mainAxisSize: MainAxisSize.min, children: [
          Row(children: [
            Expanded(child: TextField(controller: ctrl, style: const TextStyle(fontSize: 18, color: Colors.white),
              decoration: InputDecoration(hintText: 'New label...', hintStyle: TextStyle(fontSize: 16, color: Colors.grey[600]),
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none), filled: true, fillColor: Colors.grey.shade800))),
            IconButton(icon: const Icon(Icons.add, size: 24), onPressed: () {
              if (ctrl.text.trim().isNotEmpty) { labels.add(ctrl.text.trim()); ctrl.clear(); setDlgState(() {}); }
            }),
          ]),
          if (labels.isNotEmpty) Wrap(spacing: 6, runSpacing: 4, children: labels.map((l) => Chip(
            label: Text(l, style: const TextStyle(fontSize: 14)), deleteIcon: const Icon(Icons.close, size: 16),
            onDeleted: () { labels.remove(l); setDlgState(() {}); },
            backgroundColor: Colors.grey.shade800,
          )).toList()),
        ]),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
          ElevatedButton(onPressed: () async {
            try {
              await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/labels'),
                headers: {'Content-Type': 'application/json'},
                body: jsonEncode({'chat_id': _selectedContactId, 'labels': labels}),
              ).timeout(const Duration(seconds: 5));
            } catch (_) {}
            Navigator.pop(ctx);
          }, child: const Text('Save', style: TextStyle(fontSize: 18))),
        ],
      ));
    });
  }

  void _showAdvancedSearchDialog() {
    final qCtrl = TextEditingController();
    final fromCtrl = TextEditingController();
    DateTime? dateFrom;
    DateTime? dateTo;
    showDialog(context: context, builder: (ctx) => StatefulBuilder(builder: (ctx, setDlgState) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Advanced Search', style: TextStyle(fontSize: 20)),
      content: SizedBox(width: 400, child: Column(mainAxisSize: MainAxisSize.min, children: [
        TextField(controller: qCtrl, style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(labelText: 'Search text', labelStyle: const TextStyle(fontSize: 16),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
            filled: true, fillColor: Colors.grey.shade800)),
        const SizedBox(height: 8),
        TextField(controller: fromCtrl, style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(labelText: 'From (sender)', labelStyle: const TextStyle(fontSize: 16),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
            filled: true, fillColor: Colors.grey.shade800)),
        const SizedBox(height: 8),
        Row(children: [
          Expanded(child: OutlinedButton(onPressed: () async {
            final d = await showDatePicker(context: context, initialDate: dateFrom ?? DateTime.now().subtract(const Duration(days: 7)),
              firstDate: DateTime(2020), lastDate: DateTime.now(),
              builder: (c, child) => Theme(data: Theme.of(c).copyWith(dialogBackgroundColor: Colors.grey[900]), child: child!));
            if (d != null) { dateFrom = d; setDlgState(() {}); }
          }, child: Text(dateFrom != null ? 'From: ${_dateFmt.format(dateFrom!)}' : 'From date', style: const TextStyle(fontSize: 14)))),
          const SizedBox(width: 8),
          Expanded(child: OutlinedButton(onPressed: () async {
            final d = await showDatePicker(context: context, initialDate: dateTo ?? DateTime.now(),
              firstDate: DateTime(2020), lastDate: DateTime.now(),
              builder: (c, child) => Theme(data: Theme.of(c).copyWith(dialogBackgroundColor: Colors.grey[900]), child: child!));
            if (d != null) { dateTo = d; setDlgState(() {}); }
          }, child: Text(dateTo != null ? 'To: ${_dateFmt.format(dateTo!)}' : 'To date', style: const TextStyle(fontSize: 14)))),
        ]),
      ])),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
        ElevatedButton(onPressed: () async {
          Navigator.pop(ctx);
          try {
            var url = '$_apiBase/api/chat/search/advanced?q=${Uri.encodeComponent(qCtrl.text.trim())}';
            if (fromCtrl.text.trim().isNotEmpty) url += '&from=${Uri.encodeComponent(fromCtrl.text.trim())}';
            if (dateFrom != null) url += '&from_date=${dateFrom!.toIso8601String()}';
            if (dateTo != null) url += '&to_date=${dateTo!.toIso8601String()}';
            final resp = await widget.httpClient.get(Uri.parse(url)).timeout(const Duration(seconds: 10));
            if (resp.statusCode == 200) {
              final data = jsonDecode(resp.body) as Map<String, dynamic>;
              final msgs = (data['messages'] as List?)?.map((m) => ChatMessage.fromJson(m as Map<String, dynamic>)).toList() ?? [];
              if (mounted) setState(() => _messages = msgs);
            }
          } catch (_) {}
        }, child: const Text('Search', style: TextStyle(fontSize: 18))),
      ],
    )));
  }

  void _showBatchForwardDialog() {
    final selected = <ChatMessage>{};
    showDialog(context: context, builder: (ctx) => StatefulBuilder(builder: (ctx, setDlgState) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Batch Forward', style: TextStyle(fontSize: 20)),
      content: SizedBox(width: 400, child: Column(mainAxisSize: MainAxisSize.min, children: [
        if (_contacts.isEmpty)
          Text('No contacts', style: TextStyle(fontSize: 16, color: Colors.grey[500]))
        else ...[
          Text('Select messages to forward:', style: TextStyle(fontSize: 16, color: Colors.grey[400])),
          ConstrainedBox(
            constraints: const BoxConstraints(maxHeight: 200),
            child: ListView.builder(shrinkWrap: true, itemCount: _currentMessages.length,
              itemBuilder: (_, i) {
                final m = _currentMessages[i];
                return CheckboxListTile(dense: true, value: selected.contains(m),
                  title: Text(m.text, style: const TextStyle(fontSize: 14), maxLines: 1, overflow: TextOverflow.ellipsis),
                  subtitle: Text(m.displayTime, style: TextStyle(fontSize: 12, color: Colors.grey[600])),
                  onChanged: (v) { if (v == true) selected.add(m); else selected.remove(m); setDlgState(() {}); },
                );
              },
            ),
          ),
          const SizedBox(height: 8),
          DropdownButton<int>(
            value: _contacts.isNotEmpty ? _contacts.first.id : null,
            hint: const Text('Forward to...', style: TextStyle(fontSize: 16)),
            items: _contacts.map((c) => DropdownMenuItem(value: c.id, child: Text(c.displayName, style: const TextStyle(fontSize: 16)))).toList(),
            onChanged: (_) {},
          ),
        ],
      ])),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
        ElevatedButton(onPressed: () async {
          if (selected.isEmpty) return;
          Navigator.pop(ctx);
          try {
            await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/batch-forward'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'message_ids': selected.map((m) => m.id).toList(), 'chat_id': _selectedContactId}),
            ).timeout(const Duration(seconds: 15));
          } catch (_) {}
        }, child: Text('Forward (${selected.length})', style: const TextStyle(fontSize: 18))),
      ],
    )));
  }

  void _showContactLongPressActions(ContactInfo c) {
    final chatId = '@${c.id}';
    final currentTrust = _contactTrust[chatId] ?? 'none';
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.grey[900],
      builder: (ctx) => SafeArea(child: Column(mainAxisSize: MainAxisSize.min, children: [
        ListTile(leading: const Icon(Icons.edit, size: 24), title: const Text('Rename', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _showRenameDialog(c); }),
        ListTile(leading: Icon(Icons.shield, size: 24, color: currentTrust == 'trusted' ? Colors.green : null),
          title: Text(currentTrust == 'trusted' ? 'Remove Trust' : 'Trust Contact', style: const TextStyle(fontSize: 18)),
          onTap: () {
            Navigator.pop(ctx);
            _updateTrust(chatId, currentTrust == 'trusted' ? 'none' : 'trusted');
          }),
        ListTile(leading: Icon(Icons.verified, size: 24, color: currentTrust == 'verified' ? Colors.blue : null),
          title: Text(currentTrust == 'verified' ? 'Remove Verification' : 'Verify Contact', style: const TextStyle(fontSize: 18)),
          onTap: () {
            Navigator.pop(ctx);
            _updateTrust(chatId, currentTrust == 'verified' ? 'none' : 'verified');
          }),
        ListTile(leading: Icon(Icons.block, size: 24, color: currentTrust == 'blocked' ? Colors.red : null),
          title: Text(currentTrust == 'blocked' ? 'Unblock' : 'Block Contact', style: TextStyle(fontSize: 18, color: currentTrust == 'blocked' ? Colors.grey : Colors.red)),
          onTap: () {
            Navigator.pop(ctx);
            _updateTrust(chatId, currentTrust == 'blocked' ? 'none' : 'blocked');
          }),
        ListTile(leading: const Icon(Icons.notifications, size: 24), title: const Text('Notification Sound', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _showSoundDialog(); }),
        ListTile(leading: const Icon(Icons.security, size: 24, color: Colors.cyan), title: const Text('Encryption Info', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _showEncryptionInfo(chatId, c.displayName); }),
      ])),
    );
  }

  Future<void> _showEncryptionInfo(String chatId, String name) async {
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/chat/encryption?chat_id=$chatId')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200 && mounted) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        final encType = data['encryption'] as String? ?? 'unknown';
        final isE2E = encType == 'e2e';
        showDialog(context: context, builder: (ctx) => AlertDialog(
          backgroundColor: Colors.grey[900], title: Text('Encryption: $name', style: const TextStyle(fontSize: 20)),
          content: Column(mainAxisSize: MainAxisSize.min, children: [
            Icon(isE2E ? Icons.lock : Icons.lock_open, size: 48, color: isE2E ? Colors.green : Colors.amber),
            const SizedBox(height: 12),
            Container(padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(color: Colors.grey.shade800, borderRadius: BorderRadius.circular(8)),
              child: Column(children: [
                Row(children: [
                  Icon(Icons.shield, size: 16, color: isE2E ? Colors.green : Colors.grey),
                  const SizedBox(width: 6),
                  Text('${isE2E ? '✅' : '⚠️'} ${encType.toUpperCase()}', style: TextStyle(fontSize: 18, color: isE2E ? Colors.green : Colors.amber)),
                ]),
                const SizedBox(height: 8),
                Text(isE2E ? 'End-to-end encrypted' : 'Not end-to-end encrypted',
                  style: TextStyle(fontSize: 16, color: Colors.grey[300])),
              ]),
            ),
            const SizedBox(height: 12),
            Text('Chat ID: $chatId', style: TextStyle(fontSize: 14, color: Colors.grey[500])),
            Text('Encryption type determines how messages\nare secured in transit.',
              style: TextStyle(fontSize: 14, color: Colors.grey[600]), textAlign: TextAlign.center),
          ]),
          actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
        ));
      }
    } catch (_) {}
  }

  void _showRenameDialog(ContactInfo c) {
    final ctrl = TextEditingController();
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Rename Contact', style: TextStyle(fontSize: 20)),
      content: TextField(controller: ctrl, style: const TextStyle(fontSize: 18, color: Colors.white),
        decoration: InputDecoration(labelText: 'New name', labelStyle: const TextStyle(fontSize: 16),
          border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
          filled: true, fillColor: Colors.grey.shade800)),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
        ElevatedButton(onPressed: () async {
          if (ctrl.text.trim().isEmpty) return;
          Navigator.pop(ctx);
          try {
            await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/contact/alias'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'chat_id': '@${c.id}', 'alias': ctrl.text.trim()}),
            ).timeout(const Duration(seconds: 5));
            _loadContacts();
          } catch (_) {}
        }, child: const Text('Rename', style: TextStyle(fontSize: 18))),
      ],
    ));
  }

  void _showAutoDeleteDialog() {
    final ctrl = TextEditingController(text: '60');
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Auto-Delete Scheduler', style: TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        Text('Auto-delete messages after N minutes', style: TextStyle(fontSize: 16, color: Colors.grey[400])),
        const SizedBox(height: 8),
        TextField(controller: ctrl, keyboardType: TextInputType.number, style: const TextStyle(fontSize: 24, color: Colors.white),
          decoration: InputDecoration(labelText: 'Minutes', labelStyle: const TextStyle(fontSize: 16),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
            filled: true, fillColor: Colors.grey.shade800)),
      ]),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
        ElevatedButton(onPressed: () async {
          final mins = int.tryParse(ctrl.text);
          if (mins == null) return;
          Navigator.pop(ctx);
          try {
            await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/auto-delete'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'minutes': mins, 'chat_id': _selectedContactId}),
            ).timeout(const Duration(seconds: 5));
          } catch (_) {}
        }, child: const Text('Set', style: TextStyle(fontSize: 18))),
      ],
    ));
  }

  void _showExportDialog() async {
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/chat/export')).timeout(const Duration(seconds: 10));
      if (resp.statusCode == 200 && mounted) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        showDialog(context: context, builder: (ctx) => AlertDialog(
          backgroundColor: Colors.grey[900], title: const Text('Export Chat', style: TextStyle(fontSize: 20)),
          content: SizedBox(width: 400, child: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text('Total messages: ${data['total'] ?? 0}', style: const TextStyle(fontSize: 16)),
            Text('Exported: ${data['exported'] ?? 0}', style: const TextStyle(fontSize: 16)),
            Text('File: ${data['file'] ?? ''}', style: TextStyle(fontSize: 14, color: Colors.grey[500])),
          ])),
          actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
        ));
      }
    } catch (_) {}
  }

  void _showBackupRestoreDialog() {
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Backup & Restore', style: TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        ElevatedButton.icon(icon: const Icon(Icons.download), onPressed: () async {
          try {
            final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/chat/backup')).timeout(const Duration(seconds: 10));
            if (resp.statusCode == 200 && mounted) {
              final path = '/tmp/simplex-chat-backup-${DateTime.now().millisecondsSinceEpoch}.json';
              await File(path).writeAsString(resp.body);
              if (mounted) ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(content: Text('Backup saved to $path'), backgroundColor: Colors.green));
            }
          } catch (_) {}
          Navigator.pop(ctx);
        }, label: const Text('Download Backup', style: TextStyle(fontSize: 16))),
        const SizedBox(height: 12),
        ElevatedButton.icon(icon: const Icon(Icons.upload), onPressed: () async {
          Navigator.pop(ctx);
          try {
            final result = await FilePicker.platform.pickFiles(type: FileType.any);
            if (result != null && result.files.single.path != null) {
              final file = File(result.files.single.path!);
              final req = http.MultipartRequest('POST', Uri.parse('$_apiBase/api/chat/backup'));
              req.files.add(await http.MultipartFile.fromPath('file', file.path));
              final resp = await req.send();
              if (resp.statusCode == 200 && mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Backup restored'), backgroundColor: Colors.green));
              }
            }
          } catch (_) {}
        }, label: const Text('Restore Backup', style: TextStyle(fontSize: 16))),
      ]),
      actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
    ));
  }

  void _showClearOldDialog() {
    final daysCtrl = TextEditingController(text: '7');
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Clear Old Messages', style: TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        Text('Delete messages older than N days', style: TextStyle(fontSize: 16, color: Colors.grey[400])),
        const SizedBox(height: 8),
        TextField(controller: daysCtrl, keyboardType: TextInputType.number, style: const TextStyle(fontSize: 24, color: Colors.white),
          decoration: InputDecoration(labelText: 'Days', labelStyle: const TextStyle(fontSize: 16),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
            filled: true, fillColor: Colors.grey.shade800)),
      ]),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
        ElevatedButton(onPressed: () async {
          final days = int.tryParse(daysCtrl.text);
          if (days == null) return;
          Navigator.pop(ctx);
          try {
            var url = '$_apiBase/api/chat/clear-old?days=$days';
            if (_selectedContactId.isNotEmpty) url += '&chat_id=$_selectedContactId';
            await widget.httpClient.post(Uri.parse(url)).timeout(const Duration(seconds: 10));
            _loadHistory();
          } catch (_) {}
        }, child: const Text('Delete', style: TextStyle(fontSize: 18))),
      ],
    ));
  }

  void _showCreateGroupDialog() {
    final nameCtrl = TextEditingController();
    final selectedIds = <int>{};
    showDialog(context: context, builder: (ctx) => StatefulBuilder(builder: (ctx, setDlgState) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: const Text('Create Group', style: TextStyle(fontSize: 20)),
      content: SizedBox(width: 400, child: Column(mainAxisSize: MainAxisSize.min, children: [
        TextField(controller: nameCtrl, style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(labelText: 'Group name', labelStyle: const TextStyle(fontSize: 16),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
            filled: true, fillColor: Colors.grey.shade800)),
        const SizedBox(height: 8),
        if (_contacts.isEmpty)
          Text('No contacts available', style: TextStyle(fontSize: 16, color: Colors.grey[500]))
        else
          ConstrainedBox(
            constraints: const BoxConstraints(maxHeight: 250),
            child: ListView.builder(shrinkWrap: true, itemCount: _contacts.length,
              itemBuilder: (_, i) {
                final c = _contacts[i];
                return CheckboxListTile(
                  dense: true, value: selectedIds.contains(c.id),
                  title: Text(c.displayName, style: const TextStyle(fontSize: 16)),
                  subtitle: Text('#${c.id}', style: TextStyle(fontSize: 14, color: Colors.grey[600])),
                  onChanged: (v) { if (v == true) selectedIds.add(c.id); else selectedIds.remove(c.id); setDlgState(() {}); },
                );
              },
            ),
          ),
      ])),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 18))),
        ElevatedButton(onPressed: () async {
          if (nameCtrl.text.trim().isEmpty || selectedIds.isEmpty) return;
          await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/groups'),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({'name': nameCtrl.text.trim(), 'member_ids': selectedIds.toList()}),
          ).timeout(const Duration(seconds: 5));
          Navigator.pop(ctx);
          _loadContactGroups();
        }, child: const Text('Create', style: TextStyle(fontSize: 18))),
      ],
    )));
  }

  void _showScheduledMessages() async {
    try {
      final resp = await widget.httpClient.get(Uri.parse('$_apiBase/api/chat/schedule')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200 && mounted) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        final list = (data['scheduled'] as List?)?.map((e) => e as Map<String, dynamic>).toList() ?? [];
        if (list.isEmpty) {
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('No scheduled messages'), backgroundColor: Colors.blue));
          return;
        }
        showDialog(context: context, builder: (ctx) => AlertDialog(
          backgroundColor: Colors.grey[900], title: const Text('Scheduled Messages', style: TextStyle(fontSize: 20)),
          content: SizedBox(width: 400, child: Column(mainAxisSize: MainAxisSize.min, children: [
            Text('${list.length} scheduled', style: TextStyle(fontSize: 16, color: Colors.grey[400])),
            const SizedBox(height: 8),
            ConstrainedBox(constraints: const BoxConstraints(maxHeight: 300),
              child: ListView.builder(shrinkWrap: true, itemCount: list.length,
                itemBuilder: (_, i) {
                  final m = list[i];
                  return ListTile(dense: true,
                    title: Text(m['text'] as String? ?? '', style: const TextStyle(fontSize: 14), maxLines: 2, overflow: TextOverflow.ellipsis),
                    subtitle: Text('${m['chat_id']} @ ${m['send_at']}', style: TextStyle(fontSize: 12, color: Colors.grey[600])),
                    trailing: IconButton(icon: const Icon(Icons.cancel, size: 18, color: Colors.red),
                      onPressed: () async {
                        await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/schedule'),
                          headers: {'Content-Type': 'application/json'},
                          body: jsonEncode({'cancel_id': m['id']})).timeout(const Duration(seconds: 5));
                        _showScheduledMessages();
                        Navigator.pop(ctx);
                      }),
                  );
                },
              ),
            ),
          ])),
          actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
        ));
      }
    } catch (_) {}
  }

  void _showGroupManager() {
    showDialog(context: context, builder: (ctx) => StatefulBuilder(builder: (ctx, setDlgState) => AlertDialog(
      backgroundColor: Colors.grey[900], title: const Text('Contact Groups', style: TextStyle(fontSize: 20)),
      content: SizedBox(width: 400, child: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          ElevatedButton.icon(icon: const Icon(Icons.add, size: 18), onPressed: () { Navigator.pop(ctx); _showCreateGroupDialog(); },
            label: const Text('New Group', style: TextStyle(fontSize: 14))),
        ]),
        const SizedBox(height: 8),
        if (_contactGroups.isEmpty)
          Text('No groups created yet', style: TextStyle(fontSize: 16, color: Colors.grey[500]))
        else
          ConstrainedBox(constraints: const BoxConstraints(maxHeight: 300),
            child: ListView.builder(shrinkWrap: true, itemCount: _contactGroups.length,
              itemBuilder: (_, i) {
                final g = _contactGroups[i];
                return ListTile(dense: true,
                  title: Text(g.name, style: const TextStyle(fontSize: 16)),
                  subtitle: Text('${g.memberIds.length} members', style: TextStyle(fontSize: 14, color: Colors.grey[600])),
                  trailing: IconButton(icon: const Icon(Icons.delete, size: 18, color: Colors.red),
                    onPressed: () async {
                      await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/groups'),
                        headers: {'Content-Type': 'application/json'},
                        body: jsonEncode({'delete': g.name})).timeout(const Duration(seconds: 5));
                      _loadContactGroups();
                      setDlgState(() {});
                    }),
                );
              },
            ),
          ),
      ])),
      actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
    )));
  }

  void _showQRScanner() async {
    try {
      final client = http.Client();
      final resp = await client.get(Uri.parse('$_apiBase/api/chat/address')).timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200 && mounted) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        final link = data['link'] as String? ?? '';
        showDialog(context: context, builder: (ctx) => AlertDialog(
          backgroundColor: Colors.grey[900], title: const Text('Connect via SimpleX', style: TextStyle(fontSize: 20)),
          content: Column(mainAxisSize: MainAxisSize.min, children: [
            Text('Share this link or scan QR:', style: TextStyle(fontSize: 16, color: Colors.grey[400])),
            const SizedBox(height: 8),
            SelectableText(link, style: const TextStyle(fontSize: 14, color: Colors.cyan)),
            const SizedBox(height: 8),
            Text('Open SimpleX Chat → Add Contact → Scan QR (QR tab)', style: TextStyle(fontSize: 14, color: Colors.grey[500])),
          ]),
          actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
        ));
      }
    } catch (_) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('QR scanner: camera not available on desktop'), backgroundColor: Colors.amber));
    }
  }

  Widget _buildSetupGuide(ThemeData theme) {
    return Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(32),
        child: Column(mainAxisSize: MainAxisSize.min, children: [
          Icon(Icons.chat, size: 64, color: theme.colorScheme.primary),
          const SizedBox(height: 24),
          Text('Simplex Chat', style: theme.textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.bold, fontSize: 28)),
          const SizedBox(height: 16),
          Card(child: Padding(padding: const EdgeInsets.all(20), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text('How to use:', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold, fontSize: 22)),
            const SizedBox(height: 12),
            _step(theme, '1', 'Open SimpleX Chat on your phone'),
            _step(theme, '2', 'Tap "Add contact" → "Scan QR code"'),
            _step(theme, '3', 'Scan the QR from the QR tab'),
            _step(theme, '4', 'Send /help to Island Bot'),
            const SizedBox(height: 16),
            Text('Bot commands:', style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.bold, fontSize: 18)),
            const SizedBox(height: 8),
            _command(theme, '/wallet <pubkey>', 'check balance'),
            _command(theme, '/pay <from> <to> <amount>', 'send ng'),
            _command(theme, '/pos create <merchant> <amount>', 'create invoice'),
            _command(theme, '/mining summary', 'mining status'),
          ]))),
          const SizedBox(height: 16),
          OutlinedButton.icon(onPressed: _loadEverything, icon: const Icon(Icons.refresh), label: const Text('Check connection', style: TextStyle(fontSize: 18))),
        ]),
      ),
    );
  }

  Widget _step(ThemeData theme, String num, String text) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Container(width: 28, height: 28, decoration: BoxDecoration(color: theme.colorScheme.primary, shape: BoxShape.circle),
          child: Center(child: Text(num, style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold)))),
        const SizedBox(width: 10),
        Expanded(child: Text(text, style: const TextStyle(fontSize: 18))),
      ]),
    );
  }

  Widget _command(ThemeData theme, String cmd, String desc) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(children: [
        Container(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
          decoration: BoxDecoration(color: theme.colorScheme.surfaceContainerHighest, borderRadius: BorderRadius.circular(4)),
          child: Text(cmd, style: const TextStyle(fontFamily: 'monospace', fontSize: 16))),
        const SizedBox(width: 8),
        Text(desc, style: TextStyle(fontSize: 16, color: Colors.grey[400])),
      ]),
    );
  }

  Widget _buildContactHeader() {
    final contactStatus = _contactStatuses[_selectedContactId] ?? 'offline';
    final trustLevel = _contactTrust[_selectedContactId];
    final soundName = _contactSounds[_selectedContactId] ?? 'default';
    return Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Row(children: [
        IconButton(icon: const Icon(Icons.arrow_back, size: 22), onPressed: () => setState(() => _selectedContactId = ''),
          padding: EdgeInsets.zero, constraints: const BoxConstraints()),
        const SizedBox(width: 8),
        _statusIndicator(contactStatus),
        Text(_messages.where((m) => m.chatId == _selectedContactId).firstOrNull?.from ?? _selectedContactId,
          style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
        if (trustLevel != null && trustLevel != 'none') ...[
          const SizedBox(width: 6),
          _trustBadge(trustLevel),
        ],
        if (soundName != 'default') ...[
          const SizedBox(width: 4),
          Icon(Icons.notifications, size: 16, color: Colors.grey[500]),
        ],
        const SizedBox(width: 6),
        E2EBadge(active: _currentMessages.any((m) => m.encryption == 'e2e')),
      ]),
      if (contactStatus == 'online')
        Padding(padding: const EdgeInsets.only(left: 40),
          child: Text('online', style: TextStyle(fontSize: 14, color: Colors.green[400]))),
      if (_typingContact != null)
        Padding(padding: const EdgeInsets.only(left: 40),
          child: Text(_typingContact!, style: TextStyle(fontSize: 14, color: Colors.green[400], fontStyle: FontStyle.italic))),
    ]);
  }

  Widget _buildChatView(ThemeData theme) {
    final msgs = _currentMessages;
    return Column(children: [
      Container(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
        decoration: BoxDecoration(color: Colors.grey.shade900, border: Border(bottom: BorderSide(color: Colors.grey.shade800))),
        child: Row(children: [
          if (_selectedContactId.isNotEmpty)
            _buildContactHeader(),
          if (_selectedContactId.isEmpty) ...[
            Expanded(
              child: TextField(
                controller: _searchCtrl, style: const TextStyle(fontSize: 16, color: Colors.white),
                decoration: InputDecoration(hintText: 'Search messages...', hintStyle: TextStyle(fontSize: 16, color: Colors.grey[600]),
                  border: InputBorder.none, isDense: true, contentPadding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4)),
                onChanged: (v) => setState(() => _searching = v.isNotEmpty),
              ),
            ),
            if (_searching)
              IconButton(icon: const Icon(Icons.clear, size: 20), onPressed: () { setState(() { _searching = false; _searchCtrl.clear(); }); },
                padding: EdgeInsets.zero, constraints: const BoxConstraints()),
            if (!_searching && _conversations.isNotEmpty)
              Padding(padding: const EdgeInsets.only(right: 4),
                child: DropdownButton<String>(
                  value: _selectedContactId, hint: const Text('Filter', style: TextStyle(fontSize: 16)),
                  items: [
                    const DropdownMenuItem(value: '', child: Text('All', style: TextStyle(fontSize: 16))),
                    ..._conversations.map((m) => DropdownMenuItem(value: m.chatId, child: Text(m.from.isNotEmpty ? m.from : m.chatId, style: const TextStyle(fontSize: 16)))),
                  ],
                  onChanged: (v) => setState(() => _selectedContactId = v ?? ''),
                ),
              ),
          ],
        ]),
      ),
      Expanded(
          child: _loading
            ? const Center(child: CircularProgressIndicator())
            : msgs.isEmpty
                ? Center(child: Column(mainAxisSize: MainAxisSize.min, children: [
                    Icon(Icons.chat_bubble_outline, size: 64, color: Colors.grey[700]),
                    const SizedBox(height: 12),
                    Text('No messages yet', style: TextStyle(fontSize: 18, color: Colors.grey[500])),
                    Text('Connect via SimpleX Chat app to start', style: TextStyle(fontSize: 16, color: Colors.grey[600])),
                  ]))
                : ListView.builder(controller: _scrollCtrl, padding: const EdgeInsets.all(12),
                    itemCount: msgs.length, itemBuilder: (ctx, i) => _messageBubble(msgs[i], theme)),
      ),
      if (_replyingTo != null)
        Container(padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
          decoration: BoxDecoration(color: Colors.grey.shade900, border: Border(top: BorderSide(color: Colors.grey.shade800))),
          child: Row(children: [
            Expanded(child: Text('Reply: $_replyingText', style: TextStyle(fontSize: 16, color: Colors.grey[500]), maxLines: 1, overflow: TextOverflow.ellipsis)),
            GestureDetector(onTap: () => setState(() { _replyingTo = null; _replyingText = null; }),
              child: Icon(Icons.close, size: 18, color: Colors.grey[500])),
          ])),
      // Bulk mode bar
      if (_bulkMode)
        Container(padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          color: Colors.amber.shade900.withValues(alpha: 0.3),
          child: Row(children: [
            Text('${_selectedBulkIds.length} selected', style: TextStyle(fontSize: 16, color: Colors.amber[300])),
            const Spacer(),
            TextButton(onPressed: _showBulkDeleteConfirm, child: const Text('Delete', style: TextStyle(fontSize: 16, color: Colors.red))),
            TextButton(onPressed: () => setState(() { _bulkMode = false; _selectedBulkIds.clear(); }),
              child: const Text('Cancel', style: TextStyle(fontSize: 16))),
          ])),
      // AI suggestions bar
      if (_showSuggestions && _suggestions.isNotEmpty)
        Container(
          height: 40,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            color: Colors.purple.shade900.withValues(alpha: 0.2),
            border: Border(top: BorderSide(color: Colors.grey.shade800)),
          ),
          child: ListView.builder(
            scrollDirection: Axis.horizontal,
            itemCount: _suggestions.length,
            itemBuilder: (_, i) => GestureDetector(
              onTap: () => _sendSuggestion(_suggestions[i]),
              child: Container(
                margin: const EdgeInsets.symmetric(horizontal: 3, vertical: 6),
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.purple.shade800.withValues(alpha: 0.5),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(_suggestions[i], style: TextStyle(fontSize: 16, color: Colors.purple[200])),
              ),
            ),
          ),
        ),
      if (_templates.isNotEmpty && _showTemplatePicker)
        Container(height: 36, padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(color: Colors.grey.shade900, border: Border(top: BorderSide(color: Colors.grey.shade800))),
          child: ListView.builder(scrollDirection: Axis.horizontal, itemCount: _templates.length,
            itemBuilder: (_, i) => GestureDetector(
              onTap: () { _inputCtrl.text = _templates[i].text; setState(() => _showTemplatePicker = false); _saveDraft(); },
              child: Container(margin: const EdgeInsets.symmetric(horizontal: 3, vertical: 4),
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(color: Colors.grey.shade800, borderRadius: BorderRadius.circular(8)),
                child: Text(_templates[i].name, style: TextStyle(fontSize: 18, color: Colors.grey[300]))),
            ),
          ),
      ),
      Container(padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(border: Border(top: BorderSide(color: Colors.grey.shade800))),
        child: Column(mainAxisSize: MainAxisSize.min, children: [
          // Action row: templates + wallet + pay + voice + file + AI + schedule + invoice + send
          Row(children: [
            if (_templates.isNotEmpty)
              GestureDetector(onTap: () => setState(() => _showTemplatePicker = !_showTemplatePicker),
                child: Icon(Icons.bookmark, size: 28, color: _showTemplatePicker ? Colors.amber : Colors.grey)),
            // Quick wallet
            if (_selectedContactId.isNotEmpty)
              GestureDetector(onTap: _showQuickWallet,
                child: Padding(padding: const EdgeInsets.symmetric(horizontal: 2),
                  child: Icon(Icons.account_balance_wallet, size: 26, color: Colors.teal))),
            // Pay button
            if (_selectedContactId.isNotEmpty)
              GestureDetector(onTap: _showPayDialog,
                child: Padding(padding: const EdgeInsets.symmetric(horizontal: 2),
                  child: Icon(Icons.payments, size: 26, color: Colors.amber))),
            // Voice message
            if (_selectedContactId.isNotEmpty)
              GestureDetector(onTap: _showVoiceDialog,
                child: Padding(padding: const EdgeInsets.symmetric(horizontal: 2),
                  child: Icon(Icons.mic, size: 26, color: Colors.red[300]))),
            // File attachment
            if (_selectedContactId.isNotEmpty)
              GestureDetector(onTap: _showFileDialog,
                child: Padding(padding: const EdgeInsets.symmetric(horizontal: 2),
                  child: Icon(Icons.attach_file, size: 26, color: Colors.grey))),
            // AI Steward
            if (_selectedContactId.isNotEmpty)
              GestureDetector(onTap: _showAIDialog,
                child: Padding(padding: const EdgeInsets.symmetric(horizontal: 2),
                  child: Icon(Icons.smart_toy, size: 26, color: Colors.purple[300]))),
            // Suggest replies (AI context)
            if (_selectedContactId.isNotEmpty)
              GestureDetector(onTap: () { _loadSuggestions(); },
                child: Padding(padding: const EdgeInsets.symmetric(horizontal: 2),
                  child: Icon(Icons.lightbulb, size: 26, color: Colors.orange[300]))),
            // Media gallery
            GestureDetector(onTap: _showMediaGallery,
              child: Padding(padding: const EdgeInsets.symmetric(horizontal: 2),
                child: Icon(Icons.photo_library, size: 26, color: Colors.teal))),
            // Bulk mode toggle
            GestureDetector(onTap: _showBulkModeToggle,
              child: Padding(padding: const EdgeInsets.symmetric(horizontal: 2),
                child: Icon(Icons.checklist, size: 26, color: _bulkMode ? Colors.amber : Colors.grey))),
            Expanded(
              child: TextField(
                controller: _inputCtrl, style: const TextStyle(fontSize: 24, color: Colors.white),
                decoration: InputDecoration(
                  hintText: _scheduledTime != null ? 'Scheduled: ${_timeFmt.format(_scheduledTime!)}...'
                      : (_selectedContactId.isNotEmpty ? 'Message...' : 'Broadcast...'),
                  hintStyle: TextStyle(fontSize: 24, color: _scheduledTime != null ? Colors.amber[800] : Colors.grey[600]),
                  border: InputBorder.none, isDense: true, contentPadding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6)),
                onChanged: (_) { _saveDraft(); _sendTypingIndicator(); }, onSubmitted: (_) => _sendMessage(),
                maxLines: 3, minLines: 1,
              ),
            ),
            if (_selectedContactId.isNotEmpty)
              GestureDetector(onTap: _showScheduleDialog,
                child: Padding(padding: const EdgeInsets.symmetric(horizontal: 4),
                  child: Icon(Icons.schedule, size: 28, color: _scheduledTime != null ? Colors.amber : Colors.grey))),
            if (_selectedContactId.isNotEmpty)
              IconButton(icon: const Icon(Icons.receipt, size: 28, color: Colors.amber),
                onPressed: _showCreateInvoiceDialog, padding: EdgeInsets.zero, constraints: const BoxConstraints()),
            IconButton(icon: Icon(Icons.send, size: 32, color: theme.colorScheme.primary),
              onPressed: _sendMessage, padding: EdgeInsets.zero, constraints: const BoxConstraints()),
          ]),
          // Language + Theme row
          Row(children: [
            GestureDetector(onTap: _showLanguageDialog,
              child: Padding(padding: const EdgeInsets.symmetric(horizontal: 2),
                child: Icon(Icons.translate, size: 20, color: Colors.grey[500]))),
            GestureDetector(onTap: _showThemeDialog,
              child: Padding(padding: const EdgeInsets.symmetric(horizontal: 2),
                child: Icon(Icons.palette, size: 20, color: Colors.grey[500]))),
            if (_selectedContactId.isNotEmpty)
              GestureDetector(onTap: _showContactDiscoveryDialog,
                child: Padding(padding: const EdgeInsets.symmetric(horizontal: 2),
                  child: Icon(Icons.person_search, size: 20, color: Colors.grey[500]))),
            const Spacer(),
            Text('${msgs.length} msgs  |  ${_unreadCount} unread', style: TextStyle(fontSize: 12, color: Colors.grey[700])),
          ]),
        ]),
      ),
    ]);
  }

  void _playAudio(String url) async {
    final fullUrl = url.startsWith('http') ? url : '$_apiBase$url';
    try {
      final resp = await widget.httpClient.get(Uri.parse(fullUrl)).timeout(const Duration(seconds: 15));
      if (resp.statusCode == 200) {
        final tmpFile = '/tmp/simplex-audio-${DateTime.now().millisecondsSinceEpoch}.${url.endsWith('.wav') ? 'wav' : 'ogg'}';
        await File(tmpFile).writeAsBytes(resp.bodyBytes);
        await Process.run('gst-play-1.0', [tmpFile]);
        File(tmpFile).deleteSync();
      }
    } catch (_) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Audio playback failed'), backgroundColor: Colors.red));
    }
  }

  Widget _statusIcon(String status) {
    switch (status) {
      case 'sending': return SizedBox(width: 14, height: 14, child: CircularProgressIndicator(strokeWidth: 1.5, color: Colors.grey[500]));
      case 'failed': return Icon(Icons.error_outline, size: 24, color: Colors.red[400]);
      case 'delivered': return Icon(Icons.done_all, size: 24, color: Colors.blue[300]);
      default: return Icon(Icons.done, size: 24, color: Colors.grey[500]);
    }
  }

  Widget _messageBubble(ChatMessage msg, ThemeData theme) {
    final invId = msg.text.startsWith('[INVOICE]') ? msg.text.split('\n').firstWhere((l) => l.contains('ID:')).split('ID:').last.trim() : null;
    final isPinned = msg.pinned != null && msg.pinned!.isNotEmpty;
    final isSelected = _selectedBulkIds.contains(msg.id);
    // Extract URLs for link preview
    final urlMatch = RegExp(r'https?://[^\s]+').firstMatch(msg.text);
    final hasUrl = urlMatch != null;
    return Padding(
      padding: const EdgeInsets.only(bottom: 6),
      child: GestureDetector(
        onTap: _bulkMode
            ? () => setState(() { if (isSelected) _selectedBulkIds.remove(msg.id); else _selectedBulkIds.add(msg.id); })
            : null,
        onLongPress: _bulkMode
            ? () => setState(() { if (isSelected) _selectedBulkIds.remove(msg.id); else _selectedBulkIds.add(msg.id); })
            : (msg.isUser ? () => _showUserMessageActions(msg) : () => _showOtherMessageActions(msg)),
        child: Row(
          mainAxisAlignment: msg.isUser ? MainAxisAlignment.end : MainAxisAlignment.start,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            if (!msg.isUser) ...[
              CircleAvatar(radius: 16, backgroundColor: Colors.grey.shade700,
                child: Text(msg.from.isNotEmpty ? msg.from[0].toUpperCase() : '?', style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold))),
              const SizedBox(width: 6),
            ],
            Flexible(
              child: Container(
                constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.6),
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: isSelected ? Colors.amber.withValues(alpha: 0.3) : (isPinned ? Colors.amber.withValues(alpha: 0.15) : (msg.isUser ? theme.colorScheme.primary.withValues(alpha: 0.3) : Colors.grey.shade800)),
                  borderRadius: BorderRadius.circular(12),
                  border: isSelected ? Border.all(color: Colors.amber, width: 2) : null,
                ),
                child: Column(
                  crossAxisAlignment: msg.isUser ? CrossAxisAlignment.end : CrossAxisAlignment.start,
                  children: [
                    if (isPinned)
                      Row(mainAxisSize: MainAxisSize.min, children: [
                        Icon(Icons.push_pin, size: 14, color: Colors.amber),
                        const SizedBox(width: 4),
                        Text('Pinned', style: TextStyle(fontSize: 14, color: Colors.amber, fontWeight: FontWeight.bold)),
                      ]),
                    if (!msg.isUser && msg.from.isNotEmpty)
                      Padding(padding: const EdgeInsets.only(bottom: 3),
                        child: Text(msg.from, style: TextStyle(fontSize: 14, color: Colors.grey[500], fontWeight: FontWeight.bold))),
                    if (msg.replyText != null && msg.replyText!.isNotEmpty)
                      Container(
                        margin: const EdgeInsets.only(bottom: 4),
                        padding: const EdgeInsets.all(6),
                        decoration: BoxDecoration(
                          color: Colors.grey.shade900,
                          borderRadius: BorderRadius.circular(6),
                          border: Border(left: BorderSide(color: theme.colorScheme.primary, width: 2)),
                        ),
                        child: Text(msg.replyText!, style: TextStyle(fontSize: 16, color: Colors.grey[500]), maxLines: 2, overflow: TextOverflow.ellipsis),
                      ),
                    // Recalled message
                    if (msg.recalled)
                      Row(children: [
                        Icon(Icons.block, size: 16, color: Colors.red[400]),
                        const SizedBox(width: 4),
                        Text('Message recalled', style: TextStyle(fontSize: 18, color: Colors.red[400], fontStyle: FontStyle.italic)),
                      ])
                    else ...[
                      // Money transfer display
                      if (msg.moneyAmount != null && msg.moneyAmount! > 0)
                        Container(
                          margin: const EdgeInsets.only(bottom: 4),
                          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                          decoration: BoxDecoration(
                            color: Colors.amber.shade900.withValues(alpha: 0.3),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Row(mainAxisSize: MainAxisSize.min, children: [
                            const Icon(Icons.payments, size: 20, color: Colors.amber),
                            const SizedBox(width: 6),
                            Text('${msg.moneyAmount!.toStringAsFixed(2)} ${msg.moneyAsset ?? "XAG"}',
                              style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Colors.amber)),
                            if (msg.moneyTxId != null)
                              Padding(padding: const EdgeInsets.only(left: 6),
                                child: Icon(Icons.check_circle, size: 16, color: Colors.green)),
                          ]),
                        ),
                      // Voice message
                      if (msg.voiceUrl != null)
                        GestureDetector(
                          onTap: () => _playAudio(msg.voiceUrl!),
                          child: Container(
                            margin: const EdgeInsets.only(bottom: 4),
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                            decoration: BoxDecoration(color: Colors.red.shade900.withValues(alpha: 0.3), borderRadius: BorderRadius.circular(8)),
                            child: Row(mainAxisSize: MainAxisSize.min, children: [
                              Icon(Icons.play_circle, size: 28, color: Colors.red[300]),
                              const SizedBox(width: 6),
                              Text('🎤 ${msg.voiceDuration ?? 0}s', style: TextStyle(fontSize: 16, color: Colors.grey[300])),
                            ]),
                          ),
                        ),
                      // File attachment
                      if (msg.fileName != null)
                        Container(
                          margin: const EdgeInsets.only(bottom: 4),
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                          decoration: BoxDecoration(color: Colors.blue.shade900.withValues(alpha: 0.3), borderRadius: BorderRadius.circular(8)),
                          child: Row(mainAxisSize: MainAxisSize.min, children: [
                            const Icon(Icons.attach_file, size: 20, color: Colors.blue),
                            const SizedBox(width: 6),
                            Flexible(child: Text(msg.fileName!, style: TextStyle(fontSize: 16, color: Colors.grey[300]), overflow: TextOverflow.ellipsis)),
                            if (msg.fileSize != null && msg.fileSize! > 0)
                              Text('  (${(msg.fileSize! / 1024).toStringAsFixed(0)} KB)', style: TextStyle(fontSize: 14, color: Colors.grey[500])),
                          ]),
                        ),
                      _highlightText(msg.text, style: TextStyle(fontSize: 20, color: msg.isUser ? Colors.greenAccent : Colors.grey[300], height: 1.3)),
                      // Link preview inline (first URL)
                      if (hasUrl && !msg.text.startsWith('http') && !_urlsPreviewed.contains(msg.id))
                        Padding(padding: const EdgeInsets.only(top: 4),
                          child: GestureDetector(
                            onTap: () { _urlsPreviewed.add(msg.id); _fetchLinkPreview(urlMatch!.group(0)!); },
                            child: Container(
                              padding: const EdgeInsets.all(6),
                              decoration: BoxDecoration(color: Colors.grey.shade900, borderRadius: BorderRadius.circular(6)),
                              child: Row(mainAxisSize: MainAxisSize.min, children: [
                                Icon(Icons.link, size: 16, color: Colors.blue[300]),
                                const SizedBox(width: 4),
                                Text('Preview link', style: TextStyle(fontSize: 14, color: Colors.blue[300])),
                              ]),
                            ),
                          )),
                    ],
                    if (msg.reactions != null && msg.reactions!.isNotEmpty)
                      Padding(
                        padding: const EdgeInsets.only(top: 4),
                        child: Wrap(
                          spacing: 4,
                          children: msg.reactions!.entries.map((e) => Container(
                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                            decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(8)),
                            child: Text('${e.key}', style: const TextStyle(fontSize: 28)),
                          )).toList(),
                        ),
                      ),
                    Padding(
                      padding: EdgeInsets.only(top: 2, left: invId != null ? 12 : 0, right: invId != null ? 12 : 0, bottom: invId != null ? 8 : 0),
                      child: Row(mainAxisSize: MainAxisSize.min, children: [
                        // Encryption indicator
                        if (msg.encryption == 'e2e')
                          Padding(padding: const EdgeInsets.only(right: 4),
                            child: Icon(Icons.lock, size: 14, color: Colors.green[400]))
                        else if (msg.encryption == 'none')
                          Padding(padding: const EdgeInsets.only(right: 4),
                            child: Icon(Icons.lock_open, size: 14, color: Colors.red[400])),
                        // ParanoidX shield
                        if (_pxV2Ray || _pxTor || _pxSimplex)
                          Padding(padding: const EdgeInsets.only(right: 4),
                            child: Icon(Icons.shield, size: 14, color: (_pxV2Ray && _pxSimplex) ? Colors.green : Colors.orange)),
                        // Read receipt
                        if (msg.readAt != null && msg.readAt!.isNotEmpty)
                          Padding(padding: const EdgeInsets.only(right: 4),
                            child: Icon(Icons.visibility, size: 14, color: Colors.blue[300])),
                        Text(msg.displayTime, style: TextStyle(fontSize: 16, color: Colors.grey[600])),
                        if (msg.isUser) ...[const SizedBox(width: 4), _statusIcon(msg.status)],
                      ]),
                    ),
                  ],
                ),
              ),
            ),
            if (msg.isUser) ...[
              const SizedBox(width: 6),
              CircleAvatar(radius: 16, backgroundColor: theme.colorScheme.primary, child: const Icon(Icons.person, size: 20, color: Colors.white)),
            ],
          ],
        ),
      ),
    );
  }

  void _showUserMessageActions(ChatMessage msg) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.grey[900],
      builder: (ctx) => SafeArea(child: Column(mainAxisSize: MainAxisSize.min, children: [
        ListTile(
          leading: Icon(Icons.checklist, size: 24, color: _selectedBulkIds.contains(msg.id) ? Colors.amber : null),
          title: Text(_selectedBulkIds.contains(msg.id) ? 'Deselect' : 'Select', style: const TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); setState(() { if (_selectedBulkIds.contains(msg.id)) _selectedBulkIds.remove(msg.id); else _selectedBulkIds.add(msg.id); if (!_bulkMode) _bulkMode = true; }); },
        ),
        if (_selectedContactId.isNotEmpty) ListTile(
          leading: const Icon(Icons.reply, size: 24),
          title: const Text('Reply', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); setState(() { _replyingTo = msg.id; _replyingText = msg.text; }); },
        ),
        ListTile(
          leading: const Icon(Icons.translate, size: 24),
          title: const Text('Translate', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _translateText(msg.text); },
        ),
        if (msg.text.startsWith('http'))
        ListTile(
          leading: const Icon(Icons.preview, size: 24),
          title: const Text('Link Preview', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _fetchLinkPreview(msg.text); },
        ),
        ListTile(
          leading: const Icon(Icons.edit, size: 24),
          title: const Text('Edit', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _editMessage(msg); },
        ),
        ListTile(
          leading: Icon(Icons.push_pin, size: 24, color: msg.pinned != null && msg.pinned!.isNotEmpty ? Colors.amber : null),
          title: Text(msg.pinned != null && msg.pinned!.isNotEmpty ? 'Unpin' : 'Pin', style: const TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _togglePin(msg); },
        ),
        ListTile(
          leading: const Icon(Icons.emoji_emotions, size: 24),
          title: const Text('React', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _showReactionPicker(msg); },
        ),
        ListTile(
          leading: const Icon(Icons.forward, size: 24),
          title: const Text('Forward', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _showForwardDialog(msg); },
        ),
        ListTile(
          leading: Icon(Icons.delete, size: 24, color: Colors.red[400]),
          title: Text('Delete', style: TextStyle(fontSize: 18, color: Colors.red[400])),
          onTap: () { Navigator.pop(ctx); _deleteMessage(msg); },
        ),
      ])),
    );
  }

  void _showOtherMessageActions(ChatMessage msg) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.grey[900],
      builder: (ctx) => SafeArea(child: Column(mainAxisSize: MainAxisSize.min, children: [
        ListTile(
          leading: Icon(Icons.checklist, size: 24, color: _selectedBulkIds.contains(msg.id) ? Colors.amber : null),
          title: Text(_selectedBulkIds.contains(msg.id) ? 'Deselect' : 'Select', style: const TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); setState(() { if (_selectedBulkIds.contains(msg.id)) _selectedBulkIds.remove(msg.id); else _selectedBulkIds.add(msg.id); if (!_bulkMode) _bulkMode = true; }); },
        ),
        if (_selectedContactId.isNotEmpty) ListTile(
          leading: const Icon(Icons.reply, size: 24),
          title: const Text('Reply', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); setState(() { _replyingTo = msg.id; _replyingText = msg.text; }); },
        ),
        ListTile(
          leading: const Icon(Icons.translate, size: 24),
          title: const Text('Translate', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _translateText(msg.text); },
        ),
        if (msg.text.startsWith('http'))
        ListTile(
          leading: const Icon(Icons.preview, size: 24),
          title: const Text('Link Preview', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _fetchLinkPreview(msg.text); },
        ),
        ListTile(
          leading: Icon(Icons.push_pin, size: 24, color: msg.pinned != null && msg.pinned!.isNotEmpty ? Colors.amber : null),
          title: Text(msg.pinned != null && msg.pinned!.isNotEmpty ? 'Unpin' : 'Pin', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _togglePin(msg); },
        ),
        ListTile(
          leading: const Icon(Icons.translate, size: 24),
          title: const Text('Translate', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _translateText(msg.text); },
        ),
        if (msg.text.startsWith('http'))
        ListTile(
          leading: const Icon(Icons.preview, size: 24),
          title: const Text('Link Preview', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _fetchLinkPreview(msg.text); },
        ),
        ListTile(
          leading: Icon(Icons.push_pin, size: 24, color: msg.pinned != null && msg.pinned!.isNotEmpty ? Colors.amber : null),
          title: Text(msg.pinned != null && msg.pinned!.isNotEmpty ? 'Unpin' : 'Pin', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _togglePin(msg); },
        ),
        ListTile(
          leading: const Icon(Icons.emoji_emotions, size: 24),
          title: const Text('React', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _showReactionPicker(msg); },
        ),
        ListTile(
          leading: const Icon(Icons.forward, size: 24),
          title: const Text('Forward', style: TextStyle(fontSize: 18)),
          onTap: () { Navigator.pop(ctx); _showForwardDialog(msg); },
        ),
        if (msg.recalled == false)
        ListTile(
          leading: Icon(Icons.undo, size: 24, color: Colors.red[300]),
          title: Text('Recall', style: TextStyle(fontSize: 18, color: Colors.red[300])),
          onTap: () { Navigator.pop(ctx); _recallMessage(msg); },
        ),
      ])),
    );
  }

  Future<void> _fetchLinkPreview(String url) async {
    try {
      final resp = await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/link-preview'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'url': url}),
      ).timeout(const Duration(seconds: 10));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['ok'] == true && data['preview'] != null) {
          final p = data['preview'] as Map<String, dynamic>;
          if (mounted) showDialog(context: context, builder: (ctx) => AlertDialog(
            backgroundColor: Colors.grey[900],
            title: Text(p['title'] as String? ?? 'Link Preview', style: const TextStyle(fontSize: 20)),
            content: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
              if (p['image'] != null && (p['image'] as String).isNotEmpty)
                ClipRRect(borderRadius: BorderRadius.circular(8),
                  child: Image.network(p['image'] as String, height: 150, width: double.infinity, fit: BoxFit.cover,
                    errorBuilder: (_, __, ___) => const SizedBox.shrink())),
              if (p['description'] != null) Padding(padding: const EdgeInsets.only(top: 8),
                child: Text(p['description'] as String, style: TextStyle(fontSize: 16, color: Colors.grey[400]))),
              if (p['site_name'] != null) Padding(padding: const EdgeInsets.only(top: 4),
                child: Text(p['site_name'] as String, style: TextStyle(fontSize: 14, color: Colors.grey[600]))),
            ]),
            actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
          ));
        }
      }
    } catch (_) {}
  }

  // ============ NEW FEATURES (Cycles 1-20) ============

  void _showQuickWallet() {
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: const Text('Quick Wallet', style: TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Icon(Icons.account_balance, color: Colors.teal, size: 28),
          const SizedBox(width: 8),
          Text('Balance: $_walletBalance XAG', style: const TextStyle(fontSize: 20, color: Colors.white)),
        ]),
        const SizedBox(height: 12),
        SizedBox(width: double.infinity, child: ElevatedButton.icon(
          onPressed: () { Navigator.pop(ctx); _showPayDialog(); },
          icon: const Icon(Icons.send, size: 20),
          label: const Text('Send XAG', style: TextStyle(fontSize: 16)),
          style: ElevatedButton.styleFrom(backgroundColor: Colors.amber.shade800),
        )),
        const SizedBox(height: 6),
        SizedBox(width: double.infinity, child: ElevatedButton.icon(
          onPressed: () { Navigator.pop(ctx); _showCreateInvoiceDialog(); },
          icon: const Icon(Icons.receipt, size: 20),
          label: const Text('Create Invoice', style: TextStyle(fontSize: 16)),
          style: ElevatedButton.styleFrom(backgroundColor: Colors.teal.shade800),
        )),
      ]),
      actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 16)))],
    ));
  }

  void _showPayDialog() {
    final ctrl = TextEditingController();
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: Row(children: [
        Icon(Icons.payments, color: Colors.amber, size: 24),
        const SizedBox(width: 8),
        const Text('Send XAG', style: TextStyle(fontSize: 20)),
      ]),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        TextField(controller: ctrl, keyboardType: TextInputType.number,
          style: const TextStyle(fontSize: 24, color: Colors.white),
          decoration: InputDecoration(labelText: 'Amount (XAG)', labelStyle: const TextStyle(fontSize: 18),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
            filled: true, fillColor: Colors.grey.shade800)),
      ]),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 16))),
        ElevatedButton(onPressed: () async {
          final amt = double.tryParse(ctrl.text);
          if (amt == null || amt <= 0) return;
          Navigator.pop(ctx);
          final contactId = _selectedContactId.isNotEmpty ? int.tryParse(_selectedContactId.replaceAll(RegExp(r'[^0-9]'), '')) ?? 0 : 0;
          try {
            await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/pay'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'contact_id': contactId, 'chat_id': _selectedContactId, 'amount': amt, 'asset': 'XAG'}))
              .timeout(const Duration(seconds: 15));
            _loadHistory();
          } catch (_) {}
        }, child: const Text('Send 💸', style: TextStyle(fontSize: 16))),
      ],
    ));
  }

  void _showVoiceDialog() {
    // Voice recording — on Linux desktop, pipe capture via arecord/sox
    // For now, show a dialog with a record button placeholder
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: const Text('Voice Message', style: TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        Icon(Icons.mic, size: 64, color: Colors.red[300]),
        const SizedBox(height: 12),
        const Text('Hold to record', style: TextStyle(fontSize: 18, color: Colors.white)),
        const SizedBox(height: 8),
        Text('Linux audio capture via arecord/sox\n(platform channel needed for native recording)',
          style: TextStyle(fontSize: 14, color: Colors.grey[500]), textAlign: TextAlign.center),
      ]),
      actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Close', style: TextStyle(fontSize: 18)))],
    ));
    if (mounted) ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('🎤 Voice recording uses arecord — see scripts/record-voice.sh'), backgroundColor: Colors.blue));
  }

  void _showFileDialog() async {
    try {
      final result = await FilePicker.platform.pickFiles(type: FileType.any);
      if (result == null || result.files.single.path == null) return;
      final file = result.files.single;
      final filePath = file.path!;
      final fileName = file.name;
      final fileSize = file.size;
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Uploading $fileName (${fileSize ~/ 1024}KB)...'), backgroundColor: Colors.blue));
      final req = http.MultipartRequest('POST', Uri.parse('$_apiBase/api/chat/send-file'));
      req.fields['chat_id'] = _selectedContactId;
      req.files.add(await http.MultipartFile.fromPath('file', filePath, filename: fileName));
      final resp = await req.send();
      if (resp.statusCode == 200 && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('File sent'), backgroundColor: Colors.green));
        _loadHistory();
      } else if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Upload failed'), backgroundColor: Colors.red));
      }
    } catch (_) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Upload failed'), backgroundColor: Colors.red));
    }
  }

  void _showAIDialog() {
    final ctrl = TextEditingController();
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: Row(children: [
        Icon(Icons.smart_toy, color: Colors.purple[300], size: 24),
        const SizedBox(width: 8),
        const Text('Ask Steward AI', style: TextStyle(fontSize: 20)),
      ]),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        TextField(controller: ctrl, maxLines: 3,
          style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(labelText: 'Ask anything...', labelStyle: const TextStyle(fontSize: 18),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
            filled: true, fillColor: Colors.grey.shade800)),
        if (_stewardThinking.isNotEmpty)
          Padding(padding: const EdgeInsets.only(top: 8),
            child: Text(_stewardThinking, style: TextStyle(fontSize: 14, color: Colors.purple[200], fontStyle: FontStyle.italic)),
          ),
      ]),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 16))),
        ElevatedButton(onPressed: () async {
          final text = ctrl.text.trim();
          if (text.isEmpty) return;
          setState(() => _stewardThinking = '🤖 Steward is thinking...');
          final contactId = _selectedContactId.isNotEmpty ? int.tryParse(_selectedContactId.replaceAll(RegExp(r'[^0-9]'), '')) ?? 0 : 0;
          try {
            await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/ai'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'text': text, 'chat_id': _selectedContactId, 'contact_id': contactId}))
              .timeout(const Duration(seconds: 30));
            setState(() => _stewardThinking = '');
            Navigator.pop(ctx);
            _loadHistory();
          } catch (_) {
            setState(() => _stewardThinking = 'AI request failed');
          }
        }, child: const Text('Ask AI 🤖', style: TextStyle(fontSize: 16))),
      ],
    ));
  }

  void _showLanguageDialog() {
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: const Text('Language', style: TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        _langOption(ctx, 'en', 'English 🇬🇧'),
        _langOption(ctx, 'ru', 'Русский 🇷🇺'),
        _langOption(ctx, 'es', 'Español 🇪🇸'),
      ]),
    ));
  }

  Widget _langOption(BuildContext ctx, String code, String label) {
    return ListTile(
      leading: Icon(Icons.check, size: 22, color: _activeLanguage == code ? Colors.cyan : Colors.transparent),
      title: Text(label, style: TextStyle(fontSize: 18, color: Colors.white)),
      onTap: () async {
        setState(() => _activeLanguage = code);
        try {
          await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/language'),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({'language': code})).timeout(const Duration(seconds: 5));
        } catch (_) {}
        Navigator.pop(ctx);
      },
    );
  }

  void _showThemeDialog() {
    final colorOptions = [
      ('default', Colors.cyan, 'Cyan'),
      ('teal', Colors.teal, 'Teal'),
      ('amber', Colors.amber, 'Amber'),
      ('purple', Colors.purple, 'Purple'),
      ('red', Colors.red, 'Red'),
      ('blue', Colors.blue, 'Blue'),
      ('green', Colors.green, 'Green'),
      ('pink', Colors.pink, 'Pink'),
      ('orange', Colors.deepOrange, 'Orange'),
      ('indigo', Colors.indigo, 'Indigo'),
    ];
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: const Text('Chat Theme', style: TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        Text('Current: ${_chatThemeColor}', style: TextStyle(fontSize: 16, color: Colors.grey[400])),
        const SizedBox(height: 8),
        Wrap(spacing: 12, runSpacing: 8,
          children: colorOptions.map((c) => GestureDetector(
            onTap: () async {
              setState(() => _chatThemeColor = c.$1);
              try {
                await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/theme'),
                  headers: {'Content-Type': 'application/json'},
                  body: jsonEncode({'chat_id': _selectedContactId.isNotEmpty ? _selectedContactId : 'global', 'theme': c.$1}))
                  .timeout(const Duration(seconds: 5));
              } catch (_) {}
              Navigator.pop(ctx);
            },
            child: Container(
              width: 56, height: 56,
              decoration: BoxDecoration(color: c.$2, borderRadius: BorderRadius.circular(12),
                border: _chatThemeColor == c.$1 ? Border.all(color: Colors.white, width: 3) : null),
              child: _chatThemeColor == c.$1
                  ? const Icon(Icons.check, color: Colors.white, size: 28)
                  : null,
            ),
          )).toList(),
        ),
        const SizedBox(height: 8),
        Text('Tap a color to apply', style: TextStyle(fontSize: 14, color: Colors.grey[500])),
      ]),
    ));
  }

  void _showContactDiscoveryDialog() {
    final ctrl = TextEditingController();
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: const Text('Find Contact', style: TextStyle(fontSize: 20)),
      content: Column(mainAxisSize: MainAxisSize.min, children: [
        TextField(controller: ctrl,
          style: const TextStyle(fontSize: 18, color: Colors.white),
          decoration: InputDecoration(labelText: 'Alias or public key', labelStyle: const TextStyle(fontSize: 18),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
            filled: true, fillColor: Colors.grey.shade800)),
      ]),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 16))),
        ElevatedButton(onPressed: () async {
          Navigator.pop(ctx);
          try {
            final resp = await widget.httpClient.get(
              Uri.parse('$_apiBase/api/chat/contact?q=${Uri.encodeComponent(ctrl.text.trim())}'))
              .timeout(const Duration(seconds: 8));
            if (resp.statusCode == 200) {
              if (mounted) ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('Contact search completed'), backgroundColor: Colors.green));
              _loadContacts();
            }
          } catch (_) {
            if (mounted) ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('Contact search failed'), backgroundColor: Colors.red));
          }
        }, child: const Text('Search', style: TextStyle(fontSize: 16))),
      ],
    ));
  }

  void _recallMessage(ChatMessage msg) {
    showDialog(context: context, builder: (ctx) => AlertDialog(
      backgroundColor: Colors.grey[900],
      title: const Text('Recall Message', style: TextStyle(fontSize: 18, color: Colors.red)),
      content: const Text('This will remove the message for everyone. Continue?', style: TextStyle(fontSize: 16)),
      actions: [
        TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel', style: TextStyle(fontSize: 16))),
        ElevatedButton(onPressed: () async {
          Navigator.pop(ctx);
          try {
            await widget.httpClient.post(Uri.parse('$_apiBase/api/chat/recall'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode({'id': msg.id})).timeout(const Duration(seconds: 8));
            _loadHistory();
          } catch (_) {}
        }, child: const Text('Recall', style: TextStyle(fontSize: 16, color: Colors.red))),
      ],
    ));
  }

  Widget _highlightText(String text, {required TextStyle style}) {
    if (!_searching || _searchCtrl.text.trim().isEmpty) {
      return Text(text, style: style);
    }
    final q = _searchCtrl.text.trim();
    final lower = text.toLowerCase();
    final qLower = q.toLowerCase();
    final spans = <TextSpan>[];
    var start = 0;
    while (true) {
      final idx = lower.indexOf(qLower, start);
      if (idx < 0) {
        spans.add(TextSpan(text: text.substring(start), style: style));
        break;
      }
      if (idx > start) {
        spans.add(TextSpan(text: text.substring(start, idx), style: style));
      }
      spans.add(TextSpan(text: text.substring(idx, idx + q.length), style: style.copyWith(
        backgroundColor: Colors.amber.shade700, color: Colors.white)));
      start = idx + q.length;
    }
    return RichText(text: TextSpan(children: spans));
  }
}

extension StringCapitalize on String {
  String capitalizeFirst() {
    if (isEmpty) return this;
    return this[0].toUpperCase() + substring(1);
  }
}
