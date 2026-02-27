import 'package:flutter/material.dart';
import '../../theme/design_tokens.dart';
import 'shimmer_card.dart';

class ShimmerList extends StatelessWidget {
  final int itemCount;
  final double itemHeight;
  final EdgeInsets padding;

  const ShimmerList({
    super.key,
    this.itemCount = 5,
    this.itemHeight = 80,
    this.padding = const EdgeInsets.all(Spacing.lg),
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: padding,
      child: Column(
        children: List.generate(
          itemCount,
          (_) => ShimmerCard(height: itemHeight),
        ),
      ),
    );
  }
}
