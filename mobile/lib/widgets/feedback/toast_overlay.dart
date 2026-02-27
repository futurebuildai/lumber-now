import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';

enum ToastType { success, error, info }

class ToastOverlay {
  static void show(
    BuildContext context, {
    required String message,
    ToastType type = ToastType.info,
    Duration duration = const Duration(seconds: 3),
  }) {
    final colors = AppTheme.colors;
    final Color bgColor;
    final Color fgColor;
    final IconData icon;

    switch (type) {
      case ToastType.success:
        bgColor = colors.success;
        fgColor = Colors.white;
        icon = Icons.check_circle_rounded;
      case ToastType.error:
        bgColor = colors.error;
        fgColor = Colors.white;
        icon = Icons.error_rounded;
      case ToastType.info:
        bgColor = colors.primary;
        fgColor = Colors.white;
        icon = Icons.info_rounded;
    }

    ScaffoldMessenger.of(context).hideCurrentSnackBar();
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            Icon(icon, color: fgColor, size: IconSizes.sm),
            const SizedBox(width: Spacing.sm),
            Expanded(
              child: Text(
                message,
                style: AppTypography.bodySmall.copyWith(color: fgColor),
              ),
            ),
          ],
        ),
        backgroundColor: bgColor,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: Radii.borderMd),
        duration: duration,
        margin: const EdgeInsets.all(Spacing.lg),
      ),
    );
  }
}
