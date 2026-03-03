import 'package:flutter/material.dart';
import 'package:flutter_slidable/flutter_slidable.dart';
import '../../models/models.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';
import '../../utils/haptics.dart';
import '../forms/quantity_stepper.dart';
import 'confidence_indicator.dart';

class EditableSKUCard extends StatefulWidget {
  final StructuredItem item;
  final ValueChanged<StructuredItem> onChanged;
  final VoidCallback onDelete;

  const EditableSKUCard({
    super.key,
    required this.item,
    required this.onChanged,
    required this.onDelete,
  });

  @override
  State<EditableSKUCard> createState() => _EditableSKUCardState();
}

class _EditableSKUCardState extends State<EditableSKUCard> {
  bool _expanded = false;
  late TextEditingController _nameController;
  late TextEditingController _skuController;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.item.name);
    _skuController = TextEditingController(text: widget.item.sku);
  }

  @override
  void didUpdateWidget(EditableSKUCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.item.name != widget.item.name) {
      _nameController.text = widget.item.name;
    }
    if (oldWidget.item.sku != widget.item.sku) {
      _skuController.text = widget.item.sku;
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    _skuController.dispose();
    super.dispose();
  }

  void _emitChange({String? name, String? sku, double? quantity}) {
    widget.onChanged(widget.item.copyWith(
      name: name ?? _nameController.text,
      sku: sku ?? _skuController.text,
      quantity: quantity ?? widget.item.quantity,
    ));
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    final item = widget.item;

    return Semantics(
      label: '${item.name}, ${item.quantity} ${item.unit}, confidence ${(item.confidence * 100).toInt()} percent',
      child: Slidable(
        endActionPane: ActionPane(
          motion: const BehindMotion(),
          extentRatio: 0.25,
          children: [
            SlidableAction(
              onPressed: (_) {
                Haptics.medium();
                widget.onDelete();
              },
              backgroundColor: colors.error,
              foregroundColor: Colors.white,
              icon: Icons.delete_rounded,
              label: 'Delete',
              borderRadius: const BorderRadius.horizontal(
                right: Radius.circular(Radii.md),
              ),
            ),
          ],
        ),
        child: Card(
          margin: const EdgeInsets.only(bottom: Spacing.sm),
          child: InkWell(
            onTap: () {
              Haptics.selection();
              setState(() => _expanded = !_expanded);
            },
            borderRadius: Radii.borderMd,
            child: Padding(
              padding: const EdgeInsets.all(Spacing.md),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              item.name,
                              style: AppTypography.titleSmall
                                  .copyWith(color: colors.textPrimary),
                            ),
                            const SizedBox(height: Spacing.xs),
                            Row(
                              children: [
                                if (item.sku.isNotEmpty) ...[
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                        horizontal: 6, vertical: 2),
                                    decoration: BoxDecoration(
                                      color: colors.surfaceVariant,
                                      borderRadius: Radii.borderXs,
                                    ),
                                    child: Text(item.sku,
                                        style: AppTypography.mono
                                            .copyWith(color: colors.textSecondary)),
                                  ),
                                  const SizedBox(width: Spacing.sm),
                                ],
                                Text(
                                  '${item.quantity == item.quantity.roundToDouble() ? item.quantity.toInt() : item.quantity} ${item.unit}',
                                  style: AppTypography.bodySmall
                                      .copyWith(color: colors.textSecondary),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                      ConfidenceIndicator(
                        confidence: item.confidence,
                        size: 40,
                        showLabel: false,
                      ),
                      const SizedBox(width: Spacing.xs),
                      Icon(
                        _expanded
                            ? Icons.keyboard_arrow_up
                            : Icons.keyboard_arrow_down,
                        color: colors.textTertiary,
                        semanticLabel: _expanded ? 'Collapse' : 'Expand to edit',
                      ),
                    ],
                  ),
                  AnimatedCrossFade(
                    firstChild: const SizedBox.shrink(),
                    secondChild: _buildEditFields(colors),
                    crossFadeState: _expanded
                        ? CrossFadeState.showSecond
                        : CrossFadeState.showFirst,
                    duration: AppDurations.fast,
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildEditFields(AppColors colors) {
    return Padding(
      padding: const EdgeInsets.only(top: Spacing.md),
      child: Column(
        children: [
          const Divider(),
          const SizedBox(height: Spacing.sm),
          TextFormField(
            controller: _nameController,
            decoration: const InputDecoration(labelText: 'Item Name'),
            onChanged: (val) => _emitChange(name: val),
          ),
          const SizedBox(height: Spacing.sm),
          TextFormField(
            controller: _skuController,
            decoration: const InputDecoration(labelText: 'SKU'),
            onChanged: (val) => _emitChange(sku: val),
          ),
          const SizedBox(height: Spacing.md),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Quantity',
                  style: AppTypography.label.copyWith(color: colors.textSecondary)),
              QuantityStepper(
                value: widget.item.quantity,
                unit: widget.item.unit,
                onChanged: (qty) => _emitChange(quantity: qty),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
