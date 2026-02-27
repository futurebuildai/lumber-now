import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/models.dart';
import '../services/tenant_service.dart';
import 'auth_providers.dart';

final tenantServiceProvider = Provider<TenantService>((ref) {
  return TenantService(ref.read(apiClientProvider));
});

final tenantConfigProvider = FutureProvider<TenantConfig?>((ref) async {
  final tenantService = ref.read(tenantServiceProvider);
  return tenantService.fetchConfig();
});
