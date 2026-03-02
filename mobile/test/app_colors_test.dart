import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:lumber_now/theme/app_colors.dart';

void main() {
  group('AppColors.parseHex', () {
    test('parses 6-digit hex', () {
      final color = AppColors.parseHex('#FF0000');
      expect(color, equals(const Color(0xFFFF0000)));
    });

    test('parses without hash prefix', () {
      final color = AppColors.parseHex('00FF00');
      expect(color, equals(const Color(0xFF00FF00)));
    });

    test('parses 8-digit hex with alpha', () {
      final color = AppColors.parseHex('80FF0000');
      expect(color, equals(const Color(0x80FF0000)));
    });
  });

  group('AppColors.relativeLuminance', () {
    test('black has luminance 0', () {
      final lum = AppColors.relativeLuminance(const Color(0xFF000000));
      expect(lum, closeTo(0.0, 0.001));
    });

    test('white has luminance 1', () {
      final lum = AppColors.relativeLuminance(const Color(0xFFFFFFFF));
      expect(lum, closeTo(1.0, 0.001));
    });

    test('mid-gray has intermediate luminance', () {
      final lum = AppColors.relativeLuminance(const Color(0xFF808080));
      expect(lum, greaterThan(0.1));
      expect(lum, lessThan(0.5));
    });
  });

  group('AppColors.contrastRatio', () {
    test('black on white has maximum contrast (21:1)', () {
      final ratio = AppColors.contrastRatio(
        const Color(0xFF000000),
        const Color(0xFFFFFFFF),
      );
      expect(ratio, closeTo(21.0, 0.1));
    });

    test('same colors have contrast ratio of 1', () {
      final ratio = AppColors.contrastRatio(
        const Color(0xFFFF0000),
        const Color(0xFFFF0000),
      );
      expect(ratio, closeTo(1.0, 0.001));
    });

    test('order does not affect contrast ratio', () {
      final ratio1 = AppColors.contrastRatio(
        const Color(0xFF000000),
        const Color(0xFFFFFFFF),
      );
      final ratio2 = AppColors.contrastRatio(
        const Color(0xFFFFFFFF),
        const Color(0xFF000000),
      );
      expect(ratio1, closeTo(ratio2, 0.001));
    });
  });

  group('WCAG compliance checks', () {
    test('black on white meets AA', () {
      expect(
        AppColors.meetsWcagAA(
          const Color(0xFF000000),
          const Color(0xFFFFFFFF),
        ),
        isTrue,
      );
    });

    test('black on white meets AAA', () {
      expect(
        AppColors.meetsWcagAAA(
          const Color(0xFF000000),
          const Color(0xFFFFFFFF),
        ),
        isTrue,
      );
    });

    test('light gray on white fails AA', () {
      expect(
        AppColors.meetsWcagAA(
          const Color(0xFFCCCCCC),
          const Color(0xFFFFFFFF),
        ),
        isFalse,
      );
    });

    test('dark text on white meets AA for large text', () {
      expect(
        AppColors.meetsWcagAALargeText(
          const Color(0xFF555555),
          const Color(0xFFFFFFFF),
        ),
        isTrue,
      );
    });
  });

  group('AppColors.fromTenant', () {
    test('creates valid color palette', () {
      final colors = AppColors.fromTenant(
        const Color(0xFF1E40AF),
        const Color(0xFF1E3A5F),
      );
      expect(colors.primary, equals(const Color(0xFF1E40AF)));
      expect(colors.secondary, equals(const Color(0xFF1E3A5F)));
      expect(colors.textPrimary, equals(const Color(0xFF111827)));
    });

    test('primary text on surface meets WCAG AA', () {
      final colors = AppColors.fromTenant(
        const Color(0xFF1E40AF),
        const Color(0xFF1E3A5F),
      );
      expect(
        AppColors.meetsWcagAA(colors.textPrimary, colors.surface),
        isTrue,
      );
    });
  });

  group('AppColors.confidenceColor', () {
    test('high confidence returns green', () {
      final colors = AppColors.fromTenant(
        const Color(0xFF1E40AF),
        const Color(0xFF1E3A5F),
      );
      expect(colors.confidenceColor(0.9), equals(colors.confidenceHigh));
    });

    test('medium confidence returns warning', () {
      final colors = AppColors.fromTenant(
        const Color(0xFF1E40AF),
        const Color(0xFF1E3A5F),
      );
      expect(colors.confidenceColor(0.6), equals(colors.confidenceMedium));
    });

    test('low confidence returns error', () {
      final colors = AppColors.fromTenant(
        const Color(0xFF1E40AF),
        const Color(0xFF1E3A5F),
      );
      expect(colors.confidenceColor(0.3), equals(colors.confidenceLow));
    });
  });

  group('AppColors.statusColor', () {
    test('returns correct color for each status', () {
      final colors = AppColors.fromTenant(
        const Color(0xFF1E40AF),
        const Color(0xFF1E3A5F),
      );
      expect(colors.statusColor('pending'), equals(colors.statusPending));
      expect(colors.statusColor('processing'), equals(colors.statusProcessing));
      expect(colors.statusColor('parsed'), equals(colors.statusParsed));
      expect(colors.statusColor('confirmed'), equals(colors.statusConfirmed));
      expect(colors.statusColor('sent'), equals(colors.statusSent));
      expect(colors.statusColor('failed'), equals(colors.statusFailed));
    });

    test('returns pending for unknown status', () {
      final colors = AppColors.fromTenant(
        const Color(0xFF1E40AF),
        const Color(0xFF1E3A5F),
      );
      expect(colors.statusColor('unknown'), equals(colors.statusPending));
    });
  });
}
