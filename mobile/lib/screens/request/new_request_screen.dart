import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:file_picker/file_picker.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';
import '../../utils/api_error.dart';
import '../../utils/haptics.dart';
import '../../utils/validators.dart';
import '../../widgets/feedback/error_card.dart';
import '../../widgets/feedback/toast_overlay.dart';
import '../../widgets/forms/validated_text_field.dart';
import '../../widgets/request/media_preview.dart';
import '../../widgets/voice_recorder.dart';

class NewRequestScreen extends ConsumerStatefulWidget {
  const NewRequestScreen({super.key});

  @override
  ConsumerState<NewRequestScreen> createState() => _NewRequestScreenState();
}

class _NewRequestScreenState extends ConsumerState<NewRequestScreen> {
  final _textController = TextEditingController();
  String _inputType = 'text';
  bool _submitting = false;
  String? _error;
  File? _selectedFile;
  double _uploadProgress = 0;

  @override
  void dispose() {
    _textController.dispose();
    super.dispose();
  }

  Future<bool> _confirmDiscard() async {
    if (_textController.text.isEmpty && _selectedFile == null) return true;
    final result = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Discard Request?'),
        content: const Text('You have unsaved changes. Are you sure you want to go back?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Discard'),
          ),
        ],
      ),
    );
    return result ?? false;
  }

  Future<void> _pickImage() async {
    final picker = ImagePicker();
    final result = await picker.pickImage(source: ImageSource.camera);
    if (result != null) {
      setState(() {
        _selectedFile = File(result.path);
        _inputType = 'image';
      });
    }
  }

  Future<void> _pickGalleryImage() async {
    final picker = ImagePicker();
    final result = await picker.pickImage(source: ImageSource.gallery);
    if (result != null) {
      setState(() {
        _selectedFile = File(result.path);
        _inputType = 'image';
      });
    }
  }

  Future<void> _pickPDF() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['pdf'],
    );
    if (result != null && result.files.single.path != null) {
      setState(() {
        _selectedFile = File(result.files.single.path!);
        _inputType = 'pdf';
      });
    }
  }

  Future<void> _submit() async {
    if (_inputType == 'text' && _textController.text.trim().isEmpty) {
      setState(() => _error = 'Please describe the materials you need');
      return;
    }
    if ((_inputType == 'image' || _inputType == 'pdf') && _selectedFile == null) {
      setState(() => _error = 'Please select a file to upload');
      return;
    }

    setState(() {
      _submitting = true;
      _error = null;
      _uploadProgress = 0;
    });

    try {
      final api = ref.read(apiClientProvider);
      String? mediaUrl;

      if (_selectedFile != null) {
        final mediaService = ref.read(mediaServiceProvider);
        setState(() => _uploadProgress = 0.3);
        final key = await mediaService.uploadFile(_selectedFile!);
        setState(() => _uploadProgress = 0.8);
        mediaUrl = key;
      }

      final response = await api.dio.post('/requests', data: {
        'input_type': _inputType,
        'raw_text': _textController.text,
        'media_url': mediaUrl ?? '',
      });

      setState(() => _uploadProgress = 1.0);
      Haptics.success();

      if (mounted) {
        final requestId = response.data['id'] as String;
        context.go('/request/$requestId');
      }
    } catch (e) {
      final msg = e is ApiError ? e.message : 'Failed to submit request';
      setState(() => _error = msg);
      Haptics.error();
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, _) async {
        if (didPop) return;
        final shouldDiscard = await _confirmDiscard();
        if (shouldDiscard && context.mounted) context.pop();
      },
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Semantics(
              header: true,
              child: Text(
                'How would you like to submit?',
                style: AppTypography.title.copyWith(color: colors.textPrimary),
              ),
            ),
            const SizedBox(height: Spacing.lg),

            SegmentedButton<String>(
              segments: const [
                ButtonSegment(
                    value: 'text',
                    label: Text('Text'),
                    icon: Icon(Icons.edit_rounded)),
                ButtonSegment(
                    value: 'voice',
                    label: Text('Voice'),
                    icon: Icon(Icons.mic_rounded)),
                ButtonSegment(
                    value: 'image',
                    label: Text('Photo'),
                    icon: Icon(Icons.camera_alt_rounded)),
                ButtonSegment(
                    value: 'pdf',
                    label: Text('PDF'),
                    icon: Icon(Icons.picture_as_pdf_rounded)),
              ],
              selected: {_inputType},
              onSelectionChanged: (val) {
                Haptics.selection();
                setState(() => _inputType = val.first);
              },
            ),

            const SizedBox(height: Spacing.xl),

            if (_error != null) ...[
              ErrorCard(
                message: _error!,
                onRetry: () => setState(() => _error = null),
              ),
              const SizedBox(height: Spacing.lg),
            ],

            // Text input
            if (_inputType == 'text') ...[
              ValidatedTextField(
                controller: _textController,
                label: 'Material List',
                hint:
                    'Describe the materials you need...\n\ne.g., 100 2x4x8 SPF studs, 50 sheets 1/2" OSB',
                maxLines: 8,
                validator: Validators.required('Material description'),
                semanticLabel: 'Material list description',
              ),
            ],

            // Voice input
            if (_inputType == 'voice') ...[
              VoiceRecorder(
                onRecordingComplete: (path) {
                  setState(() {
                    _selectedFile = File(path);
                  });
                },
              ),
              const SizedBox(height: Spacing.lg),
              ValidatedTextField(
                controller: _textController,
                label: 'Additional Notes (optional)',
                hint: 'Any extra details...',
                maxLines: 3,
                semanticLabel: 'Additional notes for voice request',
              ),
            ],

            // Image input
            if (_inputType == 'image') ...[
              if (_selectedFile != null) ...[
                MediaPreview(
                  file: _selectedFile!,
                  type: 'image',
                  onRemove: () => setState(() => _selectedFile = null),
                ),
                const SizedBox(height: Spacing.md),
              ],
              Row(
                children: [
                  Expanded(
                    child: Semantics(
                      button: true,
                      label: 'Take photo with camera',
                      child: OutlinedButton.icon(
                        onPressed: _pickImage,
                        icon: const Icon(Icons.camera_alt_rounded),
                        label: const Text('Camera'),
                      ),
                    ),
                  ),
                  const SizedBox(width: Spacing.sm),
                  Expanded(
                    child: Semantics(
                      button: true,
                      label: 'Choose photo from gallery',
                      child: OutlinedButton.icon(
                        onPressed: _pickGalleryImage,
                        icon: const Icon(Icons.photo_library_rounded),
                        label: const Text('Gallery'),
                      ),
                    ),
                  ),
                ],
              ),
            ],

            // PDF input
            if (_inputType == 'pdf') ...[
              if (_selectedFile != null) ...[
                MediaPreview(
                  file: _selectedFile!,
                  type: 'pdf',
                  onRemove: () => setState(() => _selectedFile = null),
                ),
                const SizedBox(height: Spacing.md),
              ],
              Semantics(
                button: true,
                label: 'Choose PDF file',
                child: OutlinedButton.icon(
                  onPressed: _pickPDF,
                  icon: const Icon(Icons.upload_file_rounded),
                  label: const Text('Choose PDF File'),
                ),
              ),
            ],

            const SizedBox(height: Spacing.xxl),

            // Upload progress
            if (_submitting && _uploadProgress > 0) ...[
              ClipRRect(
                borderRadius: Radii.borderFull,
                child: LinearProgressIndicator(
                  value: _uploadProgress,
                  backgroundColor: colors.borderLight,
                  color: colors.primary,
                  minHeight: 4,
                ),
              ),
              const SizedBox(height: Spacing.md),
            ],

            Semantics(
              button: true,
              label: 'Submit request',
              child: FilledButton(
                onPressed: _submitting ? null : _submit,
                child: _submitting
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                            strokeWidth: 2, color: Colors.white),
                      )
                    : const Text('Submit Request'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
