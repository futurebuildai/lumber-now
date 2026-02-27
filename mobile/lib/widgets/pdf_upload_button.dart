import 'dart:io';
import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';

class PDFUploadButton extends StatelessWidget {
  final void Function(File file) onPicked;

  const PDFUploadButton({super.key, required this.onPicked});

  Future<void> _pick() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['pdf'],
    );
    if (result != null && result.files.single.path != null) {
      onPicked(File(result.files.single.path!));
    }
  }

  @override
  Widget build(BuildContext context) {
    return OutlinedButton.icon(
      onPressed: _pick,
      icon: const Icon(Icons.picture_as_pdf),
      label: const Text('Upload PDF'),
    );
  }
}
