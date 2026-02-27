import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../providers/providers.dart';
import '../theme/app_theme.dart';
import '../theme/app_typography.dart';
import '../theme/design_tokens.dart';
import '../utils/formatters.dart';
import 'common/dealer_logo.dart';

class BrandedScaffold extends ConsumerWidget {
  final Widget child;

  const BrandedScaffold({super.key, required this.child});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tenantConfig = ref.watch(tenantConfigProvider);
    final userState = ref.watch(currentUserProvider);
    final colors = AppTheme.colors;

    final config = tenantConfig.valueOrNull;
    final user = userState.valueOrNull;

    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            if (config != null && config.logoUrl.isNotEmpty) ...[
              DealerLogo(logoUrl: config.logoUrl, size: 28),
              const SizedBox(width: Spacing.sm),
            ],
            Text(config?.name ?? 'LumberNow'),
          ],
        ),
        leading: Builder(
          builder: (context) => IconButton(
            icon: const Icon(Icons.menu),
            onPressed: () => Scaffold.of(context).openDrawer(),
            tooltip: 'Open navigation menu',
          ),
        ),
      ),
      drawer: Drawer(
        child: SafeArea(
          child: Column(
            children: [
              // User info card
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(Spacing.lg),
                decoration: BoxDecoration(
                  color: colors.primary,
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    CircleAvatar(
                      radius: 28,
                      backgroundColor: colors.textInverse.withValues(alpha: 0.2),
                      child: Text(
                        user != null ? Formatters.initials(user.fullName) : '?',
                        style: AppTypography.title.copyWith(color: colors.textInverse),
                      ),
                    ),
                    const SizedBox(height: Spacing.md),
                    Text(
                      user?.fullName ?? '',
                      style: AppTypography.titleSmall.copyWith(color: colors.textInverse),
                    ),
                    const SizedBox(height: Spacing.xs),
                    Text(
                      user?.email ?? '',
                      style: AppTypography.caption.copyWith(
                        color: colors.textInverse.withValues(alpha: 0.8),
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: Spacing.sm),

              // Nav links
              _DrawerItem(
                icon: Icons.home_rounded,
                label: 'Home',
                onTap: () {
                  Navigator.pop(context);
                  context.go('/home');
                },
              ),
              _DrawerItem(
                icon: Icons.add_circle_rounded,
                label: 'New Request',
                onTap: () {
                  Navigator.pop(context);
                  context.push('/request/new');
                },
              ),
              _DrawerItem(
                icon: Icons.history_rounded,
                label: 'Request History',
                onTap: () {
                  Navigator.pop(context);
                  context.go('/history');
                },
              ),

              const Divider(),

              // Dealer contact info
              if (config != null) ...[
                Padding(
                  padding: const EdgeInsets.symmetric(
                      horizontal: Spacing.lg, vertical: Spacing.sm),
                  child: Text('Contact Dealer',
                      style: AppTypography.caption
                          .copyWith(color: colors.textTertiary)),
                ),
                if (config.contactEmail.isNotEmpty)
                  _DrawerItem(
                    icon: Icons.email_outlined,
                    label: config.contactEmail,
                    onTap: () {},
                  ),
                if (config.contactPhone.isNotEmpty)
                  _DrawerItem(
                    icon: Icons.phone_outlined,
                    label: config.contactPhone,
                    onTap: () {},
                  ),
              ],

              const Spacer(),

              const Divider(),

              _DrawerItem(
                icon: Icons.logout_rounded,
                label: 'Sign Out',
                onTap: () {
                  ref.read(currentUserProvider.notifier).logout();
                  Navigator.pop(context);
                  context.go('/login');
                },
              ),

              const SizedBox(height: Spacing.sm),
            ],
          ),
        ),
      ),
      body: Padding(
        padding: const EdgeInsets.all(Spacing.lg),
        child: child,
      ),
    );
  }
}

class _DrawerItem extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  const _DrawerItem({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    return Semantics(
      button: true,
      label: label,
      child: ListTile(
        leading: Icon(icon, color: colors.textSecondary, size: IconSizes.md),
        title: Text(label,
            style: AppTypography.bodySmall.copyWith(color: colors.textPrimary)),
        onTap: onTap,
        dense: true,
        minTileHeight: TouchTargets.minimum,
      ),
    );
  }
}
