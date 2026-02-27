import 'dart:io';
import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';

class MediaPreview extends StatelessWidget {
  final File file;
  final String type;
  final VoidCallback onRemove;

  const MediaPreview({
    super.key,
    required this.file,
    required this.type,
    required this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    final fileName = file.path.split('/').last;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(Spacing.md),
        child: Row(
          children: [
            Container(
              width: 56,
              height: 56,
              decoration: BoxDecoration(
                borderRadius: Radii.borderSm,
                color: colors.surfaceVariant,
              ),
              clipBehavior: Clip.antiAlias,
              child: type == 'image'
                  ? Image.file(file, fit: BoxFit.cover)
                  : Center(
                      child: Icon(
                        Icons.picture_as_pdf,
                        color: colors.error,
                        size: IconSizes.lg,
                        semanticLabel: 'PDF file',
                      ),
                    ),
            ),
            const SizedBox(width: Spacing.md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    fileName,
                    style: AppTypography.bodySmall.copyWith(
                      color: colors.textPrimary,
                      fontWeight: FontWeight.w500,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: Spacing.xs),
                  Text(
                    type == 'image' ? 'Image' : 'PDF Document',
                    style: AppTypography.caption.copyWith(color: colors.textTertiary),
                  ),
                ],
              ),
            ),
            IconButton(
              onPressed: onRemove,
              icon: Icon(Icons.close, color: colors.textTertiary),
              tooltip: 'Remove file',
            ),
          ],
        ),
      ),
    );
  }
}
