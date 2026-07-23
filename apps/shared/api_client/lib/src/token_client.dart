import 'client.dart';

class TokenClient {
  final SimplexNodeClient _client;
  TokenClient(this._client);

  Future<Map<String, dynamic>> list() =>
      _client.get('/api/token/list');

  Future<Map<String, dynamic>> balances({required String pubkey}) =>
      _client.get('/api/token/balances', params: {'pubkey': pubkey});

  Future<Map<String, dynamic>> addCustom({
    required String symbol,
    required String name,
    int decimals = 18,
    String chain = 'custom',
    String? contractAddress,
    String? logoUrl,
  }) =>
      _client.post('/api/token/add-custom', body: {
        'symbol': symbol,
        'name': name,
        'decimals': decimals,
        'chain': chain,
        if (contractAddress != null) 'contract_address': contractAddress,
        if (logoUrl != null) 'logo_url': logoUrl,
      });

  Future<Map<String, dynamic>> removeCustom(String symbol) =>
      _client.post('/api/token/remove-custom', body: {'symbol': symbol});

  Future<Map<String, dynamic>> updateBalance({
    required String pubkey,
    required String symbol,
    required String balance,
  }) =>
      _client.post('/api/token/update-balance', body: {
        'pubkey': pubkey,
        'symbol': symbol,
        'balance': balance,
      });
}
