import 'dart:math' as math;

import 'package:flutter/material.dart';

class AppColors {
  final Color primary;
  final Color primaryLight;
  final Color primaryDark;
  final Color secondary;
  final Color secondaryLight;

  final Color success;
  final Color successLight;
  final Color warning;
  final Color warningLight;
  final Color error;
  final Color errorLight;
  final Color info;
  final Color infoLight;

  final Color confidenceHigh;
  final Color confidenceMedium;
  final Color confidenceLow;

  final Color statusPending;
  final Color statusProcessing;
  final Color statusParsed;
  final Color statusConfirmed;
  final Color statusSent;
  final Color statusFailed;

  final Color textPrimary;
  final Color textSecondary;
  final Color textTertiary;
  final Color textInverse;

  final Color surface;
  final Color surfaceVariant;
  final Color background;
  final Color border;
  final Color borderLight;
  final Color divider;

  final Color cardBackground;
  final Color shimmerBase;
  final Color shimmerHighlight;

  const AppColors._({
    required this.primary,
    required this.primaryLight,
    required this.primaryDark,
    required this.secondary,
    required this.secondaryLight,
    required this.success,
    required this.successLight,
    required this.warning,
    required this.warningLight,
    required this.error,
    required this.errorLight,
    required this.info,
    required this.infoLight,
    required this.confidenceHigh,
    required this.confidenceMedium,
    required this.confidenceLow,
    required this.statusPending,
    required this.statusProcessing,
    required this.statusParsed,
    required this.statusConfirmed,
    required this.statusSent,
    required this.statusFailed,
    required this.textPrimary,
    required this.textSecondary,
    required this.textTertiary,
    required this.textInverse,
    required this.surface,
    required this.surfaceVariant,
    required this.background,
    required this.border,
    required this.borderLight,
    required this.divider,
    required this.cardBackground,
    required this.shimmerBase,
    required this.shimmerHighlight,
  });

  factory AppColors.fromTenant(Color primary, Color secondary) {
    final hsl = HSLColor.fromColor(primary);
    final primaryLight =
        hsl.withLightness((hsl.lightness + 0.35).clamp(0.0, 1.0)).toColor();
    final primaryDark =
        hsl.withLightness((hsl.lightness - 0.15).clamp(0.0, 1.0)).toColor();

    final secHsl = HSLColor.fromColor(secondary);
    final secondaryLight =
        secHsl.withLightness((secHsl.lightness + 0.35).clamp(0.0, 1.0)).toColor();

    return AppColors._(
      primary: primary,
      primaryLight: primaryLight,
      primaryDark: primaryDark,
      secondary: secondary,
      secondaryLight: secondaryLight,
      success: const Color(0xFF16A34A),
      successLight: const Color(0xFFDCFCE7),
      warning: const Color(0xFFD97706),
      warningLight: const Color(0xFFFEF3C7),
      error: const Color(0xFFDC2626),
      errorLight: const Color(0xFFFEE2E2),
      info: const Color(0xFF2563EB),
      infoLight: const Color(0xFFDBEAFE),
      confidenceHigh: const Color(0xFF16A34A),
      confidenceMedium: const Color(0xFFD97706),
      confidenceLow: const Color(0xFFDC2626),
      statusPending: const Color(0xFFF59E0B),
      statusProcessing: const Color(0xFF3B82F6),
      statusParsed: const Color(0xFF8B5CF6),
      statusConfirmed: const Color(0xFF16A34A),
      statusSent: const Color(0xFF6B7280),
      statusFailed: const Color(0xFFDC2626),
      textPrimary: const Color(0xFF111827),
      textSecondary: const Color(0xFF6B7280),
      textTertiary: const Color(0xFF9CA3AF),
      textInverse: const Color(0xFFFFFFFF),
      surface: const Color(0xFFFFFFFF),
      surfaceVariant: const Color(0xFFF9FAFB),
      background: const Color(0xFFF3F4F6),
      border: const Color(0xFFD1D5DB),
      borderLight: const Color(0xFFE5E7EB),
      divider: const Color(0xFFE5E7EB),
      cardBackground: const Color(0xFFFFFFFF),
      shimmerBase: const Color(0xFFE5E7EB),
      shimmerHighlight: const Color(0xFFF9FAFB),
    );
  }

  Color confidenceColor(double confidence) {
    if (confidence >= 0.8) return confidenceHigh;
    if (confidence >= 0.5) return confidenceMedium;
    return confidenceLow;
  }

  Color statusColor(String status) {
    switch (status) {
      case 'pending':
        return statusPending;
      case 'processing':
        return statusProcessing;
      case 'parsed':
        return statusParsed;
      case 'confirmed':
        return statusConfirmed;
      case 'sent':
        return statusSent;
      case 'failed':
        return statusFailed;
      default:
        return statusPending;
    }
  }

  static Color parseHex(String hex) {
    hex = hex.replaceFirst('#', '');
    if (hex.length == 6) hex = 'FF$hex';
    return Color(int.parse(hex, radix: 16));
  }

  /// Linearizes an sRGB channel value per WCAG 2.1 specification.
  static double _srgbToLinear(double c) {
    return c <= 0.03928 ? c / 12.92 : math.pow((c + 0.055) / 1.055, 2.4).toDouble();
  }

  /// Calculates the WCAG 2.1 relative luminance of a color.
  /// See: https://www.w3.org/TR/WCAG21/#dfn-relative-luminance
  static double relativeLuminance(Color color) {
    final r = _srgbToLinear(color.red / 255.0);
    final g = _srgbToLinear(color.green / 255.0);
    final b = _srgbToLinear(color.blue / 255.0);
    return 0.2126 * r + 0.7152 * g + 0.0722 * b;
  }

  /// Calculates the WCAG 2.1 contrast ratio between two colors.
  /// Returns a value between 1.0 (no contrast) and 21.0 (maximum contrast).
  /// WCAG AA requires >= 4.5 for normal text, >= 3.0 for large text.
  /// WCAG AAA requires >= 7.0 for normal text, >= 4.5 for large text.
  static double contrastRatio(Color foreground, Color background) {
    final lumFg = relativeLuminance(foreground);
    final lumBg = relativeLuminance(background);
    final lighter = lumFg > lumBg ? lumFg : lumBg;
    final darker = lumFg > lumBg ? lumBg : lumFg;
    return (lighter + 0.05) / (darker + 0.05);
  }

  /// Checks if two colors meet the WCAG AA contrast requirement for normal text (>= 4.5:1).
  static bool meetsWcagAA(Color foreground, Color background) {
    return contrastRatio(foreground, background) >= 4.5;
  }

  /// Checks if two colors meet the WCAG AAA contrast requirement for normal text (>= 7.0:1).
  static bool meetsWcagAAA(Color foreground, Color background) {
    return contrastRatio(foreground, background) >= 7.0;
  }

  /// Checks if two colors meet the WCAG AA contrast requirement for large text (>= 3.0:1).
  static bool meetsWcagAALargeText(Color foreground, Color background) {
    return contrastRatio(foreground, background) >= 3.0;
  }
}
