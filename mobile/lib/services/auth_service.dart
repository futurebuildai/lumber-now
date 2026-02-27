import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../models/models.dart';
import 'api_client.dart';

class AuthService {
  final ApiClient _api;
  final FlutterSecureStorage _storage = const FlutterSecureStorage();

  AuthService(this._api);

  Future<TokenPair> login(String email, String password) async {
    final response = await _api.dio.post('/auth/login', data: {
      'email': email,
      'password': password,
    });
    final tokens = TokenPair.fromJson(response.data);
    await _storage.write(key: 'access_token', value: tokens.accessToken);
    await _storage.write(key: 'refresh_token', value: tokens.refreshToken);
    return tokens;
  }

  Future<void> register(
      String email, String password, String fullName) async {
    await _api.dio.post('/auth/register', data: {
      'email': email,
      'password': password,
      'full_name': fullName,
      'role': 'contractor',
    });
  }

  Future<User> getMe() async {
    final response = await _api.dio.get('/auth/me');
    return User.fromJson(response.data);
  }

  Future<void> logout() async {
    await _storage.deleteAll();
  }

  Future<bool> isLoggedIn() async {
    final token = await _storage.read(key: 'access_token');
    return token != null;
  }
}
