import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';
import '../../widgets/feedback/empty_state.dart';
import '../../widgets/feedback/error_card.dart';
import '../../widgets/loading/shimmer_list.dart';
import '../../widgets/request/request_summary_card.dart';

class RequestHistoryScreen extends ConsumerStatefulWidget {
  const RequestHistoryScreen({super.key});

  @override
  ConsumerState<RequestHistoryScreen> createState() =>
      _RequestHistoryScreenState();
}

class _RequestHistoryScreenState extends ConsumerState<RequestHistoryScreen> {
  final _searchController = TextEditingController();
  final _scrollController = ScrollController();
  String? _selectedStatus;
  Timer? _debounce;

  static const _statusFilters = [
    null,
    'pending',
    'processing',
    'parsed',
    'confirmed',
    'sent',
  ];

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _searchController.dispose();
    _scrollController.dispose();
    _debounce?.cancel();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
        _scrollController.position.maxScrollExtent - 200) {
      ref.read(requestListProvider.notifier).loadNextPage();
    }
  }

  void _onSearchChanged(String query) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 400), () {
      ref.read(requestListProvider.notifier).search(query);
    });
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    final listState = ref.watch(requestListProvider);

    return Column(
      children: [
        // Search bar
        Semantics(
          label: 'Search requests',
          child: TextField(
            controller: _searchController,
            decoration: InputDecoration(
              hintText: 'Search requests...',
              prefixIcon: const Icon(Icons.search_rounded),
              suffixIcon: _searchController.text.isNotEmpty
                  ? IconButton(
                      icon: const Icon(Icons.close),
                      onPressed: () {
                        _searchController.clear();
                        ref.read(requestListProvider.notifier).search('');
                      },
                      tooltip: 'Clear search',
                    )
                  : null,
            ),
            onChanged: _onSearchChanged,
          ),
        ),
        const SizedBox(height: Spacing.md),

        // Status filter chips
        SizedBox(
          height: 36,
          child: ListView(
            scrollDirection: Axis.horizontal,
            children: _statusFilters.map((status) {
              final isSelected = status == _selectedStatus;
              final label = status == null
                  ? 'All'
                  : '${status[0].toUpperCase()}${status.substring(1)}';
              return Padding(
                padding: const EdgeInsets.only(right: Spacing.sm),
                child: Semantics(
                  button: true,
                  label: 'Filter by $label',
                  child: FilterChip(
                    label: Text(label),
                    selected: isSelected,
                    onSelected: (_) {
                      setState(() => _selectedStatus = status);
                      ref.read(requestListProvider.notifier).filterByStatus(status);
                    },
                    selectedColor: colors.primary.withValues(alpha: 0.15),
                    checkmarkColor: colors.primary,
                    labelStyle: AppTypography.caption.copyWith(
                      color: isSelected ? colors.primary : colors.textSecondary,
                      fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400,
                    ),
                  ),
                ),
              );
            }).toList(),
          ),
        ),
        const SizedBox(height: Spacing.md),

        // Request list
        Expanded(
          child: RefreshIndicator(
            onRefresh: () async {
              ref.read(requestListProvider.notifier).refresh();
            },
            child: listState.when(
              loading: () => const ShimmerList(itemCount: 6, itemHeight: 72),
              error: (e, _) => SingleChildScrollView(
                physics: const AlwaysScrollableScrollPhysics(),
                child: ErrorCard(
                  message: e.toString(),
                  onRetry: () =>
                      ref.read(requestListProvider.notifier).refresh(),
                ),
              ),
              data: (paginated) {
                if (paginated.items.isEmpty) {
                  return SingleChildScrollView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    child: EmptyState(
                      icon: Icons.inbox_rounded,
                      title: 'No requests found',
                      subtitle: _searchController.text.isNotEmpty
                          ? 'Try a different search term'
                          : 'Submit your first material request',
                      actionLabel: _searchController.text.isEmpty
                          ? 'New Request'
                          : null,
                      onAction: _searchController.text.isEmpty
                          ? () => context.push('/request/new')
                          : null,
                    ),
                  );
                }

                return ListView.builder(
                  controller: _scrollController,
                  physics: const AlwaysScrollableScrollPhysics(),
                  itemCount:
                      paginated.items.length + (paginated.hasMore ? 1 : 0),
                  itemBuilder: (context, index) {
                    if (index >= paginated.items.length) {
                      return const Padding(
                        padding: EdgeInsets.all(Spacing.lg),
                        child: Center(
                          child: CircularProgressIndicator(strokeWidth: 2),
                        ),
                      );
                    }

                    final req = paginated.items[index];
                    return RequestSummaryCard(
                      request: req,
                      onTap: () => context.push('/request/${req.id}'),
                    );
                  },
                );
              },
            ),
          ),
        ),
      ],
    );
  }
}
