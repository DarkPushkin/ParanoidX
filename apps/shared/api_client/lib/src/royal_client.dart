import 'client.dart';

/// RoyalClient manages royalclient functionality.
class RoyalClient {
  final SimplexNodeClient _client;
  RoyalClient(this._client);

  Future<Map<String, dynamic>> nodes() =>
      _client.get('/api/royal/nodes');

  Future<Map<String, dynamic>> register(String pubkey, String label, String addr) =>
      _client.post('/api/royal/register', body: {'pubkey': pubkey, 'label': label, 'addr': addr});

  Future<Map<String, dynamic>> sendCommand(String command, String targetPubkey) =>
      _client.post('/api/royal/command', body: {'command': command, 'target': targetPubkey});

  Future<Map<String, dynamic>> heartbeat(String pubkey) =>
      _client.get('/api/royal/heartbeat', params: {'pubkey': pubkey});

  Future<Map<String, dynamic>> royalKey() =>
      _client.get('/api/royal/key');
}
