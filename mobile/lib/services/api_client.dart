import 'dart:math';

import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../config/env.dart';
import '../utils/api_error.dart';

class ApiClient {
  late final Dio dio;
  final FlutterSecureStorage _storage = const FlutterSecureStorage();

  ApiClient() {
    dio = Dio(BaseOptions(
      baseUrl: Env.apiBaseUrl,
      connectTimeout: const Duration(seconds: 30),
      receiveTimeout: const Duration(seconds: 30),
      headers: {
        'Content-Type': 'application/json',
        'X-Requested-With': 'XMLHttpRequest',
      },
    ));

    dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await _storage.read(key: 'access_token');
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }

        final tenantId = await _storage.read(key: 'tenant_id');
        if (tenantId != null) {
          options.headers['X-Tenant-ID'] = tenantId;
        }

        if (options.method == 'POST' && !options.headers.containsKey('Idempotency-Key')) {
          options.headers['Idempotency-Key'] = _generateUuid();
        }

        return handler.next(options);
      },
      onError: (error, handler) async {
        if (error.response?.statusCode == 401) {
          final refreshToken = await _storage.read(key: 'refresh_token');
          if (refreshToken != null) {
            try {
              final refreshDio = Dio(BaseOptions(
                baseUrl: Env.apiBaseUrl,
                connectTimeout: const Duration(seconds: 10),
                receiveTimeout: const Duration(seconds: 10),
              ));
              final response = await refreshDio.post(
                '/auth/refresh',
                data: {'refresh_token': refreshToken},
                options: Options(headers: {
                  'Content-Type': 'application/json',
                  'X-Tenant-ID': await _storage.read(key: 'tenant_id') ?? '',
                }),
              );
              final newToken = response.data['access_token'];
              await _storage.write(key: 'access_token', value: newToken);
              await _storage.write(
                  key: 'refresh_token',
                  value: response.data['refresh_token']);

              error.requestOptions.headers['Authorization'] =
                  'Bearer $newToken';
              final retryResponse = await dio.fetch(error.requestOptions);
              return handler.resolve(retryResponse);
            } catch (_) {
              await _storage.deleteAll();
            }
          }
        }
        // Wrap all DioExceptions in ApiError for user-friendly messages
        return handler.next(error.copyWith(
          error: ApiError.fromDioException(error),
        ));
      },
    ));
  }

  static final _random = Random.secure();

  static String _generateUuid() {
    final bytes = List<int>.generate(16, (_) => _random.nextInt(256));
    bytes[6] = (bytes[6] & 0x0f) | 0x40; // version 4
    bytes[8] = (bytes[8] & 0x3f) | 0x80; // variant 1
    final hex = bytes.map((b) => b.toRadixString(16).padLeft(2, '0')).join();
    return '${hex.substring(0, 8)}-${hex.substring(8, 12)}-${hex.substring(12, 16)}-${hex.substring(16, 20)}-${hex.substring(20)}';
  }
}
