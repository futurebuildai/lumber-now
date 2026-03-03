import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';
import '../../utils/api_error.dart';
import '../../widgets/common/dealer_logo.dart';
import '../../widgets/feedback/error_card.dart';
import '../../widgets/forms/validated_email_field.dart';
import '../../widgets/forms/validated_password_field.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen>
    with SingleTickerProviderStateMixin {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  bool _loading = false;
  String? _error;

  late final AnimationController _animController;
  late final Animation<double> _fadeAnimation;
  late final Animation<Offset> _slideAnimation;

  @override
  void initState() {
    super.initState();
    _animController = AnimationController(
      vsync: this,
      duration: AppDurations.slower,
    );
    _fadeAnimation = Tween<double>(begin: 0, end: 1).animate(
      CurvedAnimation(parent: _animController, curve: Curves.easeIn),
    );
    _slideAnimation = Tween<Offset>(
      begin: const Offset(0, 0.1),
      end: Offset.zero,
    ).animate(CurvedAnimation(parent: _animController, curve: Curves.easeOutCubic));
    _animController.forward();
  }

  @override
  void dispose() {
    _animController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _login() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await ref
          .read(currentUserProvider.notifier)
          .login(_emailController.text.trim(), _passwordController.text);
      if (mounted) context.go('/home');
    } catch (e) {
      final msg = e is ApiError ? e.message : 'Invalid email or password';
      setState(() => _error = msg);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;
    final tenantConfig = ref.watch(tenantConfigProvider);
    final config = tenantConfig.valueOrNull;

    return Scaffold(
      backgroundColor: colors.background,
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(Spacing.xl),
            child: FadeTransition(
              opacity: _fadeAnimation,
              child: SlideTransition(
                position: _slideAnimation,
                child: Form(
                  key: _formKey,
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      // Dealer logo
                      Center(
                        child: Container(
                          width: 80,
                          height: 80,
                          decoration: BoxDecoration(
                            color: colors.primary,
                            borderRadius: Radii.borderLg,
                          ),
                          child: Center(
                            child: DealerLogo(
                              logoUrl: config?.logoUrl,
                              size: 48,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(height: Spacing.xl),

                      Semantics(
                        header: true,
                        child: Text(
                          config?.name ?? 'LumberNow',
                          style: AppTypography.headline.copyWith(
                            color: colors.textPrimary,
                          ),
                          textAlign: TextAlign.center,
                        ),
                      ),
                      const SizedBox(height: Spacing.sm),
                      Text(
                        'Sign in to your account',
                        style: AppTypography.body.copyWith(
                          color: colors.textSecondary,
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: Spacing.xxxl),

                      if (_error != null) ...[
                        ErrorCard(message: _error!),
                        const SizedBox(height: Spacing.lg),
                      ],

                      ValidatedEmailField(
                        controller: _emailController,
                        autofocus: true,
                      ),
                      const SizedBox(height: Spacing.lg),

                      ValidatedPasswordField(
                        controller: _passwordController,
                        textInputAction: TextInputAction.done,
                      ),
                      const SizedBox(height: Spacing.xl),

                      Semantics(
                        button: true,
                        label: 'Sign in',
                        child: FilledButton(
                          onPressed: _loading ? null : _login,
                          child: _loading
                              ? const SizedBox(
                                  height: 20,
                                  width: 20,
                                  child: CircularProgressIndicator(
                                      strokeWidth: 2, color: Colors.white),
                                )
                              : const Text('Sign In'),
                        ),
                      ),
                      const SizedBox(height: Spacing.lg),
                      TextButton(
                        onPressed: () => context.go('/register'),
                        child: const Text("Don't have an account? Register"),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
