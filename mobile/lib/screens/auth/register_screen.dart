import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';
import '../../utils/api_error.dart';
import '../../utils/validators.dart';
import '../../widgets/feedback/error_card.dart';
import '../../widgets/forms/validated_email_field.dart';
import '../../widgets/forms/validated_password_field.dart';
import '../../widgets/forms/validated_text_field.dart';

class RegisterScreen extends ConsumerStatefulWidget {
  const RegisterScreen({super.key});

  @override
  ConsumerState<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends ConsumerState<RegisterScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _nameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    super.dispose();
  }

  Future<void> _register() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await ref.read(authServiceProvider).register(
            _emailController.text.trim(),
            _passwordController.text,
            _nameController.text.trim(),
          );
      if (mounted) context.go('/login');
    } catch (e) {
      final msg =
          e is ApiError ? e.message : 'Registration failed. Try a different email.';
      setState(() => _error = msg);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    return Scaffold(
      backgroundColor: colors.background,
      appBar: AppBar(
        title: const Text('Create Account'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.go('/login'),
          tooltip: 'Back to login',
        ),
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(Spacing.xl),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Semantics(
                  header: true,
                  child: Text(
                    'Get Started',
                    style: AppTypography.headline.copyWith(color: colors.textPrimary),
                  ),
                ),
                const SizedBox(height: Spacing.sm),
                Text(
                  'Create your contractor account',
                  style: AppTypography.body.copyWith(color: colors.textSecondary),
                ),
                const SizedBox(height: Spacing.xxl),

                if (_error != null) ...[
                  ErrorCard(message: _error!),
                  const SizedBox(height: Spacing.lg),
                ],

                ValidatedTextField(
                  controller: _nameController,
                  label: 'Full Name',
                  validator: Validators.required('Full name'),
                  prefixIcon: const Icon(Icons.person_outlined),
                  autofocus: true,
                  semanticLabel: 'Full name',
                ),
                const SizedBox(height: Spacing.lg),

                ValidatedEmailField(controller: _emailController),
                const SizedBox(height: Spacing.lg),

                ValidatedPasswordField(
                  controller: _passwordController,
                  showStrengthIndicator: true,
                  textInputAction: TextInputAction.next,
                ),
                const SizedBox(height: Spacing.lg),

                ValidatedPasswordField(
                  controller: _confirmPasswordController,
                  label: 'Confirm Password',
                  validator: Validators.confirmPassword(
                      () => _passwordController.text),
                ),
                const SizedBox(height: Spacing.xl),

                Semantics(
                  button: true,
                  label: 'Create account',
                  child: FilledButton(
                    onPressed: _loading ? null : _register,
                    child: _loading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(
                                strokeWidth: 2, color: Colors.white),
                          )
                        : const Text('Create Account'),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
