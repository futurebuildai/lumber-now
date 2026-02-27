import 'package:flutter/material.dart';
import '../../theme/design_tokens.dart';

class ValidatedTextField extends StatelessWidget {
  final TextEditingController controller;
  final String label;
  final String? hint;
  final String? Function(String?)? validator;
  final int maxLines;
  final TextInputType keyboardType;
  final TextInputAction textInputAction;
  final Widget? prefixIcon;
  final Widget? suffixIcon;
  final bool enabled;
  final bool autofocus;
  final ValueChanged<String>? onChanged;
  final String? semanticLabel;

  const ValidatedTextField({
    super.key,
    required this.controller,
    required this.label,
    this.hint,
    this.validator,
    this.maxLines = 1,
    this.keyboardType = TextInputType.text,
    this.textInputAction = TextInputAction.next,
    this.prefixIcon,
    this.suffixIcon,
    this.enabled = true,
    this.autofocus = false,
    this.onChanged,
    this.semanticLabel,
  });

  @override
  Widget build(BuildContext context) {
    return Semantics(
      label: semanticLabel ?? label,
      textField: true,
      child: TextFormField(
        controller: controller,
        decoration: InputDecoration(
          labelText: label,
          hintText: hint,
          prefixIcon: prefixIcon,
          suffixIcon: suffixIcon,
        ),
        validator: validator,
        maxLines: maxLines,
        keyboardType: keyboardType,
        textInputAction: textInputAction,
        enabled: enabled,
        autofocus: autofocus,
        onChanged: onChanged,
        autovalidateMode: AutovalidateMode.onUserInteraction,
      ),
    );
  }
}
