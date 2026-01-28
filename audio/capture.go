package audio

import (
	"context"
	"fmt"
	"log"

	"github.com/gen2brain/malgo"
)

type Processor interface {
	ProcessAudio(audioData []byte, sampleRate, channels int32) ([]byte, error)
}

type AudioStream struct {
	manager        *Manager
	inputDevice    *malgo.Device
	outputDevice   *malgo.Device
	processor      Processor
	ctx            context.Context
	cancel         context.CancelFunc
	sampleRate     uint32
	channels       uint32
	playbackBuffer *RingBuffer
	processingChan chan []byte
}

func NewAudioStream(ctx context.Context, cancel context.CancelFunc, manager *Manager, inputDeviceInfo, outputDeviceInfo *DeviceInfo, processor Processor) (*AudioStream, error) {

	bufferSize := int(2 * 48000 * 1 * 0.2)

	stream := &AudioStream{
		manager:        manager,
		processor:      processor,
		ctx:            ctx,
		cancel:         cancel,
		sampleRate:     48000,
		channels:       1,
		playbackBuffer: NewRingBuffer(bufferSize),
		processingChan: make(chan []byte, 100),
	}

	if err := stream.initCaptureDevice(inputDeviceInfo); err != nil {
		cancel()
		return nil, fmt.Errorf("erro ao inicializar captura: %w", err)
	}

	if err := stream.initPlaybackDevice(outputDeviceInfo); err != nil {
		stream.inputDevice.Uninit()
		cancel()
		return nil, fmt.Errorf("erro ao inicializar reprodução: %w", err)
	}

	return stream, nil
}

func (s *AudioStream) processWorker() {
	for audioData := range s.processingChan {
		processed, err := s.processor.ProcessAudio(audioData, int32(s.sampleRate), int32(s.channels))
		if err == nil {
			s.playbackBuffer.Write(processed)
		}
	}
}

func (s *AudioStream) initCaptureDevice(deviceInfo *DeviceInfo) error {
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = s.channels
	deviceConfig.Capture.DeviceID = deviceInfo.ID.Pointer()
	deviceConfig.SampleRate = s.sampleRate

	callbacks := malgo.DeviceCallbacks{
		Data: s.onCaptureData,
	}

	device, err := malgo.InitDevice(s.manager.Context(), deviceConfig, callbacks)
	if err != nil {
		return err
	}

	s.inputDevice = device
	return nil
}

func (s *AudioStream) initPlaybackDevice(deviceInfo *DeviceInfo) error {
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = s.channels
	deviceConfig.Playback.DeviceID = deviceInfo.ID.Pointer()
	deviceConfig.SampleRate = s.sampleRate

	callbacks := malgo.DeviceCallbacks{
		Data: s.onPlaybackData,
	}

	device, err := malgo.InitDevice(s.manager.Context(), deviceConfig, callbacks)
	if err != nil {
		return err
	}

	s.outputDevice = device
	return nil
}

func (s *AudioStream) onCaptureData(pOutputSample, pInputSample []byte, frameCount uint32) {
	if len(pInputSample) == 0 {
		return
	}

	select {
	case s.processingChan <- append([]byte(nil), pInputSample...):
	default:
		log.Println("Aviso: Fila de processamento cheia, descartando frame")
	}
}

func (s *AudioStream) onPlaybackData(pOutputSample, pInputSample []byte, frameCount uint32) {
	if len(pOutputSample) == 0 {
		return
	}

	bytesRead := s.playbackBuffer.Read(pOutputSample)

	if bytesRead < len(pOutputSample) {
		for i := bytesRead; i < len(pOutputSample); i++ {
			pOutputSample[i] = 0
		}
	}
}

func (s *AudioStream) Start() error {
	go s.processWorker()

	if err := s.outputDevice.Start(); err != nil {
		return fmt.Errorf("erro ao iniciar reprodução: %w", err)
	}

	if err := s.inputDevice.Start(); err != nil {
		s.outputDevice.Stop()
		return fmt.Errorf("erro ao iniciar captura: %w", err)
	}

	log.Println("Stream de áudio iniciado")
	return nil
}

func (s *AudioStream) Stop() error {
	s.cancel()

	var errs []error

	if s.inputDevice != nil {
		if err := s.inputDevice.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("erro ao parar captura: %w", err))
		}
	}

	if s.outputDevice != nil {
		if err := s.outputDevice.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("erro ao parar reprodução: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("erros ao parar stream: %v", errs)
	}

	log.Println("✓ Stream de áudio parado")
	return nil
}

func (s *AudioStream) Close() error {
	s.Stop()

	if s.inputDevice != nil {
		s.inputDevice.Uninit()
	}

	if s.outputDevice != nil {
		s.outputDevice.Uninit()
	}

	close(s.processingChan)

	return nil
}

func (s *AudioStream) GetBufferStatus() (used, available int) {
	return s.playbackBuffer.Used(), s.playbackBuffer.Available()
}
