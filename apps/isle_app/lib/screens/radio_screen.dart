import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:models/models.dart';
import '../services/radio_player.dart';

/// RadioScreen manages the radio station browser and audio player interface.
class RadioScreen extends StatefulWidget {
  final String apiBase;
  final http.Client? httpClient;
  final RadioPlayer radioPlayer;

  const RadioScreen({
    super.key,
    required this.apiBase,
    this.httpClient,
    required this.radioPlayer,
  });

  @override
  State<RadioScreen> createState() => _RadioScreenState();
}

class _RadioScreenState extends State<RadioScreen> {
  @override
  void initState() {
    super.initState();
    if (widget.radioPlayer.stations.isEmpty) {
      widget.radioPlayer.loadStations();
    }
  }

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: widget.radioPlayer,
      builder: (context, _) {
        final rp = widget.radioPlayer;
        if (rp.loadingStations && rp.stations.isEmpty) {
          return const Scaffold(body: Center(child: CircularProgressIndicator()));
        }
        if (rp.error != null && rp.stations.isEmpty) {
          return Scaffold(
            appBar: AppBar(title: const Text('Radio')),
            body: Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.radio, size: 64, color: Colors.grey),
                  const SizedBox(height: 16),
                  Text('Could not load stations', style: Theme.of(context).textTheme.bodyLarge),
                  const SizedBox(height: 8),
                  Text(rp.error!, style: TextStyle(color: Colors.red[700], fontSize: 12)),
                  const SizedBox(height: 16),
                  ElevatedButton.icon(
                    icon: const Icon(Icons.refresh),
                    label: const Text('Retry'),
                    onPressed: rp.loadStations,
                  ),
                ],
              ),
            ),
          );
        }
        if (rp.selectedStation != null) {
          return _buildPlayerView();
        }
        return _buildStationList();
      },
    );
  }

  Widget _buildStationList() {
    final rp = widget.radioPlayer;
    return Scaffold(
      appBar: AppBar(
        title: const Text('The Island Radio'),
        actions: [
          if (rp.loadingStations)
            const Padding(
              padding: EdgeInsets.all(16),
              child: SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2)),
            ),
          IconButton(icon: const Icon(Icons.refresh), onPressed: rp.loadStations),
        ],
      ),
      body: rp.stations.isEmpty
          ? const Center(child: Text('No stations available'))
          : ListView.builder(
              itemCount: rp.stations.length,
              itemBuilder: (ctx, i) {
                final s = rp.stations[i];
                final isActive = rp.selectedStation?.id == s.id;
                return Card(
                  margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  color: isActive ? Colors.green.withValues(alpha: 0.08) : null,
                  child: ListTile(
                    leading: Text('${s.typeIcon}${s.languageFlag}', style: const TextStyle(fontSize: 28)),
                    title: Text(s.name, style: const TextStyle(fontWeight: FontWeight.bold)),
                    subtitle: Text('${s.languageLabel}  ${s.description}',
                        maxLines: 2, overflow: TextOverflow.ellipsis,
                        style: TextStyle(fontSize: 12, color: Colors.grey[600])),
                    trailing: s.enabled
                        ? Icon(isActive ? Icons.play_circle_fill : Icons.play_circle_fill,
                            color: isActive ? Colors.green : Colors.grey, size: 32)
                        : const Icon(Icons.stop, color: Colors.grey),
                    onTap: s.enabled ? () => widget.radioPlayer.selectStation(s) : null,
                  ),
                );
              },
            ),
    );
  }

  Widget _buildPlayerView() {
    final rp = widget.radioPlayer;
    final s = rp.selectedStation!;
    final currentTrack = rp.currentTrackIndex < rp.playlist.length ? rp.playlist[rp.currentTrackIndex] : null;

    return Scaffold(
      appBar: AppBar(
        title: Text('${s.typeIcon} ${s.name}'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => widget.radioPlayer.resetStation(),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => widget.radioPlayer.selectStation(s),
          ),
        ],
      ),
      body: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            color: rp.playing ? Colors.green.withValues(alpha: 0.1) : Theme.of(context).colorScheme.primaryContainer,
            child: Column(
              children: [
                Icon(Icons.radio, size: 36, color: rp.playing ? Colors.green : Colors.grey),
                const SizedBox(height: 6),
                if (currentTrack != null) ...[
                  Text(currentTrack.title, style: Theme.of(context).textTheme.titleSmall, textAlign: TextAlign.center),
                  const SizedBox(height: 2),
                  Text(currentTrack.typeLabel,
                      style: TextStyle(color: Colors.grey[600], fontSize: 11)),
                ] else
                  Text('Loading playlist...', style: TextStyle(color: Colors.grey[600], fontSize: 12)),
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    IconButton(
                      icon: const Icon(Icons.skip_previous, size: 20),
                      onPressed: rp.currentTrackIndex > 0 ? rp.skipPrevious : null,
                    ),
                    IconButton(
                      icon: Icon(rp.playing ? Icons.stop : Icons.play_arrow, size: 28),
                      color: rp.playing ? Colors.red : Colors.green,
                      onPressed: rp.togglePlayPause,
                    ),
                    IconButton(
                      icon: const Icon(Icons.skip_next, size: 20),
                      onPressed: rp.currentTrackIndex + 1 < rp.playlist.length ? rp.skipNext : null,
                    ),
                  ],
                ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 32),
                  child: Row(
                    children: [
                      const Icon(Icons.volume_down, size: 14),
                      Expanded(
                        child: Slider(
                          value: rp.volume,
                          onChanged: rp.setVolume,
                          min: 0.0,
                          max: 1.0,
                          divisions: 20,
                        ),
                      ),
                      const Icon(Icons.volume_up, size: 14),
                    ],
                  ),
                ),
                if (rp.playlist.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Text('Track ${rp.currentTrackIndex + 1} of ${rp.playlist.length}',
                        style: TextStyle(color: Colors.grey[600], fontSize: 10)),
                  ),
              ],
            ),
          ),
          Expanded(
            child: rp.loadingPlaylist
                ? const Center(child: CircularProgressIndicator())
                : rp.playlist.isEmpty
                    ? const Center(child: Text('No tracks — upload audio via API'))
                    : ListView.builder(
                        itemCount: rp.playlist.length,
                        itemBuilder: (ctx, i) {
                          final t = rp.playlist[i];
                          final isCurrent = i == rp.currentTrackIndex;
                          return Card(
                            margin: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                            color: isCurrent ? Colors.green.withValues(alpha: 0.08) : null,
                            child: ListTile(
                              leading: Icon(_iconForTrack(t),
                                  color: isCurrent ? Colors.green : _colorForTrack(t)),
                              title: Text(t.title, maxLines: 1, overflow: TextOverflow.ellipsis,
                                  style: TextStyle(
                                      fontWeight: isCurrent ? FontWeight.bold : FontWeight.normal)),
                              subtitle: Text('${t.typeLabel}  ${t.duration}s',
                                  style: TextStyle(fontSize: 11, color: Colors.grey[600])),
                              trailing: IconButton(
                                icon: Icon(Icons.play_arrow,
                                    color: isCurrent && rp.playing ? Colors.green : null),
                                tooltip: 'Play',
                                onPressed: () => widget.radioPlayer.playTrack(i),
                              ),
                            ),
                          );
                        },
                      ),
          ),
        ],
      ),
    );
  }

  IconData _iconForTrack(RadioTrack t) {
    if (t.isAd) return Icons.paid;
    if (t.isAnnounce) return Icons.campaign;
    return Icons.music_note;
  }

  Color _colorForTrack(RadioTrack t) {
    if (t.isAd) return Colors.orange;
    if (t.isAnnounce) return Colors.blue;
    return Colors.grey;
  }
}
