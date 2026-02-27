import 'package:flutter/material.dart';
import 'design_tokens.dart';
import 'app_colors.dart';
import 'app_typography.dart';

abstract class AppButtonStyles {
  static ButtonStyle primary(AppColors colors) => FilledButton.styleFrom(
        backgroundColor: colors.primary,
        foregroundColor: colors.textInverse,
        textStyle: AppTypography.button,
        padding: const EdgeInsets.symmetric(
          horizontal: Spacing.xl,
          vertical: Spacing.lg,
        ),
        minimumSize: const Size(0, TouchTargets.minimum),
        shape: RoundedRectangleBorder(
          borderRadius: Radii.borderMd,
        ),
        elevation: Elevations.none,
      );

  static ButtonStyle secondary(AppColors colors) => OutlinedButton.styleFrom(
        foregroundColor: colors.primary,
        textStyle: AppTypography.button,
        padding: const EdgeInsets.symmetric(
          horizontal: Spacing.xl,
          vertical: Spacing.lg,
        ),
        minimumSize: const Size(0, TouchTargets.minimum),
        shape: RoundedRectangleBorder(
          borderRadius: Radii.borderMd,
        ),
        side: BorderSide(color: colors.border),
      );

  static ButtonStyle danger(AppColors colors) => FilledButton.styleFrom(
        backgroundColor: colors.error,
        foregroundColor: colors.textInverse,
        textStyle: AppTypography.button,
        padding: const EdgeInsets.symmetric(
          horizontal: Spacing.xl,
          vertical: Spacing.lg,
        ),
        minimumSize: const Size(0, TouchTargets.minimum),
        shape: RoundedRectangleBorder(
          borderRadius: Radii.borderMd,
        ),
        elevation: Elevations.none,
      );

  static ButtonStyle ghost(AppColors colors) => TextButton.styleFrom(
        foregroundColor: colors.primary,
        textStyle: AppTypography.button,
        padding: const EdgeInsets.symmetric(
          horizontal: Spacing.lg,
          vertical: Spacing.md,
        ),
        minimumSize: const Size(0, TouchTargets.minimum),
        shape: RoundedRectangleBorder(
          borderRadius: Radii.borderMd,
        ),
      );
}
