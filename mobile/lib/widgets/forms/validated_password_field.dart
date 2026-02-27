import 'package:flutter/material.dart';
import '../../theme/app_theme.dart';
import '../../theme/design_tokens.dart';
import '../../utils/validators.dart';

class ValidatedPasswordField extends StatefulWidget {
  final TextEditingController controller;
  final String label;
  final String? Function(String?)? validator;
  final bool showStrengthIndicator;
  final bool enabled;
  final TextInputAction textInputAction;

  const ValidatedPasswordField({
    super.key,
    required this.controller,
    this.label = 'Password',
    this.validator,
    this.showStrengthIndicator = false,
    this.enabled = true,
    this.textInputAction = TextInputAction.done,
  });

  @override
  State<ValidatedPasswordField> createState() => _ValidatedPasswordFieldState();
}

class _ValidatedPasswordFieldState extends State<ValidatedPasswordField> {
  bool _obscured = true;

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Semantics(
          label: widget.label,
          textField: true,
          child: TextFormField(
            controller: widget.controller,
            decoration: InputDecoration(
              labelText: widget.label,
              prefixIcon: const Icon(Icons.lock_outlined),
              suffixIcon: IconButton(
                icon: Icon(
                  _obscured ? Icons.visibility_outlined : Icons.visibility_off_outlined,
                  semanticLabel: _obscured ? 'Show password' : 'Hide password',
                ),
                onPressed: () => setState(() => _obscured = !_obscured),
                tooltip: _obscured ? 'Show password' : 'Hide password',
              ),
            ),
            obscureText: _obscured,
            validator: widget.validator ?? Validators.password(),
            enabled: widget.enabled,
            textInputAction: widget.textInputAction,
            autovalidateMode: AutovalidateMode.onUserInteraction,
          ),
        ),
        if (widget.showStrengthIndicator) ...[
          const SizedBox(height: Spacing.sm),
          ValueListenableBuilder<TextEditingValue>(
            valueListenable: widget.controller,
            builder: (_, value, __) {
              final strength = Validators.passwordStrength(value.text);
              final Color barColor;
              final String label;
              if (strength < 0.3) {
                barColor = colors.error;
                label = 'Weak';
              } else if (strength < 0.6) {
                barColor = colors.warning;
                label = 'Fair';
              } else if (strength < 0.85) {
                barColor = colors.confidenceMedium;
                label = 'Good';
              } else {
                barColor = colors.success;
                label = 'Strong';
              }

              if (value.text.isEmpty) return const SizedBox.shrink();

              return Row(
                children: [
                  Expanded(
                    child: ClipRRect(
                      borderRadius: Radii.borderFull,
                      child: LinearProgressIndicator(
                        value: strength,
                        backgroundColor: colors.borderLight,
                        color: barColor,
                        minHeight: 4,
                      ),
                    ),
                  ),
                  const SizedBox(width: Spacing.sm),
                  Text(label,
                      style: TextStyle(fontSize: 12, color: barColor, fontWeight: FontWeight.w500)),
                ],
              );
            },
          ),
        ],
      ],
    );
  }
}
