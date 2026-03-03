import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/models.dart';
import '../models/pagination.dart';
import '../services/api_client.dart';
import '../services/media_service.dart';
import '../utils/api_error.dart';
import 'package:dio/dio.dart';
import 'auth_providers.dart';

final mediaServiceProvider = Provider<MediaService>((ref) {
  return MediaService(ref.read(apiClientProvider));
});

final requestsProvider = FutureProvider<List<MaterialRequest>>((ref) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/requests', queryParameters: {'limit': 50, 'offset': 0});
  final items = (response.data['requests'] as List<dynamic>?)
          ?.map((e) => MaterialRequest.fromJson(e as Map<String, dynamic>))
          .toList() ??
      [];
  return items;
});

final requestListProvider =
    StateNotifierProvider<RequestListNotifier, AsyncValue<PaginatedResponse<MaterialRequest>>>((ref) {
  return RequestListNotifier(ref.read(apiClientProvider));
});

class RequestListNotifier extends StateNotifier<AsyncValue<PaginatedResponse<MaterialRequest>>> {
  final ApiClient _api;
  String _search = '';
  String? _statusFilter;

  RequestListNotifier(this._api) : super(const AsyncValue.loading()) {
    _load();
  }

  Future<void> _load({int offset = 0, bool append = false}) async {
    try {
      if (!append) state = const AsyncValue.loading();

      final params = <String, dynamic>{
        'limit': 20,
        'offset': offset,
      };
      if (_search.isNotEmpty) params['search'] = _search;
      if (_statusFilter != null) params['status'] = _statusFilter;

      final response = await _api.dio.get('/requests', queryParameters: params);
      final rawItems = (response.data['requests'] as List<dynamic>?) ?? [];
      final items = rawItems
          .map((e) => MaterialRequest.fromJson(e as Map<String, dynamic>))
          .toList();
      final total = response.data['total'] as int? ?? items.length;

      final existingItems = append
          ? (state.valueOrNull?.items ?? <MaterialRequest>[])
          : <MaterialRequest>[];

      state = AsyncValue.data(PaginatedResponse(
        items: [...existingItems, ...items],
        total: total,
        offset: offset,
        limit: 20,
      ));
    } on DioException catch (e, st) {
      state = AsyncValue.error(ApiError.fromDioException(e), st);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  Future<void> loadNextPage() async {
    final current = state.valueOrNull;
    if (current == null || !current.hasMore) return;
    await _load(offset: current.nextOffset, append: true);
  }

  Future<void> refresh() async {
    await _load();
  }

  void search(String query) {
    _search = query;
    _load();
  }

  void filterByStatus(String? status) {
    _statusFilter = status;
    _load();
  }
}
