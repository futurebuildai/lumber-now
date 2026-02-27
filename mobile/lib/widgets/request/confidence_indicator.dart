import 'dart:math' as math;
import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';

class ConfidenceIndicator extends StatefulWidget {
  final double confidence;
  final double size;
  final bool showLabel;

  const ConfidenceIndicator({
    super.key,
    required this.confidence,
    this.size = 56,
    this.showLabel = true,
  });

  @override
  State<ConfidenceIndicator> createState() => _ConfidenceIndicatorState();
}

class _ConfidenceIndicatorState extends State<ConfidenceIndicator>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late Animation<double> _animation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(vsync: this, duration: Durations.slower);
    _animation = Tween<double>(begin: 0, end: widget.confidence).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeOutCubic),
    );
    _controller.forward();
  }

  @override
  void didUpdateWidget(ConfidenceIndicator oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.confidence != widget.confidence) {
      _animation = Tween<double>(
        begin: _animation.value,
        end: widget.confidence,
      ).animate(CurvedAnimation(parent: _controller, curve: Curves.easeOutCubic));
      _controller
        ..reset()
        ..forward();
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    return AnimatedBuilder(
      animation: _animation,
      builder: (context, _) {
        final value = _animation.value;
        final color = colors.confidenceColor(value);
        final pct = (value * 100).toInt();
        return Semantics(
          label: 'Confidence: $pct percent',
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(
                width: widget.size,
                height: widget.size,
                child: CustomPaint(
                  painter: _RingPainter(
                    progress: value,
                    color: color,
                    trackColor: colors.borderLight,
                  ),
                  child: Center(
                    child: Text(
                      '$pct%',
                      style: AppTypography.label.copyWith(
                        color: color,
                        fontWeight: FontWeight.w700,
                        fontSize: widget.size * 0.24,
                      ),
                    ),
                  ),
                ),
              ),
              if (widget.showLabel) ...[
                const SizedBox(height: Spacing.xs),
                Text(
                  'Confidence',
                  style: AppTypography.caption.copyWith(color: colors.textTertiary),
                ),
              ],
            ],
          ),
        );
      },
    );
  }
}

class _RingPainter extends CustomPainter {
  final double progress;
  final Color color;
  final Color trackColor;

  _RingPainter({
    required this.progress,
    required this.color,
    required this.trackColor,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = (size.width - 6) / 2;

    final trackPaint = Paint()
      ..color = trackColor
      ..style = PaintingStyle.stroke
      ..strokeWidth = 4
      ..strokeCap = StrokeCap.round;

    final progressPaint = Paint()
      ..color = color
      ..style = PaintingStyle.stroke
      ..strokeWidth = 4
      ..strokeCap = StrokeCap.round;

    canvas.drawCircle(center, radius, trackPaint);
    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      -math.pi / 2,
      2 * math.pi * progress,
      false,
      progressPaint,
    );
  }

  @override
  bool shouldRepaint(_RingPainter oldDelegate) =>
      oldDelegate.progress != progress || oldDelegate.color != color;
}
