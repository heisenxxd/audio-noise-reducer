package grpc

import (
	"context"
	"fmt"
	"log"

	pb "audio/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.AudioProcessorClient
	stream pb.AudioProcessor_ProcessAudioStreamClient
}

func NewClient(ctx context.Context, serverAddress string) (*Client, error) {
	log.Printf("Conectando ao servidor gRPC em %s...", serverAddress)

	conn, err := grpc.NewClient(
		serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente: %w", err)
	}

	client := pb.NewAudioProcessorClient(conn)

	log.Println("Conectado ao servidor gRPC")

	stream, err := client.ProcessAudioStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir stream: %w", err)
	}

	return &Client{
		conn:   conn,
		client: client,
		stream: stream,
	}, nil
}

func (c *Client) ProcessAudio(audioData []byte, sampleRate, channels int32) ([]byte, error) {
	request := &pb.AudioChunk{
		AudioData:  audioData,
		SampleRate: sampleRate,
		Channels:   channels,
	}
	if err := c.stream.Send(request); err != nil {
		return nil, err
	}

	response, err := c.stream.Recv()
	if err != nil {
		return nil, err
	}

	return response.AudioData, nil
}

func (c *Client) Close() error {
	if c.stream != nil {
		c.stream.CloseSend()
	}
	if c.conn != nil {
		log.Println("Fechando conexão gRPC...")
		return c.conn.Close()
	}
	return nil
}
