import 'package:intl/intl.dart';

abstract class Formatters {
  static String date(String isoDate) {
    try {
      final dt = DateTime.parse(isoDate);
      return DateFormat('MMM d, yyyy').format(dt);
    } catch (_) {
      return isoDate;
    }
  }

  static String dateTime(String isoDate) {
    try {
      final dt = DateTime.parse(isoDate);
      return DateFormat('MMM d, yyyy h:mm a').format(dt);
    } catch (_) {
      return isoDate;
    }
  }

  static String relativeTime(String isoDate) {
    try {
      final dt = DateTime.parse(isoDate);
      final now = DateTime.now();
      final diff = now.difference(dt);

      if (diff.inSeconds < 60) return 'just now';
      if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
      if (diff.inHours < 24) return '${diff.inHours}h ago';
      if (diff.inDays < 7) return '${diff.inDays}d ago';
      if (diff.inDays < 30) return '${(diff.inDays / 7).floor()}w ago';
      return date(isoDate);
    } catch (_) {
      return isoDate;
    }
  }

  static String quantity(double qty, String unit) {
    final formatted = qty == qty.roundToDouble()
        ? qty.toInt().toString()
        : qty.toStringAsFixed(1);
    return '$formatted $unit';
  }

  static String confidence(double value) {
    return '${(value * 100).toInt()}%';
  }

  static String initials(String fullName) {
    final parts = fullName.trim().split(RegExp(r'\s+'));
    if (parts.isEmpty) return '?';
    if (parts.length == 1) return parts[0][0].toUpperCase();
    return '${parts[0][0]}${parts[parts.length - 1][0]}'.toUpperCase();
  }
}
