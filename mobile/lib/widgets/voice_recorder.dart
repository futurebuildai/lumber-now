import 'package:audio_waveforms/audio_waveforms.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/voice_provider.dart';
import '../theme/app_theme.dart';
import '../theme/app_typography.dart';
import '../theme/design_tokens.dart';
import '../utils/haptics.dart';

class VoiceRecorder extends ConsumerStatefulWidget {
  final ValueChanged<String>? onRecordingComplete;

  const VoiceRecorder({super.key, this.onRecordingComplete});

  @override
  ConsumerState<VoiceRecorder> createState() => _VoiceRecorderState();
}

class _VoiceRecorderState extends ConsumerState<VoiceRecorder>
    with SingleTickerProviderStateMixin {
  late final AnimationController _pulseController;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    );
  }

  @override
  void dispose() {
    _pulseController.dispose();
    super.dispose();
  }

  String _formatDuration(Duration d) {
    final mm = d.inMinutes.remainder(60).toString().padLeft(2, '0');
    final ss = d.inSeconds.remainder(60).toString().padLeft(2, '0');
    return '$mm:$ss';
  }

  @override
  Widget build(BuildContext context) {
    final voiceState = ref.watch(voiceProvider);
    final notifier = ref.read(voiceProvider.notifier);
    final colors = AppTheme.colors;
    final isRecording = voiceState.state == VoiceState.recording;
    final isDone = voiceState.state == VoiceState.done;

    if (isRecording && !_pulseController.isAnimating) {
      _pulseController.repeat(reverse: true);
    } else if (!isRecording && _pulseController.isAnimating) {
      _pulseController.stop();
      _pulseController.reset();
    }

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(Spacing.xl),
        child: Column(
          children: [
            if (!isDone) ...[
              // Recording waveform
              if (isRecording)
                SizedBox(
                  height: 48,
                  child: AudioWaveforms(
                    recorderController: notifier.service.recorderController,
                    size: Size(MediaQuery.of(context).size.width - 96, 48),
                    waveStyle: WaveStyle(
                      waveColor: colors.primary,
                      middleLineColor: Colors.transparent,
                      extendWaveform: true,
                      showMiddleLine: false,
                    ),
                  ),
                )
              else
                const SizedBox(height: 48),

              const SizedBox(height: Spacing.lg),

              // Timer
              Text(
                _formatDuration(voiceState.elapsed),
                style: AppTypography.headline.copyWith(
                  color: isRecording ? colors.error : colors.textSecondary,
                  fontFeatures: [const FontFeature.tabularFigures()],
                ),
              ),

              const SizedBox(height: Spacing.xl),

              // Main action button
              AnimatedBuilder(
                animation: _pulseController,
                builder: (context, child) {
                  final scale = isRecording
                      ? 1.0 + (_pulseController.value * 0.08)
                      : 1.0;
                  return Transform.scale(
                    scale: scale,
                    child: child,
                  );
                },
                child: Semantics(
                  button: true,
                  label: isRecording ? 'Stop recording' : 'Start recording',
                  child: GestureDetector(
                    onTap: () async {
                      Haptics.medium();
                      if (isRecording) {
                        await notifier.stopRecording();
                        // Read fresh state after async stop to avoid stale filePath
                        final updatedState = ref.read(voiceProvider);
                        if (updatedState.filePath != null) {
                          widget.onRecordingComplete?.call(updatedState.filePath!);
                        }
                      } else {
                        await notifier.startRecording();
                      }
                    },
                    child: Container(
                      width: 72,
                      height: 72,
                      decoration: BoxDecoration(
                        color: isRecording ? colors.error : colors.primary,
                        shape: BoxShape.circle,
                        boxShadow: [
                          BoxShadow(
                            color: (isRecording ? colors.error : colors.primary)
                                .withValues(alpha: 0.3),
                            blurRadius: 16,
                            spreadRadius: 2,
                          ),
                        ],
                      ),
                      child: Icon(
                        isRecording ? Icons.stop_rounded : Icons.mic_rounded,
                        color: Colors.white,
                        size: IconSizes.lg,
                      ),
                    ),
                  ),
                ),
              ),

              const SizedBox(height: Spacing.md),
              Text(
                isRecording ? 'Tap to stop' : 'Tap to record',
                style: AppTypography.bodySmall.copyWith(color: colors.textSecondary),
              ),
            ],

            // Playback view
            if (isDone) ...[
              AudioFileWaveforms(
                playerController: notifier.service.playerController,
                size: Size(MediaQuery.of(context).size.width - 96, 48),
                waveformType: WaveformType.fitWidth,
                playerWaveStyle: PlayerWaveStyle(
                  fixedWaveColor: colors.borderLight,
                  liveWaveColor: colors.primary,
                  seekLineColor: colors.primary,
                  spacing: 4,
                ),
              ),
              const SizedBox(height: Spacing.lg),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  IconButton(
                    onPressed: () => notifier.service.play(),
                    icon: Icon(Icons.play_arrow_rounded,
                        color: colors.primary, size: IconSizes.lg),
                    tooltip: 'Play recording',
                  ),
                  const SizedBox(width: Spacing.lg),
                  IconButton(
                    onPressed: () => notifier.service.pausePlayback(),
                    icon: Icon(Icons.pause_rounded,
                        color: colors.primary, size: IconSizes.lg),
                    tooltip: 'Pause playback',
                  ),
                  const SizedBox(width: Spacing.lg),
                  OutlinedButton.icon(
                    onPressed: () {
                      Haptics.light();
                      notifier.reset();
                    },
                    icon: const Icon(Icons.refresh_rounded),
                    label: const Text('Re-record'),
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }
}
