import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';

class SuccessAnimation extends StatefulWidget {
  final String? message;
  final VoidCallback? onComplete;
  final Duration displayDuration;

  const SuccessAnimation({
    super.key,
    this.message,
    this.onComplete,
    this.displayDuration = const Duration(seconds: 2),
  });

  @override
  State<SuccessAnimation> createState() => _SuccessAnimationState();
}

class _SuccessAnimationState extends State<SuccessAnimation>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _scaleAnimation;
  late final Animation<double> _fadeAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: AppDurations.slow,
    );
    _scaleAnimation = Tween<double>(begin: 0, end: 1).animate(
      CurvedAnimation(parent: _controller, curve: Curves.elasticOut),
    );
    _fadeAnimation = Tween<double>(begin: 0, end: 1).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeIn),
    );
    _controller.forward();
    Future.delayed(widget.displayDuration, () {
      if (mounted) widget.onComplete?.call();
    });
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    return Center(
      child: FadeTransition(
        opacity: _fadeAnimation,
        child: ScaleTransition(
          scale: _scaleAnimation,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(
                  color: colors.successLight,
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  Icons.check_rounded,
                  size: IconSizes.xl,
                  color: colors.success,
                  semanticLabel: 'Success',
                ),
              ),
              if (widget.message != null) ...[
                const SizedBox(height: Spacing.lg),
                Text(
                  widget.message!,
                  style: AppTypography.title.copyWith(color: colors.textPrimary),
                  textAlign: TextAlign.center,
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
