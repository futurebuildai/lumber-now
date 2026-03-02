import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';

class LoadingOverlay extends StatelessWidget {
  final String? message;

  const LoadingOverlay({super.key, this.message});

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    return Semantics(
      label: message ?? 'Loading',
      liveRegion: true,
      child: Container(
      color: Colors.black.withValues(alpha: 0.4),
      child: Center(
        child: Container(
          padding: const EdgeInsets.all(Spacing.xl),
          decoration: BoxDecoration(
            color: colors.surface,
            borderRadius: Radii.borderLg,
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.1),
                blurRadius: Elevations.lg,
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(
                width: 48,
                height: 48,
                child: CircularProgressIndicator(
                  strokeWidth: 3,
                  color: colors.primary,
                ),
              ),
              if (message != null) ...[
                const SizedBox(height: Spacing.lg),
                Text(
                  message!,
                  style: AppTypography.body.copyWith(color: colors.textSecondary),
                  textAlign: TextAlign.center,
                ),
              ],
            ],
          ),
        ),
      ),
    ),
    );
  }
}
