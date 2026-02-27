typedef Validator = String? Function(String?);

abstract class Validators {
  static Validator required(String fieldName) => (value) {
        if (value == null || value.trim().isEmpty) {
          return '$fieldName is required';
        }
        return null;
      };

  static Validator email() => (value) {
        if (value == null || value.trim().isEmpty) {
          return 'Email is required';
        }
        final emailRegex = RegExp(r'^[^@\s]+@[^@\s]+\.[^@\s]+$');
        if (!emailRegex.hasMatch(value.trim())) {
          return 'Enter a valid email address';
        }
        return null;
      };

  static Validator password({int minLength = 8}) => (value) {
        if (value == null || value.isEmpty) {
          return 'Password is required';
        }
        if (value.length < minLength) {
          return 'Password must be at least $minLength characters';
        }
        return null;
      };

  static Validator minLength(int length, String fieldName) => (value) {
        if (value != null && value.length < length) {
          return '$fieldName must be at least $length characters';
        }
        return null;
      };

  static Validator quantity() => (value) {
        if (value == null || value.trim().isEmpty) {
          return 'Quantity is required';
        }
        final qty = double.tryParse(value);
        if (qty == null || qty <= 0) {
          return 'Enter a valid quantity';
        }
        return null;
      };

  static Validator confirmPassword(String Function() getPassword) => (value) {
        if (value == null || value.isEmpty) {
          return 'Please confirm your password';
        }
        if (value != getPassword()) {
          return 'Passwords do not match';
        }
        return null;
      };

  static Validator compose(List<Validator> validators) => (value) {
        for (final validator in validators) {
          final error = validator(value);
          if (error != null) return error;
        }
        return null;
      };

  static double passwordStrength(String password) {
    if (password.isEmpty) return 0;
    double score = 0;
    if (password.length >= 8) score += 0.25;
    if (password.length >= 12) score += 0.15;
    if (RegExp(r'[a-z]').hasMatch(password)) score += 0.15;
    if (RegExp(r'[A-Z]').hasMatch(password)) score += 0.15;
    if (RegExp(r'[0-9]').hasMatch(password)) score += 0.15;
    if (RegExp(r'[!@#$%^&*(),.?":{}|<>]').hasMatch(password)) score += 0.15;
    return score.clamp(0.0, 1.0);
  }
}
