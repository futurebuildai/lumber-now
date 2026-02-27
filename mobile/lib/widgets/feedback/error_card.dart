import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';

class ErrorCard extends StatelessWidget {
  final String message;
  final VoidCallback? onRetry;

  const ErrorCard({
    super.key,
    required this.message,
    this.onRetry,
  });

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    return Semantics(
      label: 'Error: $message',
      child: Container(
        padding: const EdgeInsets.all(Spacing.lg),
        decoration: BoxDecoration(
          color: colors.errorLight,
          borderRadius: Radii.borderMd,
          border: Border.all(color: colors.error.withValues(alpha: 0.3)),
        ),
        child: Row(
          children: [
            Icon(Icons.error_outline, color: colors.error, size: IconSizes.md,
                semanticLabel: 'Error icon'),
            const SizedBox(width: Spacing.md),
            Expanded(
              child: Text(
                message,
                style: AppTypography.bodySmall.copyWith(color: colors.error),
              ),
            ),
            if (onRetry != null) ...[
              const SizedBox(width: Spacing.sm),
              TextButton(
                onPressed: onRetry,
                child: Text('Try Again',
                    style: AppTypography.label.copyWith(color: colors.error)),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
