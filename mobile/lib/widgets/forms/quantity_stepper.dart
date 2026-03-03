import 'package:flutter/material.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';
import '../../utils/haptics.dart';

class QuantityStepper extends StatelessWidget {
  final double value;
  final ValueChanged<double> onChanged;
  final double min;
  final double step;
  final String unit;

  const QuantityStepper({
    super.key,
    required this.value,
    required this.onChanged,
    this.min = 1,
    this.step = 1,
    this.unit = '',
  });

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    final displayValue = value == value.roundToDouble()
        ? value.toInt().toString()
        : value.toStringAsFixed(1);

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        _StepButton(
          icon: Icons.remove,
          onPressed: value > min
              ? () {
                  Haptics.light();
                  onChanged((value - step).clamp(min, double.infinity));
                }
              : null,
          colors: colors,
          semanticLabel: 'Decrease quantity',
        ),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: Spacing.md),
          child: Semantics(
            label: 'Quantity: $displayValue $unit',
            child: Text(
              unit.isEmpty ? displayValue : '$displayValue $unit',
              style: AppTypography.label.copyWith(color: colors.textPrimary),
            ),
          ),
        ),
        _StepButton(
          icon: Icons.add,
          onPressed: () {
            Haptics.light();
            onChanged(value + step);
          },
          colors: colors,
          semanticLabel: 'Increase quantity',
        ),
      ],
    );
  }
}

class _StepButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback? onPressed;
  final AppColors colors;
  final String semanticLabel;

  const _StepButton({
    required this.icon,
    required this.onPressed,
    required this.colors,
    required this.semanticLabel,
  });

  @override
  Widget build(BuildContext context) {
    final enabled = onPressed != null;
    return Semantics(
      button: true,
      label: semanticLabel,
      child: Material(
        color: enabled ? colors.primary.withValues(alpha: 0.1) : colors.borderLight,
        borderRadius: Radii.borderSm,
        child: InkWell(
          onTap: onPressed,
          borderRadius: Radii.borderSm,
          child: SizedBox(
            width: TouchTargets.minimum,
            height: 36,
            child: Icon(
              icon,
              size: IconSizes.sm,
              color: enabled ? colors.primary : colors.textTertiary,
            ),
          ),
        ),
      ),
    );
  }
}
