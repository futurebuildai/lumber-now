import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../config/env.dart';
import '../models/models.dart';
import 'api_client.dart';

class TenantService {
  final ApiClient _api;
  final FlutterSecureStorage _storage = const FlutterSecureStorage();

  TenantService(this._api);

  Future<TenantConfig?> fetchConfig() async {
    final slug = Env.tenantSlug;
    if (slug.isEmpty) return null;

    try {
      final response =
          await _api.dio.get('/tenant/config', queryParameters: {'slug': slug});
      final config = TenantConfig.fromJson(response.data);
      await _storage.write(key: 'tenant_id', value: config.dealerId);
      return config;
    } catch (_) {
      return null;
    }
  }
}
