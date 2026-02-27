import 'dart:io';
import 'package:dio/dio.dart';
import 'api_client.dart';

class MediaService {
  final ApiClient _api;

  MediaService(this._api);

  Future<String> uploadFile(File file) async {
    final formData = FormData.fromMap({
      'file': await MultipartFile.fromFile(file.path,
          filename: file.path.split('/').last),
    });

    final response = await _api.dio.post(
      '/media/upload',
      data: formData,
      options: Options(headers: {'Content-Type': 'multipart/form-data'}),
    );
    return response.data['key'] as String;
  }
}
