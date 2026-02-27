import 'dart:io';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';

class ImagePickerButton extends StatelessWidget {
  final void Function(File file) onPicked;

  const ImagePickerButton({super.key, required this.onPicked});

  Future<void> _pick(BuildContext context, ImageSource source) async {
    final picker = ImagePicker();
    final result = await picker.pickImage(source: source);
    if (result != null) {
      onPicked(File(result.path));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: OutlinedButton.icon(
            onPressed: () => _pick(context, ImageSource.camera),
            icon: const Icon(Icons.camera_alt),
            label: const Text('Camera'),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: OutlinedButton.icon(
            onPressed: () => _pick(context, ImageSource.gallery),
            icon: const Icon(Icons.photo_library),
            label: const Text('Gallery'),
          ),
        ),
      ],
    );
  }
}
