package gcloud

import (
	"context"
	"fmt"
	"io"
	"strings"

	speech "cloud.google.com/go/speech/apiv2"
	speechpb "cloud.google.com/go/speech/apiv2/speechpb"
	"google.golang.org/api/option"
)

type SpeechClient struct {
	client *speech.Client
}

func NewSpeechClient(ctx context.Context, credentialsFile string) (*SpeechClient, error) {
	client, err := speech.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("create speech client: %w", err)
	}
	return &SpeechClient{client: client}, nil
}

func (s *SpeechClient) Transcribe(ctx context.Context, audioData io.Reader) (string, error) {
	data, err := io.ReadAll(audioData)
	if err != nil {
		return "", fmt.Errorf("read audio data: %w", err)
	}

	req := &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			DecodingConfig: &speechpb.RecognitionConfig_AutoDecodingConfig{
				AutoDecodingConfig: &speechpb.AutoDetectDecodingConfig{},
			},
			LanguageCodes: []string{"en-US"},
			Model:         "long",
		},
		AudioSource: &speechpb.RecognizeRequest_Content{
			Content: data,
		},
	}

	resp, err := s.client.Recognize(ctx, req)
	if err != nil {
		return "", fmt.Errorf("recognize speech: %w", err)
	}

	var parts []string
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			parts = append(parts, alt.Transcript)
		}
	}

	transcript := strings.Join(parts, " ")
	if transcript == "" {
		return "", fmt.Errorf("no transcription results returned")
	}

	return transcript, nil
}

func (s *SpeechClient) Close() error {
	return s.client.Close()
}
