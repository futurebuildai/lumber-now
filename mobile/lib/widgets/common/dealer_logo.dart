import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';

class DealerLogo extends StatelessWidget {
  final String? logoUrl;
  final double size;

  const DealerLogo({
    super.key,
    this.logoUrl,
    this.size = 32,
  });

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    if (logoUrl == null || logoUrl!.isEmpty) {
      return Icon(
        Icons.store_rounded,
        size: size,
        color: colors.textInverse,
        semanticLabel: 'Dealer logo',
      );
    }

    return CachedNetworkImage(
      imageUrl: logoUrl!,
      width: size,
      height: size,
      fit: BoxFit.contain,
      placeholder: (_, __) => SizedBox(
        width: size,
        height: size,
        child: const Center(
          child: CircularProgressIndicator(strokeWidth: 2),
        ),
      ),
      errorWidget: (_, __, ___) => Icon(
        Icons.store_rounded,
        size: size,
        color: colors.textInverse,
        semanticLabel: 'Dealer logo',
      ),
    );
  }
}
