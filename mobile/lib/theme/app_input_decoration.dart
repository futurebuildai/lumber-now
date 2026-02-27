import 'package:flutter/material.dart';
import 'design_tokens.dart';
import 'app_colors.dart';
import 'app_typography.dart';

abstract class AppInputDecoration {
  static InputDecorationTheme theme(AppColors colors) => InputDecorationTheme(
        filled: true,
        fillColor: colors.surface,
        contentPadding: const EdgeInsets.symmetric(
          horizontal: Spacing.lg,
          vertical: Spacing.md,
        ),
        hintStyle: AppTypography.body.copyWith(color: colors.textTertiary),
        labelStyle: AppTypography.label.copyWith(color: colors.textSecondary),
        floatingLabelStyle: AppTypography.label.copyWith(color: colors.primary),
        errorStyle: AppTypography.caption.copyWith(color: colors.error),
        border: OutlineInputBorder(
          borderRadius: Radii.borderMd,
          borderSide: BorderSide(color: colors.border),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: Radii.borderMd,
          borderSide: BorderSide(color: colors.border),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: Radii.borderMd,
          borderSide: BorderSide(color: colors.primary, width: 2),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: Radii.borderMd,
          borderSide: BorderSide(color: colors.error),
        ),
        focusedErrorBorder: OutlineInputBorder(
          borderRadius: Radii.borderMd,
          borderSide: BorderSide(color: colors.error, width: 2),
        ),
        disabledBorder: OutlineInputBorder(
          borderRadius: Radii.borderMd,
          borderSide: BorderSide(color: colors.borderLight),
        ),
      );
}
