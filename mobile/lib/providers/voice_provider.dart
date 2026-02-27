import 'dart:async';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/voice_service.dart';

enum VoiceState { idle, recording, paused, done }

class VoiceRecordingState {
  final VoiceState state;
  final Duration elapsed;
  final String? filePath;

  const VoiceRecordingState({
    this.state = VoiceState.idle,
    this.elapsed = Duration.zero,
    this.filePath,
  });

  VoiceRecordingState copyWith({
    VoiceState? state,
    Duration? elapsed,
    String? filePath,
  }) {
    return VoiceRecordingState(
      state: state ?? this.state,
      elapsed: elapsed ?? this.elapsed,
      filePath: filePath ?? this.filePath,
    );
  }
}

class VoiceNotifier extends StateNotifier<VoiceRecordingState> {
  final VoiceService _service;
  Timer? _timer;

  VoiceNotifier(this._service) : super(const VoiceRecordingState());

  VoiceService get service => _service;

  Future<void> startRecording() async {
    await _service.startRecording();
    state = state.copyWith(state: VoiceState.recording, elapsed: Duration.zero);
    _startTimer();
  }

  Future<void> stopRecording() async {
    _stopTimer();
    final path = await _service.stopRecording();
    state = state.copyWith(state: VoiceState.done, filePath: path);
    await _service.preparePlayer();
  }

  Future<void> pauseRecording() async {
    _stopTimer();
    await _service.pauseRecording();
    state = state.copyWith(state: VoiceState.paused);
  }

  void reset() {
    _stopTimer();
    _service.reset();
    state = const VoiceRecordingState();
  }

  void _startTimer() {
    _timer = Timer.periodic(const Duration(seconds: 1), (_) {
      state = state.copyWith(
        elapsed: state.elapsed + const Duration(seconds: 1),
      );
    });
  }

  void _stopTimer() {
    _timer?.cancel();
    _timer = null;
  }

  @override
  void dispose() {
    _stopTimer();
    _service.dispose();
    super.dispose();
  }
}

final voiceServiceProvider = Provider<VoiceService>((ref) => VoiceService());

final voiceProvider =
    StateNotifierProvider<VoiceNotifier, VoiceRecordingState>((ref) {
  return VoiceNotifier(ref.read(voiceServiceProvider));
});
