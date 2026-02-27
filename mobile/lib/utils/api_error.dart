import 'package:dio/dio.dart';

class ApiError implements Exception {
  final String message;
  final bool isRetryable;
  final int? statusCode;

  const ApiError({
    required this.message,
    this.isRetryable = false,
    this.statusCode,
  });

  factory ApiError.fromDioException(DioException e) {
    switch (e.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return const ApiError(
          message: 'Connection timed out. Please check your internet and try again.',
          isRetryable: true,
        );

      case DioExceptionType.connectionError:
        return const ApiError(
          message: 'Unable to connect to server. Please check your connection.',
          isRetryable: true,
        );

      case DioExceptionType.badResponse:
        return _fromResponse(e.response);

      case DioExceptionType.cancel:
        return const ApiError(
          message: 'Request was cancelled.',
          isRetryable: false,
        );

      default:
        return const ApiError(
          message: 'Something went wrong. Please try again.',
          isRetryable: true,
        );
    }
  }

  static ApiError _fromResponse(Response? response) {
    final statusCode = response?.statusCode;
    final data = response?.data;

    String message = 'Something went wrong.';
    if (data is Map<String, dynamic>) {
      message = data['error'] as String? ??
          data['message'] as String? ??
          message;
    }

    switch (statusCode) {
      case 400:
        return ApiError(
          message: message,
          isRetryable: false,
          statusCode: 400,
        );
      case 401:
        return const ApiError(
          message: 'Your session has expired. Please sign in again.',
          isRetryable: false,
          statusCode: 401,
        );
      case 403:
        return const ApiError(
          message: "You don't have permission to perform this action.",
          isRetryable: false,
          statusCode: 403,
        );
      case 404:
        return const ApiError(
          message: 'The requested resource was not found.',
          isRetryable: false,
          statusCode: 404,
        );
      case 409:
        return ApiError(
          message: message,
          isRetryable: false,
          statusCode: 409,
        );
      case 422:
        return ApiError(
          message: message,
          isRetryable: false,
          statusCode: 422,
        );
      case 429:
        return const ApiError(
          message: 'Too many requests. Please wait a moment and try again.',
          isRetryable: true,
          statusCode: 429,
        );
      case 500:
      case 502:
      case 503:
        return const ApiError(
          message: 'Server error. Please try again later.',
          isRetryable: true,
          statusCode: 500,
        );
      default:
        return ApiError(
          message: message,
          isRetryable: true,
          statusCode: statusCode,
        );
    }
  }

  @override
  String toString() => message;
}
