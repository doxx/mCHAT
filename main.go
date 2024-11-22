package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	multicastAddr = "224.0.0.251:5353"
	maxChunkSize  = 768 // Size per chunk before encryption
	maxInputSize  = 768 // Safe size for input before encryption
)

// Encryption helpers
func deriveKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

func encrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and combine nonce with ciphertext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func copyToClipboard(text string) error {
	cmd := exec.Command("pbcopy") // for macOS
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// Add helper function to split messages
func splitMessage(message string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(message); i += chunkSize {
		end := i + chunkSize
		if end > len(message) {
			end = len(message)
		}
		chunks = append(chunks, message[i:end])
	}
	return chunks
}

// Add URL detection regex
var urlRegex = regexp.MustCompile(`(https?:\/\/[^\s]+)`)

// Add function to open URLs
func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // Linux and others
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

// Add function to format message with clickable URLs
func formatMessageWithURLs(message string) string {
	return urlRegex.ReplaceAllStringFunc(message, func(url string) string {
		// Make URLs cyan and underlined for better visibility
		return fmt.Sprintf(`[#00FFFF::u]["%s"]%s[""][::-]`, url, url)
	})
}

func main() {
	username := flag.String("u", "", "Your username")
	debug := flag.Bool("debug", false, "Enable debug logging")
	password := flag.String("p", "", "Encryption password")
	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Println("Please provide both username (-u) and password (-p)")
		return
	}

	// Create the UI
	app := tview.NewApplication()

	// Create message view with proper type assertion
	messageView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true)

	// Add mouse handler for URL clicks
	messageView.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick {
			// Get text at click position
			row, col := event.Position()
			text := messageView.GetText(true)
			lines := strings.Split(text, "\n")

			// Make sure we're within bounds
			if row < len(lines) {
				line := lines[row]
				// Find URLs in this line
				matches := urlRegex.FindAllStringIndex(line, -1)
				for _, match := range matches {
					// Check if click was within URL bounds
					if col >= match[0] && col <= match[1] {
						url := line[match[0]:match[1]]
						if err := openURL(url); err != nil {
							currentText := messageView.GetText(true)
							messageView.SetText(currentText +
								fmt.Sprintf("[red]Error opening URL: %v\n", err))
						} else {
							currentText := messageView.GetText(true)
							messageView.SetText(currentText +
								fmt.Sprintf("[green]Opening URL: %s\n", url))
						}
						break
					}
				}
			}
		}
		return action, event
	})

	// Set changed function to ensure proper drawing
	messageView.SetChangedFunc(func() {
		app.Draw()
	})

	messageView.SetBorder(true).SetTitle(" Messages (⌘S to save, Tab to copy, Click URLs to open) ")

	// Create input field (bottom)
	inputField := tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0)

	// Create layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(messageView, 0, 1, false).
		AddItem(inputField, 1, 0, true)

	// Derive encryption key from password
	key := deriveKey(*password)

	// Add key handler for saving chat log and copying
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 's' && event.Modifiers() == tcell.ModMeta {
			filename := fmt.Sprintf("chat_log_%s.txt", time.Now().Format("2006-01-02_15-04-05"))
			f, err := os.Create(filename)
			if err != nil {
				messageView.SetText(messageView.GetText(true) +
					fmt.Sprintf("[red]Error saving chat log: %v\n", err))
				return nil
			}
			defer f.Close()

			f.WriteString(messageView.GetText(true))
			messageView.SetText(messageView.GetText(true) +
				fmt.Sprintf("[green]Chat log saved to %s\n", filename))
			return nil
		} else if event.Key() == tcell.KeyTab {
			text := messageView.GetText(true)
			lines := strings.Split(text, "\n")

			start := 0
			if len(lines) > 10 {
				start = len(lines) - 10
			}

			textToCopy := strings.Join(lines[start:], "\n")
			if err := copyToClipboard(textToCopy); err != nil {
				messageView.SetText(messageView.GetText(true) +
					fmt.Sprintf("[red]Error copying to clipboard: %v\n", err))
			} else {
				messageView.SetText(messageView.GetText(true) +
					"[green]Last messages copied to clipboard\n")
			}
			return nil
		}
		return event
	})

	// Set up UDP connection for broadcasting
	addr, err := net.ResolveUDPAddr("udp", multicastAddr)
	if err != nil {
		fmt.Printf("Error resolving address: %v\n", err)
		return
	}

	// Create connections
	sendConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("Error creating send connection: %v\n", err)
		return
	}
	defer sendConn.Close()

	receiveConn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("Error creating receive connection: %v\n", err)
		return
	}
	defer receiveConn.Close()

	// Start listener
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, _, err := receiveConn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Printf("Error reading: %v\n", err)
				continue
			}

			message := string(buffer[:n])
			if strings.HasPrefix(message, "CHAT:") {
				encrypted := message[5:]
				decrypted, err := decrypt(encrypted, key)
				if err != nil {
					if *debug {
						app.QueueUpdateDraw(func() {
							currentText := messageView.GetText(true)
							messageView.SetText(currentText +
								fmt.Sprintf("[red]Failed to decrypt message: %v\n", err))
						})
					}
					continue
				}

				// Fix the parsing here
				parts := strings.SplitN(decrypted, ":", 3)
				if len(parts) == 3 && parts[0] != *username {
					senderName := parts[0]
					timestamp := parts[1]
					messageText := parts[2]

					// Ensure timestamp is properly formatted
					if !strings.Contains(timestamp, ":") {
						// Try to parse and reformat the timestamp from the message
						timeParts := strings.Split(messageText, ":")
						if len(timeParts) >= 3 {
							timestamp = fmt.Sprintf("%s:%s:%s",
								timestamp,    // hour
								timeParts[0], // minute
								timeParts[1]) // second
							messageText = strings.Join(timeParts[2:], ":")
						}
					}

					formattedMessage := formatMessageWithURLs(messageText)

					app.QueueUpdateDraw(func() {
						currentText := messageView.GetText(true)
						newMessage := fmt.Sprintf("[yellow][%s][white] %s: %s\n",
							timestamp, senderName, formattedMessage)
						messageView.SetText(currentText + newMessage)
					})
				}
			}
		}
	}()

	// Update input handler to ensure consistent format
	inputField.SetDoneFunc(func(keyPress tcell.Key) {
		if keyPress != tcell.KeyEnter {
			return
		}

		message := inputField.GetText()
		if message == "" {
			return
		}

		timestamp := time.Now().Format("15:04:05")
		formattedMessage := formatMessageWithURLs(message)

		plaintext := fmt.Sprintf("%s:%s:%s", *username, timestamp, message)
		encrypted, err := encrypt(plaintext, key)
		if err != nil {
			messageView.SetText(messageView.GetText(true) +
				fmt.Sprintf("[red]Error encrypting message: %v\n", err))
			return
		}

		broadcastMsg := fmt.Sprintf("CHAT:%s", encrypted)

		if *debug {
			messageView.SetText(messageView.GetText(true) +
				"[gray]Broadcasting encrypted message\n")
		}

		_, err = sendConn.Write([]byte(broadcastMsg))
		if err != nil {
			messageView.SetText(messageView.GetText(true) +
				fmt.Sprintf("[red]Error broadcasting: %v\n", err))
		}

		// Add the message to the view
		currentText := messageView.GetText(true)
		newMessage := fmt.Sprintf("[yellow][%s][white] %s: %s\n",
			timestamp, *username, formattedMessage)
		messageView.SetText(currentText + newMessage)

		inputField.SetText("")
	})

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
	}
}
