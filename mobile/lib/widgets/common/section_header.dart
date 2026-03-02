import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';

class SectionHeader extends StatelessWidget {
  final String title;
  final int? count;
  final Widget? trailing;

  const SectionHeader({
    super.key,
    required this.title,
    this.count,
    this.trailing,
  });

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    return Semantics(
      header: true,
      label: count != null ? '$title ($count)' : title,
      child: Padding(
      padding: const EdgeInsets.only(bottom: Spacing.sm),
      child: Row(
        children: [
          Expanded(
            child: Text.rich(
              TextSpan(
                text: title,
                style: AppTypography.titleSmall.copyWith(color: colors.textPrimary),
                children: count != null
                    ? [
                        TextSpan(
                          text: ' ($count)',
                          style: AppTypography.bodySmall.copyWith(
                            color: colors.textTertiary,
                            fontWeight: FontWeight.w400,
                          ),
                        ),
                      ]
                    : null,
              ),
            ),
          ),
          if (trailing != null) trailing!,
        ],
      ),
    ),
    );
  }
}
