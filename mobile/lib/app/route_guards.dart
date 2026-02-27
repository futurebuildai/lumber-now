import 'package:flutter_secure_storage/flutter_secure_storage.dart';

const _storage = FlutterSecureStorage();

Future<String?> authRedirect(String currentPath) async {
  final token = await _storage.read(key: 'access_token');
  final isLoggedIn = token != null && token.isNotEmpty;
  final isAuthRoute = currentPath == '/login' || currentPath == '/register';

  if (!isLoggedIn && !isAuthRoute) {
    return '/login';
  }
  if (isLoggedIn && isAuthRoute) {
    return '/home';
  }
  return null;
}
