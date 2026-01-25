package ui

import (
	"audio/audio"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gen2brain/malgo"
)

func Menu() (*audio.DeviceInfo, *audio.DeviceInfo, *audio.Manager, error) {
    reader := bufio.NewReader(os.Stdin)
    
    for {
        printMainMenu()
        escolha := readInput(reader)
        
        switch escolha {
        case "1":
            manager, err := audio.NewManager()
            if err != nil {
                return nil, nil, nil, fmt.Errorf("erro ao criar manager: %w", err)
            }
            
            fmt.Println("\n=== DISPOSITIVO DE ENTRADA (Microfone) ===")
            inputDevice, err := DeviceListMenu(reader, manager, malgo.Capture)
            if err != nil {
                manager.Close()
                fmt.Printf("Erro: %v\n", err)
                continue
            }
            
            fmt.Println("\n=== DISPOSITIVO DE SAÍDA (Output Virtual) ===")
            outputDevice, err := DeviceListMenu(reader, manager, malgo.Playback)
            if err != nil {
                manager.Close()
                fmt.Printf("Erro: %v\n", err)
                continue
            }
            
            fmt.Printf("\n- Entrada: %s\n", inputDevice.Name())
            fmt.Printf("- Saída: %s\n", outputDevice.Name())
            
            return inputDevice, outputDevice, manager, nil
            
        case "2":
            fmt.Println("Saindo...")
            return nil, nil, nil, nil
            
        default:
            fmt.Println("Opção inválida! Tente novamente.")
        }
    }
}

func DeviceListMenu(reader *bufio.Reader, manager *audio.Manager, deviceType malgo.DeviceType) (*audio.DeviceInfo, error) {
	var devices []audio.DeviceInfo
	var err error
	var tipoNome string

	switch deviceType {
	case malgo.Capture:
		devices, err = manager.GetCaptureDevices()
		tipoNome = "Entrada"
	case malgo.Playback:
		devices, err = manager.GetPlaybackDevices()
		tipoNome = "Saída"
	default:
		return nil, fmt.Errorf("tipo de dispositivo inválido")
	}

	if err != nil {
		return nil, fmt.Errorf("erro ao listar dispositivos: %w", err)
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("nenhum dispositivo de %s encontrado", tipoNome)
	}

	printDeviceList(devices, tipoNome)

	for {
		fmt.Print("\n> ")
		opcao := readInput(reader)

		numeroEscolha, err := strconv.Atoi(opcao)
		if err != nil {
			fmt.Println("Digite um número válido")
			continue
		}

		if numeroEscolha < 1 || numeroEscolha > len(devices) {
			fmt.Printf("Escolha entre 1 e %d\n", len(devices))
			continue
		}

		return &devices[numeroEscolha-1], nil
	}
}

func printMainMenu() {
	fmt.Println("\n========================================")
	fmt.Println("  	Sistema de Limpeza de Áudio")
	fmt.Println("========================================")
	fmt.Println("[1] - Iniciar")
	fmt.Println("[2] - Sair")
	fmt.Print("\n> ")
}

func printDeviceList(devices []malgo.DeviceInfo, tipo string) {
	fmt.Println("\n========================================")
	fmt.Printf("  	Selecione o Dispositivo de %s\n", tipo)
	fmt.Println("========================================")
	
	for i, device := range devices {
		fmt.Printf("[%d] - %s", i+1, device.Name())
		if device.IsDefault == 1 {
			fmt.Print("- (Padrão)")
		}
		fmt.Println()
	}
}

func readInput(reader *bufio.Reader) string {
	s, _ := reader.ReadString('\n')
	return strings.TrimSpace(s)
}