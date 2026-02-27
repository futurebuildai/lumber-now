import 'package:flutter/material.dart';
import '../../utils/validators.dart';
import 'validated_text_field.dart';

class ValidatedEmailField extends StatelessWidget {
  final TextEditingController controller;
  final bool enabled;
  final bool autofocus;
  final TextInputAction textInputAction;

  const ValidatedEmailField({
    super.key,
    required this.controller,
    this.enabled = true,
    this.autofocus = false,
    this.textInputAction = TextInputAction.next,
  });

  @override
  Widget build(BuildContext context) {
    return ValidatedTextField(
      controller: controller,
      label: 'Email',
      hint: 'you@example.com',
      validator: Validators.email(),
      keyboardType: TextInputType.emailAddress,
      textInputAction: textInputAction,
      prefixIcon: const Icon(Icons.email_outlined),
      enabled: enabled,
      autofocus: autofocus,
      semanticLabel: 'Email address',
    );
  }
}
