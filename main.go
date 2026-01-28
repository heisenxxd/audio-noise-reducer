package main

import (
	"audio/audio"
	"audio/grpc"
	"audio/ui"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	CaptureDevice, PlaybackDevice, manager, err := ui.Menu()
	if err != nil {
		log.Fatalf("Erro: %s", err)
	}
	if manager == nil {
		fmt.Println("Até")
		return
	}
	defer manager.Close()

	fmt.Println("Conectando ao servidor GRPC")

	grpcClient, err := grpc.NewClient("localhost:50051")
	if err != nil {
		log.Fatalf("Erro ao conectar com o servidor GRPC, erro: %s", err)
	}

	defer grpcClient.Close()

	fmt.Println("Conexão estabelecida com sucesso.")

	fmt.Println("Iniciando Stream de áudio")
	stream, err := audio.NewAudioStream(manager, CaptureDevice, PlaybackDevice, grpcClient)
	if err != nil {
		log.Fatalf("Erro ao criar stream de áudio, erro: %s", err)
	}

	defer stream.Close()

	err = stream.Start()
	if err != nil {
		log.Fatalf("Erro ao iniciar captura de áudio")
	}

	fmt.Println("\nProcessando áudio em tempo real")
	fmt.Printf("Capturando de: %s\n", CaptureDevice.Name())
	fmt.Printf("Reproduzindo áudio em: %s\n", PlaybackDevice)
	fmt.Println("Pressione Ctrl+C para parar")

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)
	<-signChan

	fmt.Println("Parando processamento")
	stream.Stop()
	fmt.Println("Finalizado com sucesso")
}
