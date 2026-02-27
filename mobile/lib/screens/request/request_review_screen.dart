import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../models/models.dart';
import '../../providers/providers.dart';
import '../../theme/app_theme.dart';
import '../../theme/app_typography.dart';
import '../../theme/design_tokens.dart';
import '../../utils/api_error.dart';
import '../../utils/haptics.dart';
import '../../widgets/common/section_header.dart';
import '../../widgets/feedback/error_card.dart';
import '../../widgets/feedback/success_animation.dart';
import '../../widgets/feedback/toast_overlay.dart';
import '../../widgets/loading/shimmer_list.dart';
import '../../widgets/request/add_item_sheet.dart';
import '../../widgets/request/confidence_indicator.dart';
import '../../widgets/request/editable_sku_card.dart';
import '../../widgets/request/status_timeline.dart';

class RequestReviewScreen extends ConsumerStatefulWidget {
  final String requestId;

  const RequestReviewScreen({super.key, required this.requestId});

  @override
  ConsumerState<RequestReviewScreen> createState() =>
      _RequestReviewScreenState();
}

class _RequestReviewScreenState extends ConsumerState<RequestReviewScreen> {
  MaterialRequest? _request;
  List<StructuredItem> _editedItems = [];
  bool _loading = true;
  bool _processing = false;
  bool _showSuccess = false;
  String? _error;
  final _notesController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _loadRequest();
  }

  @override
  void dispose() {
    _notesController.dispose();
    super.dispose();
  }

  Future<void> _loadRequest() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final api = ref.read(apiClientProvider);
      final response = await api.dio.get('/requests/${widget.requestId}');
      final req = MaterialRequest.fromJson(response.data);
      setState(() {
        _request = req;
        _editedItems = List.from(req.structuredItems);
        _notesController.text = req.notes;
        _loading = false;
      });
    } catch (e) {
      final msg = e is ApiError ? e.message : 'Failed to load request';
      setState(() {
        _error = msg;
        _loading = false;
      });
    }
  }

  Future<void> _process() async {
    setState(() {
      _processing = true;
      _error = null;
    });
    try {
      final api = ref.read(apiClientProvider);
      await api.dio.post('/requests/${widget.requestId}/process');
      Haptics.success();
      await _loadRequest();
    } catch (e) {
      final msg = e is ApiError ? e.message : 'Processing failed';
      setState(() => _error = msg);
      Haptics.error();
    } finally {
      setState(() => _processing = false);
    }
  }

  Future<void> _confirm() async {
    setState(() {
      _processing = true;
      _error = null;
    });
    try {
      final api = ref.read(apiClientProvider);
      await api.dio.post('/requests/${widget.requestId}/confirm', data: {
        'items': _editedItems.map((e) => e.toJson()).toList(),
        'notes': _notesController.text,
      });
      Haptics.success();
      await _loadRequest();
    } catch (e) {
      final msg = e is ApiError ? e.message : 'Confirmation failed';
      setState(() => _error = msg);
      Haptics.error();
    } finally {
      setState(() => _processing = false);
    }
  }

  Future<void> _send() async {
    setState(() {
      _processing = true;
      _error = null;
    });
    try {
      final api = ref.read(apiClientProvider);
      await api.dio.post('/requests/${widget.requestId}/send');
      Haptics.success();
      setState(() => _showSuccess = true);
    } catch (e) {
      final msg = e is ApiError ? e.message : 'Send failed';
      setState(() => _error = msg);
      Haptics.error();
      setState(() => _processing = false);
    }
  }

  void _addItem(StructuredItem item) {
    setState(() {
      _editedItems = [..._editedItems, item];
    });
    Haptics.success();
  }

  void _updateItem(int index, StructuredItem item) {
    setState(() {
      _editedItems = List.from(_editedItems)..[index] = item;
    });
  }

  void _deleteItem(int index) {
    setState(() {
      _editedItems = List.from(_editedItems)..removeAt(index);
    });
    Haptics.medium();
    if (mounted) {
      ToastOverlay.show(context,
          message: 'Item removed', type: ToastType.info);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = AppTheme.colors;

    if (_showSuccess) {
      return SuccessAnimation(
        message: 'Request Sent!',
        onComplete: () {
          if (mounted) context.go('/home');
        },
      );
    }

    if (_loading) {
      return const ShimmerList(itemCount: 5, itemHeight: 80);
    }

    if (_request == null) {
      return ErrorCard(
        message: _error ?? 'Request not found',
        onRetry: _loadRequest,
      );
    }

    final req = _request!;
    final overallConfidence = _editedItems.isEmpty
        ? 0.0
        : _editedItems.fold<double>(0, (sum, i) => sum + i.confidence) /
            _editedItems.length;

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // Status timeline
          StatusTimeline(currentStatus: req.status),
          const SizedBox(height: Spacing.lg),

          if (_error != null) ...[
            ErrorCard(message: _error!, onRetry: () => setState(() => _error = null)),
            const SizedBox(height: Spacing.lg),
          ],

          // Original request card
          if (req.rawText.isNotEmpty) ...[
            SectionHeader(title: 'Original Request'),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(Spacing.lg),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Icon(_inputIcon(req.inputType),
                        color: colors.primary, size: IconSizes.sm),
                    const SizedBox(width: Spacing.md),
                    Expanded(
                      child: Text(
                        req.rawText,
                        style: AppTypography.body.copyWith(color: colors.textPrimary),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: Spacing.xl),
          ],

          // Matched items section
          if (_editedItems.isNotEmpty) ...[
            SectionHeader(
              title: 'Matched Items',
              count: _editedItems.length,
              trailing: req.status == 'parsed'
                  ? TextButton.icon(
                      onPressed: () {
                        showModalBottomSheet(
                          context: context,
                          isScrollControlled: true,
                          builder: (_) => AddItemSheet(onAdd: _addItem),
                        );
                      },
                      icon: const Icon(Icons.add, size: 18),
                      label: const Text('Add'),
                    )
                  : null,
            ),
            ...List.generate(_editedItems.length, (index) {
              return EditableSKUCard(
                key: ValueKey('${_editedItems[index].name}-$index'),
                item: _editedItems[index],
                onChanged: (item) => _updateItem(index, item),
                onDelete: () => _deleteItem(index),
              );
            }),
            const SizedBox(height: Spacing.lg),

            // Overall confidence
            Center(
              child: ConfidenceIndicator(
                confidence: overallConfidence,
                size: 64,
              ),
            ),
            const SizedBox(height: Spacing.xl),
          ],

          // Notes
          if (req.status == 'parsed' || req.status == 'confirmed') ...[
            SectionHeader(title: 'Notes'),
            TextFormField(
              controller: _notesController,
              maxLines: 3,
              decoration: const InputDecoration(
                hintText: 'Add any notes or special instructions...',
              ),
            ),
            const SizedBox(height: Spacing.xl),
          ],

          // Action buttons
          if (req.status == 'pending')
            Semantics(
              button: true,
              label: 'Process request with AI',
              child: FilledButton.icon(
                onPressed: _processing ? null : _process,
                icon: const Icon(Icons.psychology_rounded),
                label: _processing
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                            strokeWidth: 2, color: Colors.white))
                    : const Text('Process with AI'),
              ),
            ),

          if (req.status == 'parsed')
            Semantics(
              button: true,
              label: 'Confirm edited items',
              child: FilledButton.icon(
                onPressed: _processing ? null : _confirm,
                icon: const Icon(Icons.check_circle_rounded),
                label: _processing
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                            strokeWidth: 2, color: Colors.white))
                    : const Text('Confirm Order'),
              ),
            ),

          if (req.status == 'confirmed')
            Semantics(
              button: true,
              label: 'Send to dealer',
              child: FilledButton.icon(
                onPressed: _processing ? null : _send,
                icon: const Icon(Icons.send_rounded),
                label: _processing
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                            strokeWidth: 2, color: Colors.white))
                    : const Text('Send to Dealer'),
              ),
            ),

          const SizedBox(height: Spacing.xxl),
        ],
      ),
    );
  }

  IconData _inputIcon(String inputType) {
    switch (inputType) {
      case 'voice':
        return Icons.mic_rounded;
      case 'image':
        return Icons.camera_alt_rounded;
      case 'pdf':
        return Icons.picture_as_pdf_rounded;
      default:
        return Icons.edit_rounded;
    }
  }
}
