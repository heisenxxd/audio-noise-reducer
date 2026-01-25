package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "audio/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.AudioProcessorClient
}

func NewClient(serverAddress string) (*Client, error) {
	log.Printf("Conectando ao servidor gRPC em %s...", serverAddress)

	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := pb.NewAudioProcessorClient(conn)

	_, err = client.ProcessAudio(ctx, &pb.AudioChunk{
		AudioData:  []byte{},
		SampleRate: 48000,
		Channels:   1,
		Format:     16,
	})

	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao conectar ao servidor: %w", err)
	}

	log.Println("Conectado ao servidor gRPC")

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c *Client) ProcessAudio(audioData []byte, sampleRate, channels int32) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	request := &pb.AudioChunk{
		AudioData:  audioData,
		SampleRate: sampleRate,
		Channels:   channels,
		Format:     16, 
		Timestamp:  time.Now().UnixNano(),
	}

	response, err := c.client.ProcessAudio(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar áudio: %w", err)
	}

	return response.AudioData, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		log.Println("Fechando conexão gRPC...")
		return c.conn.Close()
	}
	return nil
}
