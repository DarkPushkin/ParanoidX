import 'dart:convert';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import 'package:http/http.dart' as http;
import 'package:path_provider/path_provider.dart';
import '../services/isle_api_service.dart';

/// VaultScreen manages the encrypted file vault with upload, download and container management.
class VaultScreen extends StatefulWidget {
  final IsleApiService api;
  const VaultScreen({super.key, required this.api});

  @override
  State<VaultScreen> createState() => _VaultScreenState();
}

class _VaultScreenState extends State<VaultScreen> {
  List<Map<String, dynamic>> _files = [];
  bool _loading = true;
  bool _uploading = false;
  double _uploadProgress = 0;
  double _usedMb = 0;
  double _quotaMb = 2048;

  // Container state
  List<Map<String, dynamic>> _containers = [];
  bool _containersLoading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final files = await widget.api.getVaultFiles();
      final usage = await widget.api.getVaultUsage();
      if (mounted) {
        setState(() {
          _files = files.cast<Map<String, dynamic>>();
          _usedMb = (usage['used_mb'] as num?)?.toDouble() ?? 0;
          _quotaMb = (usage['quota_mb'] as num?)?.toDouble() ?? 2048;
          _loading = false;
        });
      }
    } catch (_) {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _uploadFile() async {
    final result = await FilePicker.platform.pickFiles();
    if (result == null || result.files.isEmpty) return;
    final file = result.files.first;
    final path = file.path;
    if (path == null) return;
    setState(() { _uploading = true; _uploadProgress = 0; });
    try {
      await widget.api.uploadVaultFile(path, file.name);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Uploaded ${file.name}'), backgroundColor: Colors.green),
        );
      }
      await _load();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Upload failed: $e'), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _uploading = false);
    }
  }

  Future<void> _downloadFile(String name) async {
    try {
      final bytes = await widget.api.downloadVaultFile(name);
      final dir = (await getApplicationDocumentsDirectory()).path;
      await Directory(dir).create(recursive: true);
      final file = File('$dir/$name');
      await file.writeAsBytes(bytes);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Downloaded $name'), backgroundColor: Colors.green),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Download failed: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _deleteFile(String name) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Delete File'),
        content: Text('Permanently delete "$name"?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Cancel')),
          TextButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Delete', style: TextStyle(color: Colors.red))),
        ],
      ),
    );
    if (confirm != true) return;
    try {
      await widget.api.deleteVaultFile(name);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Deleted $name'), backgroundColor: Colors.orange),
        );
      }
      await _load();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Delete failed: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  IconData _iconForFile(String name) {
    final ext = name.split('.').last.toLowerCase();
    switch (ext) {
      case 'pdf': return Icons.picture_as_pdf;
      case 'jpg': case 'jpeg': case 'png': case 'gif': case 'webp': return Icons.image;
      case 'mp3': case 'wav': case 'ogg': case 'flac': return Icons.audiotrack;
      case 'mp4': case 'avi': case 'mkv': case 'mov': return Icons.videocam;
      case 'zip': case 'tar': case 'gz': case 'rar': case '7z': return Icons.folder_zip;
      case 'txt': case 'md': case 'json': case 'xml': case 'csv': return Icons.description;
      case 'doc': case 'docx': return Icons.article;
      default: return Icons.insert_drive_file;
    }
  }

  String _formatSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
  }

