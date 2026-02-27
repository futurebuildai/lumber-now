import 'package:flutter/material.dart';
import '../../models/models.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';
import '../../utils/validators.dart';
import '../forms/validated_text_field.dart';

class AddItemSheet extends StatefulWidget {
  final ValueChanged<StructuredItem> onAdd;

  const AddItemSheet({super.key, required this.onAdd});

  @override
  State<AddItemSheet> createState() => _AddItemSheetState();
}

class _AddItemSheetState extends State<AddItemSheet> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _qtyController = TextEditingController(text: '1');
  final _unitController = TextEditingController(text: 'EA');
  final _skuController = TextEditingController();

  @override
  void dispose() {
    _nameController.dispose();
    _qtyController.dispose();
    _unitController.dispose();
    _skuController.dispose();
    super.dispose();
  }

  void _submit() {
    if (!_formKey.currentState!.validate()) return;
    widget.onAdd(StructuredItem(
      name: _nameController.text.trim(),
      quantity: double.parse(_qtyController.text),
      unit: _unitController.text.trim(),
      sku: _skuController.text.trim(),
      confidence: 1.0,
      matched: false,
    ));
    Navigator.of(context).pop();
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    return Padding(
      padding: EdgeInsets.only(
        left: Spacing.lg,
        right: Spacing.lg,
        top: Spacing.lg,
        bottom: MediaQuery.of(context).viewInsets.bottom + Spacing.lg,
      ),
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: colors.border,
                  borderRadius: Radii.borderFull,
                ),
              ),
            ),
            const SizedBox(height: Spacing.lg),
            Text('Add Item',
                style: AppTypography.title.copyWith(color: colors.textPrimary)),
            const SizedBox(height: Spacing.lg),
            ValidatedTextField(
              controller: _nameController,
              label: 'Item Name',
              validator: Validators.required('Item name'),
              autofocus: true,
            ),
            const SizedBox(height: Spacing.md),
            Row(
              children: [
                Expanded(
                  flex: 2,
                  child: ValidatedTextField(
                    controller: _qtyController,
                    label: 'Quantity',
                    validator: Validators.quantity(),
                    keyboardType: TextInputType.number,
                  ),
                ),
                const SizedBox(width: Spacing.md),
                Expanded(
                  child: ValidatedTextField(
                    controller: _unitController,
                    label: 'Unit',
                    validator: Validators.required('Unit'),
                  ),
                ),
              ],
            ),
            const SizedBox(height: Spacing.md),
            ValidatedTextField(
              controller: _skuController,
              label: 'SKU (optional)',
              textInputAction: TextInputAction.done,
            ),
            const SizedBox(height: Spacing.xl),
            FilledButton(
              onPressed: _submit,
              child: const Text('Add Item'),
            ),
          ],
        ),
      ),
    );
  }
}
