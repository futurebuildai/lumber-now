import 'dart:async';
import 'package:audio_waveforms/audio_waveforms.dart';
import 'package:path_provider/path_provider.dart';

class VoiceService {
  RecorderController? _recorderController;
  PlayerController? _playerController;
  String? _recordedPath;

  RecorderController get recorderController {
    _recorderController ??= RecorderController()
      ..androidEncoder = AndroidEncoder.aac
      ..androidOutputFormat = AndroidOutputFormat.mpeg4
      ..iosEncoder = IosEncoder.kAudioFormatMPEG4AAC
      ..sampleRate = 44100;
    return _recorderController!;
  }

  PlayerController get playerController {
    _playerController ??= PlayerController();
    return _playerController!;
  }

  String? get recordedPath => _recordedPath;

  Future<void> startRecording() async {
    final dir = await getTemporaryDirectory();
    final path = '${dir.path}/voice_${DateTime.now().millisecondsSinceEpoch}.m4a';
    await recorderController.record(path: path);
    _recordedPath = path;
  }

  Future<String?> stopRecording() async {
    final path = await recorderController.stop();
    _recordedPath = path;
    return path;
  }

  Future<void> pauseRecording() async {
    await recorderController.pause();
  }

  Future<void> preparePlayer() async {
    if (_recordedPath != null) {
      await playerController.preparePlayer(
        path: _recordedPath!,
        shouldExtractWaveform: true,
      );
    }
  }

  Future<void> play() async {
    await playerController.startPlayer();
  }

  Future<void> pausePlayback() async {
    await playerController.pausePlayer();
  }

  Future<void> stopPlayback() async {
    await playerController.stopPlayer();
  }

  void dispose() {
    _recorderController?.dispose();
    _playerController?.dispose();
  }

  void reset() {
    _recorderController?.dispose();
    _recorderController = null;
    _playerController?.dispose();
    _playerController = null;
    _recordedPath = null;
  }
}
