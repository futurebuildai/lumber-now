class Env {
  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://10.0.2.2:8080/v1',
  );

  static const String tenantSlug = String.fromEnvironment(
    'TENANT_SLUG',
    defaultValue: '',
  );

  static const String primaryColor = String.fromEnvironment(
    'PRIMARY_COLOR',
    defaultValue: '1E40AF',
  );

  static const String secondaryColor = String.fromEnvironment(
    'SECONDARY_COLOR',
    defaultValue: '1E3A5F',
  );

  static const String appName = String.fromEnvironment(
    'APP_NAME',
    defaultValue: 'LumberNow',
  );
}
