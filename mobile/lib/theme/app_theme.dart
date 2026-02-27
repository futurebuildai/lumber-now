import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'app_colors.dart';
import 'app_input_decoration.dart';
import 'app_typography.dart';
import 'design_tokens.dart';

class AppTheme {
  static AppColors? _currentColors;
  static AppColors get colors =>
      _currentColors ?? AppColors.fromTenant(const Color(0xFF1E40AF), const Color(0xFF1E3A5F));

  static ThemeData light(Color primary, Color secondary) {
    final colors = AppColors.fromTenant(primary, secondary);
    _currentColors = colors;

    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.light,
      colorSchemeSeed: primary,
      scaffoldBackgroundColor: colors.background,
      cardTheme: CardThemeData(
        color: colors.cardBackground,
        elevation: Elevations.xs,
        margin: const EdgeInsets.only(bottom: Spacing.sm),
        shape: RoundedRectangleBorder(
          borderRadius: Radii.borderMd,
          side: BorderSide(color: colors.borderLight, width: 0.5),
        ),
      ),
      appBarTheme: AppBarTheme(
        backgroundColor: colors.primary,
        foregroundColor: colors.textInverse,
        elevation: 0,
        scrolledUnderElevation: Elevations.sm,
        centerTitle: false,
        systemOverlayStyle: SystemUiOverlayStyle(
          statusBarColor: Colors.transparent,
          statusBarIconBrightness: Brightness.light,
          systemNavigationBarColor: colors.surface,
        ),
        titleTextStyle: AppTypography.title.copyWith(
          color: colors.textInverse,
        ),
      ),
      inputDecorationTheme: AppInputDecoration.theme(colors),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: colors.primary,
          foregroundColor: colors.textInverse,
          textStyle: AppTypography.button,
          padding: const EdgeInsets.symmetric(
            horizontal: Spacing.xl,
            vertical: Spacing.lg,
          ),
          minimumSize: const Size(0, TouchTargets.minimum),
          shape: RoundedRectangleBorder(borderRadius: Radii.borderMd),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: colors.primary,
          textStyle: AppTypography.button,
          padding: const EdgeInsets.symmetric(
            horizontal: Spacing.xl,
            vertical: Spacing.lg,
          ),
          minimumSize: const Size(0, TouchTargets.minimum),
          shape: RoundedRectangleBorder(borderRadius: Radii.borderMd),
          side: BorderSide(color: colors.border),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: colors.primary,
          textStyle: AppTypography.button,
          minimumSize: const Size(0, TouchTargets.minimum),
          shape: RoundedRectangleBorder(borderRadius: Radii.borderMd),
        ),
      ),
      chipTheme: ChipThemeData(
        shape: RoundedRectangleBorder(borderRadius: Radii.borderFull),
        labelStyle: AppTypography.caption.copyWith(fontWeight: FontWeight.w600),
      ),
      dividerTheme: DividerThemeData(
        color: colors.divider,
        thickness: 1,
        space: 0,
      ),
      drawerTheme: DrawerThemeData(
        backgroundColor: colors.surface,
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.horizontal(right: Radius.circular(Radii.lg)),
        ),
      ),
      bottomSheetTheme: BottomSheetThemeData(
        backgroundColor: colors.surface,
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(Radii.xl)),
        ),
      ),
      snackBarTheme: SnackBarThemeData(
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: Radii.borderMd),
      ),
    );
  }
}
