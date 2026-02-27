import 'package:flutter/material.dart';
import '../../models/models.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';
import '../../utils/formatters.dart';

class RequestSummaryCard extends StatelessWidget {
  final MaterialRequest request;
  final VoidCallback onTap;

  const RequestSummaryCard({
    super.key,
    required this.request,
    required this.onTap,
  });

  IconData get _inputTypeIcon {
    switch (request.inputType) {
      case 'voice':
        return Icons.mic_rounded;
      case 'image':
        return Icons.camera_alt_rounded;
      case 'pdf':
        return Icons.picture_as_pdf_rounded;
      default:
        return Icons.edit_rounded;
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    final statusColor = colors.statusColor(request.status);

    return Semantics(
      button: true,
      label: '${request.inputType} request, status ${request.status}, ${request.structuredItems.length} items',
      child: Card(
        margin: const EdgeInsets.only(bottom: Spacing.sm),
        child: InkWell(
          onTap: onTap,
          borderRadius: Radii.borderMd,
          child: Padding(
            padding: const EdgeInsets.all(Spacing.md),
            child: Row(
              children: [
                Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: colors.primary.withValues(alpha: 0.1),
                    borderRadius: Radii.borderSm,
                  ),
                  child: Icon(_inputTypeIcon, color: colors.primary, size: IconSizes.sm),
                ),
                const SizedBox(width: Spacing.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _description,
                        style: AppTypography.bodySmall
                            .copyWith(color: colors.textPrimary, fontWeight: FontWeight.w500),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: Spacing.xs),
                      Row(
                        children: [
                          Text(
                            Formatters.relativeTime(request.createdAt),
                            style: AppTypography.caption.copyWith(color: colors.textTertiary),
                          ),
                          if (request.structuredItems.isNotEmpty) ...[
                            Text(' \u2022 ', style: AppTypography.caption.copyWith(color: colors.textTertiary)),
                            Text(
                              '${request.structuredItems.length} items',
                              style: AppTypography.caption.copyWith(color: colors.textTertiary),
                            ),
                          ],
                        ],
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: Spacing.sm),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: statusColor.withValues(alpha: 0.1),
                    borderRadius: Radii.borderFull,
                  ),
                  child: Text(
                    request.status,
                    style: AppTypography.caption.copyWith(
                      color: statusColor,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  String get _description {
    if (request.rawText.isNotEmpty) {
      return request.rawText.length > 60
          ? '${request.rawText.substring(0, 60)}...'
          : request.rawText;
    }
    return '${request.inputType[0].toUpperCase()}${request.inputType.substring(1)} request';
  }
}
