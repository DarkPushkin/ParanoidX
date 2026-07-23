import 'client.dart';

class ExternalWalletClient {
  final SimplexNodeClient _client;
  ExternalWalletClient(this._client);

  Future<Map<String, dynamic>> list({required String pubkey}) =>
      _client.get('/api/external-wallet/list', params: {'pubkey': pubkey});

  Future<Map<String, dynamic>> link({
    required String pubkey,
    required String walletType,
    required String walletAddress,
    String? label,
    String? chain,
  }) =>
      _client.post('/api/external-wallet/link', body: {
        'pubkey': pubkey,
        'wallet_type': walletType,
        'wallet_address': walletAddress,
        if (label != null) 'label': label,
        if (chain != null) 'chain': chain,
      });

  Future<Map<String, dynamic>> unlink(String pubkey, String walletType) =>
      _client.post('/api/external-wallet/unlink', body: {
        'pubkey': pubkey,
        'wallet_type': walletType,
      });

  Future<Map<String, dynamic>> sync(String pubkey) =>
      _client.post('/api/external-wallet/sync', body: {'pubkey': pubkey});

  Future<Map<String, dynamic>> verify(String pubkey, String walletType) =>
      _client.post('/api/external-wallet/verify', body: {
        'pubkey': pubkey,
        'wallet_type': walletType,
      });
}
