package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/go-vgo/robotgo"
	"github.com/gorilla/websocket"
)

var (
	addr      = flag.String("addr", ":8080", "http service address")
	tokenFlag = flag.String("token", "", "WebSocket auth token")
	scaleFlag = flag.Float64("scale", 5.0, "Scroll / zoom scale factor")

	threeSwipeDX   float64
	threeSwipeDY   float64
	touchpadWidth  = 500.0
	touchpadHeight = 800.0
	screenWidth    = 0
	screenHeight   = 0
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Type           string  `json:"type"`
	Dx             float64 `json:"dx,omitempty"`
	Dy             float64 `json:"dy,omitempty"`
	Button         string  `json:"button,omitempty"`
	Delta          float64 `json:"delta,omitempty"` // for zoom
	TouchpadWidth  float64 `json:"touchpadWidth,omitempty"`
	TouchpadHeight float64 `json:"touchpadHeight,omitempty"`
	ScreenWidth    float64 `json:"screenWidth,omitempty"`
	ScreenHeight   float64 `json:"screenHeight,omitempty"`
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("X-Auth-Token")
	}
	if token != *tokenFlag {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer conn.Close()
	log.Println("Client connected")

	// Get Mac screen size
	screenWMac, screenHMac := robotgo.GetScreenSize()

	moveScaleX := 1.0
	moveScaleY := 1.0
	scrollScale := *scaleFlag
	zoomScale := *scaleFlag / 2.0

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		var m Message
		if err := json.Unmarshal(msgBytes, &m); err != nil {
			log.Println("invalid message:", string(msgBytes))
			continue
		}

		log.Printf("Received message (type %s): %+v\n", m.Type, m)

		switch strings.ToLower(m.Type) {

		case "deviceinfo":
			// Capture mobile touchpad and screen size for scaling
			if m.TouchpadWidth > 0 && m.TouchpadHeight > 0 && m.ScreenWidth > 0 && m.ScreenHeight > 0 {
				touchpadWidth = m.TouchpadWidth
				touchpadHeight = m.TouchpadHeight
				screenWidth = int(m.ScreenWidth)
				screenHeight = int(m.ScreenHeight)

				moveScaleX = float64(screenWMac) / touchpadWidth
				moveScaleY = float64(screenHMac) / touchpadHeight
				log.Printf("Device info received. Move scale: X=%.2f Y=%.2f", moveScaleX, moveScaleY)
			}

		case "move":
			x, y := robotgo.GetMousePos()
			dx := int(m.Dx * moveScaleX)
			dy := int(m.Dy * moveScaleY)
			robotgo.MoveMouseSmooth(x+dx, y+dy, 0.5, 0.5)

		case "click":
			btn := "left"
			if m.Button != "" {
				btn = strings.ToLower(m.Button)
			}
			robotgo.Click(btn)

		case "scroll":
			h := int(m.Dx * scrollScale)
			v := int(-m.Dy * scrollScale) // invert vertical for natural feel
			if h != 0 || v != 0 {
				robotgo.Scroll(h, v)
			}

		case "zoom":
			if m.Delta != 0 {
				delta := int(-m.Delta * zoomScale)
				robotgo.KeyToggle("ctrl", "down")
				robotgo.Scroll(0, delta)
				robotgo.KeyToggle("ctrl", "up")
			}

		case "threeswipe":
			threeSwipeDX += m.Dx
			threeSwipeDY += m.Dy
			threshold := 5.0 // pixels

			// Horizontal swipe
			if threeSwipeDX > threshold {
				log.Println("Three-finger swipe right → next desktop")
				robotgo.KeyTap("right", []string{"ctrl"})
				threeSwipeDX, threeSwipeDY = 0, 0
			} else if threeSwipeDX < -threshold {
				log.Println("Three-finger swipe left → previous desktop")
				robotgo.KeyTap("left", []string{"ctrl"})
				threeSwipeDX, threeSwipeDY = 0, 0
			}

			// Vertical swipe
			if threeSwipeDY > threshold {
				log.Println("Three-finger swipe down → App Exposé")
				robotgo.KeyTap("down", []string{"ctrl"})
				threeSwipeDX, threeSwipeDY = 0, 0
			} else if threeSwipeDY < -threshold {
				log.Println("Three-finger swipe up → Mission Control")
				robotgo.KeyTap("up", []string{"ctrl"})
				threeSwipeDX, threeSwipeDY = 0, 0
			}

		default:
			log.Println("Unknown message type:", m.Type)
		}
	}
}

func main() {
	flag.Parse()
	if *tokenFlag == "" {
		log.Fatal("WebSocket auth token must be provided using -token flag")
	}

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", wsHandler)

	log.Println("Touchpad server starting on", *addr)
	log.Println("WebSocket token:", *tokenFlag)
	log.Println("Scroll/zoom scale:", *scaleFlag)
	log.Println("Make sure this binary has Accessibility access in System Settings → Privacy → Accessibility")
	log.Fatal(http.ListenAndServe(*addr, nil))
}