/// Returns the current  apiBase value.
  String get _apiBase => widget.api.client.baseUrl.toString().replaceAll('/api', '');

  Future<void> _loadContainers() async {
    setState(() => _containersLoading = true);
    try {
      final resp = await http.Client()
          .get(Uri.parse('$_apiBase/api/container/status'))
          .timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (mounted) setState(() => _containers = [data]);
      }
    } catch (_) {}
    try {
      final resp = await http.Client()
          .get(Uri.parse('$_apiBase/api/dc/list'))
          .timeout(const Duration(seconds: 5));
      if (resp.statusCode == 200) {
        final data = jsonDecode(resp.body) as Map<String, dynamic>;
        if (data['containers'] is List) {
          if (mounted) setState(() {
            _containers = (data['containers'] as List).map((e) => e as Map<String, dynamic>).toList();
          });
        }
      }
    } catch (_) {}
    if (mounted) setState(() => _containersLoading = false);
  }

  Future<void> _backupContainer() async {
    try {
      final resp = await http.Client()
          .post(Uri.parse('$_apiBase/api/container/backup'), headers: {'Content-Type': 'application/json'}, body: '{}')
          .timeout(const Duration(seconds: 30));
      if (resp.statusCode == 200 && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Container backup started'), backgroundColor: Colors.green));
        _loadContainers();
      }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Backup failed: $e'), backgroundColor: Colors.red));
    }
  }

  Future<void> _restoreContainer() async {
    final result = await FilePicker.platform.pickFiles();
    if (result == null || result.files.isEmpty) return;
    final file = result.files.first;
    final path = file.path;
    if (path == null) return;
    try {
      final uri = Uri.parse('$_apiBase/api/container/restore');
      final request = http.MultipartRequest('POST', uri);
      request.files.add(await http.MultipartFile.fromPath('file', path));
      final streamedResp = await request.send().timeout(const Duration(seconds: 30));
      final resp = await http.Response.fromStream(streamedResp);
      if (resp.statusCode == 200 && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Container restored'), backgroundColor: Colors.green));
        _loadContainers();
      }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Restore failed: $e'), backgroundColor: Colors.red));
    }
  }

  Future<void> _panicContainer() async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('PANIC!'),
        content: const Text('This will WIPE all container data. Are you sure?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Cancel')),
          TextButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('PANIC', style: TextStyle(color: Colors.red))),
        ],
      ),
    );
    if (confirm != true) return;
    try {
      final resp = await http.Client()
          .post(Uri.parse('$_apiBase/api/panic'), headers: {'Content-Type': 'application/json'}, body: '{}')
          .timeout(const Duration(seconds: 10));
      if (resp.statusCode == 200 && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('PANIC executed — container wiped'), backgroundColor: Colors.red));
      }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Panic failed: $e'), backgroundColor: Colors.red));
    }
  }

  Widget _buildContainerSection(ThemeData theme) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.grey.shade100,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey.shade300),
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Row(children: [
          Icon(Icons.inventory_2, size: 18, color: theme.colorScheme.primary),
          const SizedBox(width: 6),
          Text('Containers', style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: theme.colorScheme.primary)),
          const Spacer(),
          if (_containersLoading)
            const SizedBox(width: 14, height: 14, child: CircularProgressIndicator(strokeWidth: 2))
          else
            IconButton(icon: const Icon(Icons.refresh, size: 16), onPressed: _loadContainers, padding: EdgeInsets.zero, constraints: const BoxConstraints()),
          IconButton(icon: const Icon(Icons.backup, size: 16, color: Colors.green), onPressed: _backupContainer, tooltip: 'Backup', padding: EdgeInsets.zero, constraints: const BoxConstraints()),
          IconButton(icon: const Icon(Icons.restore, size: 16, color: Colors.blue), onPressed: _restoreContainer, tooltip: 'Restore', padding: EdgeInsets.zero, constraints: const BoxConstraints()),
          IconButton(icon: const Icon(Icons.warning, size: 16, color: Colors.red), onPressed: _panicContainer, tooltip: 'Panic', padding: EdgeInsets.zero, constraints: const BoxConstraints()),
        ]),
        if (_containers.isEmpty && !_containersLoading)
          Padding(
            padding: const EdgeInsets.only(top: 4),
            child: Text('No containers — tap refresh to check', style: TextStyle(fontSize: 12, color: Colors.grey[600])),
          ),
        if (_containers.isNotEmpty)
          ...(_containers.take(3).map((c) => Padding(
            padding: const EdgeInsets.only(top: 3),
            child: Row(children: [
              Icon(Icons.storage, size: 14, color: theme.colorScheme.primary),
              const SizedBox(width: 4),
              Text(c['name'] as String? ?? c['id'] as String? ?? 'container', style: const TextStyle(fontSize: 12)),
              const Spacer(),
              if (c['size'] != null)
                Text(_formatSize((c['size'] as num).toInt()), style: TextStyle(fontSize: 11, color: Colors.grey[600])),
              if (c['encrypted'] == true)
                Padding(
                  padding: const EdgeInsets.only(left: 4),
                  child: Icon(Icons.lock, size: 12, color: Colors.green),
                ),
              if (c['created'] != null)
                Text('  ${(c['created'] as String).substring(0, 10)}', style: TextStyle(fontSize: 11, color: Colors.grey[500])),
            ]),
          ))),
      ]),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final usagePct = _quotaMb > 0 ? (_usedMb / _quotaMb) : 0.0;
    final usageColor = usagePct > 0.9 ? Colors.red : (usagePct > 0.7 ? Colors.orange : Colors.green);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Vault'),
        actions: [
          if (_uploading)
            const Padding(
              padding: EdgeInsets.symmetric(horizontal: 12),
              child: SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2)),
            )
          else
            IconButton(
              icon: const Icon(Icons.upload_file, size: 20),
              tooltip: 'Upload File',
              onPressed: _uploadFile,
            ),
          IconButton(icon: const Icon(Icons.refresh, size: 20), onPressed: _load),
        ],
      ),
      body: Column(
        children: [
          _buildContainerSection(theme),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text('Storage', style: theme.textTheme.labelSmall),
                    const Spacer(),
                    Text('${_usedMb.toStringAsFixed(1)} MB / ${_quotaMb.toStringAsFixed(0)} MB',
                        style: TextStyle(fontSize: 11, color: Colors.grey)),
                  ],
                ),
                const SizedBox(height: 4),
                ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: LinearProgressIndicator(
                    value: usagePct.clamp(0.0, 1.0),
                    backgroundColor: Colors.grey.shade200,
                    valueColor: AlwaysStoppedAnimation(usageColor),
                    minHeight: 8,
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : _files.isEmpty
                    ? const Center(child: Text('No files — tap upload to add one'))
                    : RefreshIndicator(
                        onRefresh: _load,
                        child: ListView.builder(
                          padding: const EdgeInsets.symmetric(horizontal: 12),
                          itemCount: _files.length,
                          itemBuilder: (ctx, i) {
                            final f = _files[i];
                            final name = f['name'] as String? ?? 'unknown';
                            final size = f['size'] as int? ?? 0;
                            final mtime = f['mtime'] as String? ?? '';
                            return Card(
                              margin: const EdgeInsets.symmetric(vertical: 3),
                              child: ListTile(
                                leading: Icon(_iconForFile(name), color: theme.colorScheme.primary),
                                title: Text(name, overflow: TextOverflow.ellipsis),
                                subtitle: Text('${_formatSize(size)}  ${mtime.isNotEmpty ? mtime.substring(0, 10) : ''}',
                                    style: TextStyle(fontSize: 11, color: Colors.grey[600])),
                                trailing: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    IconButton(
                                      icon: const Icon(Icons.download, size: 20),
                                      tooltip: 'Download',
                                      onPressed: () => _downloadFile(name),
                                    ),
                                    IconButton(
                                      icon: const Icon(Icons.delete_outline, size: 20, color: Colors.red),
                                      tooltip: 'Delete',
                                      onPressed: () => _deleteFile(name),
                                    ),
                                  ],
                                ),
                              ),
                            );
                          },
                        ),
                      ),
          ),
        ],
      ),
    );
  }
}
