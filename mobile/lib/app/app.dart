import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../animations/page_transitions.dart';
import '../config/env.dart';
import '../providers/providers.dart';
import '../screens/auth/login_screen.dart';
import '../screens/auth/register_screen.dart';
import '../screens/home/home_screen.dart';
import '../screens/request/new_request_screen.dart';
import '../screens/request/request_review_screen.dart';
import '../screens/history/request_history_screen.dart';
import '../screens/splash/splash_screen.dart';
import '../theme/app_colors.dart';
import '../theme/app_theme.dart';
import '../widgets/branded_scaffold.dart';
import 'route_guards.dart';

final _shellNavigatorKey = GlobalKey<NavigatorState>();

final _router = GoRouter(
  initialLocation: '/splash',
  redirect: (context, state) => authRedirect(state.matchedLocation),
  routes: [
    GoRoute(
      path: '/splash',
      pageBuilder: (context, state) => fadeThrough(state, const SplashScreen()),
    ),
    GoRoute(
      path: '/login',
      pageBuilder: (context, state) => fadeThrough(state, const LoginScreen()),
    ),
    GoRoute(
      path: '/register',
      pageBuilder: (context, state) => slideUp(state, const RegisterScreen()),
    ),
    ShellRoute(
      navigatorKey: _shellNavigatorKey,
      builder: (context, state, child) => BrandedScaffold(child: child),
      routes: [
        GoRoute(
          path: '/home',
          pageBuilder: (context, state) => fadeThrough(state, const HomeScreen()),
        ),
        GoRoute(
          path: '/request/new',
          pageBuilder: (context, state) =>
              slideUp(state, const NewRequestScreen()),
        ),
        GoRoute(
          path: '/request/:id',
          pageBuilder: (context, state) => slideUp(
            state,
            RequestReviewScreen(
              requestId: state.pathParameters['id']!,
            ),
          ),
        ),
        GoRoute(
          path: '/history',
          pageBuilder: (context, state) =>
              fadeThrough(state, const RequestHistoryScreen()),
        ),
      ],
    ),
  ],
);

class LumberNowApp extends ConsumerWidget {
  const LumberNowApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tenantConfig = ref.watch(tenantConfigProvider);

    final primaryColor = tenantConfig.whenOrNull(
          data: (config) =>
              config != null ? AppColors.parseHex(config.primaryColor) : null,
        ) ??
        AppColors.parseHex(Env.primaryColor);

    final secondaryColor = tenantConfig.whenOrNull(
          data: (config) =>
              config != null ? AppColors.parseHex(config.secondaryColor) : null,
        ) ??
        AppColors.parseHex(Env.secondaryColor);

    final appName = tenantConfig.whenOrNull(
          data: (config) => config?.name,
        ) ??
        Env.appName;

    return MaterialApp.router(
      title: appName,
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light(primaryColor, secondaryColor),
      routerConfig: _router,
    );
  }
}
