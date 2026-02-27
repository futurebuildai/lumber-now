import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';
import '../../utils/formatters.dart';
import '../../widgets/common/section_header.dart';
import '../../widgets/feedback/empty_state.dart';
import '../../widgets/loading/shimmer_list.dart';
import '../../widgets/request/request_summary_card.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final userState = ref.watch(currentUserProvider);
    final requestsAsync = ref.watch(requestsProvider);
    final colors = AppTheme.colors;

    return RefreshIndicator(
      onRefresh: () async {
        ref.invalidate(requestsProvider);
      },
      child: SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Welcome card
            userState.when(
              data: (user) => user != null
                  ? Card(
                      child: Padding(
                        padding: const EdgeInsets.all(Spacing.lg),
                        child: Row(
                          children: [
                            CircleAvatar(
                              radius: 24,
                              backgroundColor: colors.primary.withValues(alpha: 0.1),
                              child: Text(
                                Formatters.initials(user.fullName),
                                style: AppTypography.titleSmall
                                    .copyWith(color: colors.primary),
                              ),
                            ),
                            const SizedBox(width: Spacing.md),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    'Welcome back,',
                                    style: AppTypography.caption
                                        .copyWith(color: colors.textSecondary),
                                  ),
                                  Text(
                                    user.fullName,
                                    style: AppTypography.title
                                        .copyWith(color: colors.textPrimary),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    )
                  : const SizedBox(),
              loading: () => const SizedBox(height: 80),
              error: (_, __) => const SizedBox(),
            ),

            const SizedBox(height: Spacing.lg),

            // Hero "New Request" card
            Semantics(
              button: true,
              label: 'Create new material request',
              child: Card(
                clipBehavior: Clip.antiAlias,
                child: InkWell(
                  onTap: () => context.push('/request/new'),
                  child: Container(
                    padding: const EdgeInsets.all(Spacing.xl),
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        colors: [colors.primary, colors.primaryDark],
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      ),
                    ),
                    child: Row(
                      children: [
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                'New Material Request',
                                style: AppTypography.title
                                    .copyWith(color: colors.textInverse),
                              ),
                              const SizedBox(height: Spacing.xs),
                              Text(
                                'Text, voice, photo, or PDF',
                                style: AppTypography.bodySmall.copyWith(
                                  color: colors.textInverse.withValues(alpha: 0.8),
                                ),
                              ),
                            ],
                          ),
                        ),
                        Container(
                          width: 48,
                          height: 48,
                          decoration: BoxDecoration(
                            color: colors.textInverse.withValues(alpha: 0.2),
                            borderRadius: Radii.borderMd,
                          ),
                          child: Icon(Icons.add_rounded,
                              color: colors.textInverse, size: IconSizes.md),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),

            const SizedBox(height: Spacing.xl),

            // Recent requests
            SectionHeader(
              title: 'Recent Requests',
              trailing: TextButton(
                onPressed: () => context.go('/history'),
                child: const Text('View All'),
              ),
            ),

            requestsAsync.when(
              loading: () => const ShimmerList(itemCount: 3, itemHeight: 72),
              error: (e, _) => EmptyState(
                icon: Icons.error_outline,
                title: 'Failed to load requests',
                actionLabel: 'Try Again',
                onAction: () => ref.invalidate(requestsProvider),
              ),
              data: (requests) {
                if (requests.isEmpty) {
                  return const EmptyState(
                    icon: Icons.inbox_rounded,
                    title: 'No requests yet',
                    subtitle: 'Submit your first material request to get started',
                  );
                }

                final recent = requests.take(3).toList();
                return Column(
                  children: recent
                      .map((req) => RequestSummaryCard(
                            request: req,
                            onTap: () => context.push('/request/${req.id}'),
                          ))
                      .toList(),
                );
              },
            ),

            const SizedBox(height: Spacing.lg),

            // Quick stats row
            requestsAsync.whenOrNull(
                  data: (requests) {
                    if (requests.isEmpty) return null;
                    final pending =
                        requests.where((r) => r.status == 'pending').length;
                    final processing =
                        requests.where((r) => r.status == 'processing').length;
                    final confirmed =
                        requests.where((r) => r.status == 'confirmed').length;
                    return Row(
                      children: [
                        _StatChip(
                            label: 'Pending',
                            count: pending,
                            color: colors.statusPending),
                        const SizedBox(width: Spacing.sm),
                        _StatChip(
                            label: 'Processing',
                            count: processing,
                            color: colors.statusProcessing),
                        const SizedBox(width: Spacing.sm),
                        _StatChip(
                            label: 'Confirmed',
                            count: confirmed,
                            color: colors.statusConfirmed),
                      ],
                    );
                  },
                ) ??
                const SizedBox(),
          ],
        ),
      ),
    );
  }
}

class _StatChip extends StatelessWidget {
  final String label;
  final int count;
  final Color color;

  const _StatChip({
    required this.label,
    required this.count,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(Spacing.md),
          child: Column(
            children: [
              Text(
                count.toString(),
                style: AppTypography.title.copyWith(color: color),
              ),
              const SizedBox(height: Spacing.xs),
              Text(
                label,
                style: AppTypography.caption
                    .copyWith(color: AppTheme.colors.textSecondary),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
