import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';

class StatusTimeline extends StatelessWidget {
  final String currentStatus;

  const StatusTimeline({super.key, required this.currentStatus});

  static const _steps = [
    ('pending', 'Submitted', Icons.upload_rounded),
    ('processing', 'Processing', Icons.psychology_rounded),
    ('parsed', 'Review', Icons.rate_review_rounded),
    ('confirmed', 'Confirmed', Icons.check_circle_rounded),
    ('sent', 'Sent', Icons.send_rounded),
  ];

  int get _currentIndex {
    final idx = _steps.indexWhere((s) => s.$1 == currentStatus);
    return idx >= 0 ? idx : 0;
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    final current = _currentIndex;

    return Semantics(
      label: 'Status: ${_steps[current].$2}',
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: Spacing.md),
        child: Row(
          children: List.generate(_steps.length * 2 - 1, (i) {
            if (i.isOdd) {
              final stepIndex = i ~/ 2;
              final completed = stepIndex < current;
              return Expanded(
                child: Container(
                  height: 2,
                  color: completed ? colors.primary : colors.borderLight,
                ),
              );
            }
            final stepIndex = i ~/ 2;
            final step = _steps[stepIndex];
            final isCompleted = stepIndex < current;
            final isCurrent = stepIndex == current;
            final isFailed = currentStatus == 'failed' && stepIndex == current;

            final Color circleColor;
            if (isFailed) {
              circleColor = colors.error;
            } else if (isCompleted) {
              circleColor = colors.primary;
            } else if (isCurrent) {
              circleColor = colors.primary;
            } else {
              circleColor = colors.borderLight;
            }

            return Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: 28,
                  height: 28,
                  decoration: BoxDecoration(
                    color: (isCompleted || isCurrent) ? circleColor : Colors.transparent,
                    border: Border.all(color: circleColor, width: 2),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    isCompleted
                        ? Icons.check
                        : isFailed
                            ? Icons.close
                            : step.$3,
                    size: 14,
                    color: (isCompleted || isCurrent)
                        ? colors.textInverse
                        : colors.textTertiary,
                  ),
                ),
                const SizedBox(height: Spacing.xs),
                Text(
                  step.$2,
                  style: AppTypography.caption.copyWith(
                    color: isCurrent ? colors.primary : colors.textTertiary,
                    fontWeight: isCurrent ? FontWeight.w600 : FontWeight.w400,
                  ),
                  textAlign: TextAlign.center,
                ),
              ],
            );
          }),
        ),
      ),
    );
  }
}
