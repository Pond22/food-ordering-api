package main

import (
	"bytes"
	"fmt"
	"food-ordering-api/api_handlers"
	"food-ordering-api/db"
	_ "food-ordering-api/docs"
	"food-ordering-api/routes"
	service "food-ordering-api/services"
	"image"
	"image/draw"
	"image/png"
	"log"
	"net"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

// @title Food Ordering API
// @version 1.0
// @description This is a simple API for ordering food.
// @termsOfService https://example.com/terms

// @contact.name API Support
// @contact.url https://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @tags categories - การจัดการหมวดหมู่อาหาร
// @tags menu - การจัดการเมนูอาหาร
// @tags orders - การจัดการคำสั่งซื้อ
// @tags tables - การจัดกโต๊ะ
func main() {
	db.InitDatabase()

	printerIP := "192.168.1.100" // เปลี่ยนเป็น IP จริงของเครื่องพิมพ์
	port := 9100

	err := testDirectImagePrint(printerIP, port)
	if err != nil {
		fmt.Printf("Test failed: %v\n", err)
		return
	}
	fmt.Println("Test completed successfully")

	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,

		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		ExposeHeaders:    "Content-Length, Access-Control-Allow-Origin",
		AllowCredentials: false,
		MaxAge:           300,
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Get("/ws/tables", api_handlers.TableWebSocketHandler(db.DB))

	// app.Use("/ws/printer", api_handlers.HandlePrinterWebSocket)
	app.Use("/ws/printer", service.HandlePrinterWebSocket)

	routes.SetupRoutes(app)

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}

// type Command struct {
// 	// สำหรับสะสมคำสั่ง ESC/POS
// 	data []byte
// }

// func NewCommand() *Command {
// 	return &Command{
// 		data: make([]byte, 0),
// 	}
// }

// func (c *Command) AddBytes(b []byte) {
// 	c.data = append(c.data, b...)
// }

// func (c *Command) AddString(s string) {
// 	c.data = append(c.data, []byte(s)...)
// }

// func testCharacterControl(printerIP string, port int) error {
// 	// สร้างชุดคำสั่งทั้งหมดที่เกี่ยวกับตัวอักษร
// 	controls := []struct {
// 		name string
// 		cmds [][]byte
// 	}{
// 		// FS ! n Select print mode(s) for Kanji characters
// 		{"Kanji Mode 1", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1C, 0x21, 0x00}, // FS ! n
// 			{0x1C, 0x26},       // FS & Select Kanji character mode
// 		}},

// 		// Select character code table
// 		{"Code Table Test 1", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1B, 0x74, 0x01}, // ESC t n
// 			{0x1C, 0x43, 0x01}, // FS C n
// 		}},

// 		{"Code Table Test 2", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1B, 0x74, 0x02}, // ESC t n
// 			{0x1C, 0x43, 0x02}, // FS C n
// 		}},

// 		// Try different character sets
// 		{"Character Set 1", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1B, 0x52, 0x0D}, // ESC R n
// 			{0x1B, 0x74, 0x11}, // ESC t n
// 		}},

// 		{"Character Set 2", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1B, 0x52, 0x0E}, // ESC R n
// 			{0x1B, 0x74, 0x1C}, // ESC t n
// 		}},

// 		// Try different combinations
// 		{"Combo 1", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1B, 0x4D, 0x01}, // Select character font
// 			{0x1B, 0x74, 0x11}, // Select character code table
// 			{0x1B, 0x52, 0x0D}, // Select international character set
// 		}},

// 		{"Combo 2", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1B, 0x4D, 0x00}, // Select character font
// 			{0x1B, 0x74, 0x1C}, // Select character code table
// 			{0x1B, 0x52, 0x0E}, // Select international character set
// 		}},

// 		// Try extended commands
// 		{"Extended 1", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1C, 0x21, 0x00}, // Select character size for Kanji
// 			{0x1C, 0x2E},       // Cancel Kanji character mode
// 			{0x1B, 0x74, 0x11}, // Select character code table
// 		}},

// 		{"Extended 2", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1C, 0x21, 0x00}, // Select character size for Kanji
// 			{0x1C, 0x26},       // Select Kanji character mode
// 			{0x1B, 0x74, 0x1C}, // Select character code table
// 		}},

// 		// Try user-defined characters
// 		{"User Defined", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1B, 0x25, 0x01}, // Select user-defined character set
// 			{0x1B, 0x74, 0x11}, // Select character code table
// 		}},

// 		// Try different text modes
// 		{"Text Mode 1", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1B, 0x21, 0x00}, // Select print mode(s)
// 			{0x1B, 0x74, 0x11}, // Select character code table
// 			{0x1B, 0x52, 0x0D}, // Select international character set
// 		}},

// 		// Additional combinations
// 		{"Additional 1", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1C, 0x26},       // Select Kanji mode
// 			{0x1B, 0x74, 0x11}, // TIS-620
// 			{0x1B, 0x52, 0x0E}, // Thai character set
// 		}},

// 		{"Additional 2", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1C, 0x26},       // Select Kanji mode
// 			{0x1B, 0x74, 0x1C}, // CP-874
// 			{0x1B, 0x52, 0x0D}, // Thai character set
// 		}},

// 		// Try every possible code page one by one
// 		{"Code Page Loop", [][]byte{
// 			{0x1B, 0x40},       // Initialize
// 			{0x1B, 0x74, 0xFF}, // Will be replaced in loop
// 		}},
// 	}

// 	// ทดสอบแต่ละชุดคำสั่ง
// 	for _, control := range controls {
// 		cmd := NewCommand()

// 		// Special case for code page loop
// 		if control.name == "Code Page Loop" {
// 			for i := 0; i < 256; i++ {
// 				cmd.AddBytes([]byte{0x1B, 0x40})          // Initialize
// 				cmd.AddBytes([]byte{0x1B, 0x74, byte(i)}) // Try each code page
// 				cmd.AddString(fmt.Sprintf("\n=== Testing Code Page %d ===\n", i))
// 				cmd.AddString("ทดสอบภาษาไทย\n")
// 				cmd.AddString("สวัสดีครับ/ค่ะ\n")
// 				cmd.AddString("กขคงจฉชซ\n")
// 				cmd.AddString("---------------\n")
// 			}
// 		} else {
// 			cmd.AddString(fmt.Sprintf("\n=== %s ===\n", control.name))
// 			for _, command := range control.cmds {
// 				cmd.AddBytes(command)
// 			}
// 			cmd.AddString("ทดสอบภาษาไทย\n")
// 			cmd.AddString("สวัสดีครับ/ค่ะ\n")
// 			cmd.AddString("กขคงจฉชซ\n")
// 			cmd.AddString("---------------\n")
// 		}

// 		// Cut paper between tests
// 		cmd.AddBytes([]byte{0x1D, 0x56, 0x41, 0x03})

// 		// Send to printer
// 		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", printerIP, port))
// 		if err != nil {
// 			return fmt.Errorf("connection failed for %s: %v", control.name, err)
// 		}

// 		_, err = conn.Write(cmd.data)
// 		if err != nil {
// 			return fmt.Errorf("write failed for %s: %v", control.name, err)
// 		}

// 		conn.Close()
// 		time.Sleep(2 * time.Second) // รอระหว่างแต่ละชุดทดสอบ
// 	}

// 	return nil
// }

func testDirectImagePrint(printerIP string, port int) error {
	// สร้างภาพทดสอบ
	width := 384 // เพิ่มความกว้าง
	height := 400
	dpi := float64(203)             // ตั้งค่า DPI ตามเครื่องพิมพ์
	fontSize := 16.0                // เพิ่มขนาดฟอนต์
	lineSpacing := (fontSize * 1.5) // ระยะห่างระหว่างบรรทัด
	maxLineWidth := 32              // จำนวนตัวอักษรต่อบรรทัด

	// สร้างภาพใหม่พื้นขาว
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	// โหลด font
	fontBytes, err := os.ReadFile("THSarabunNew.ttf")
	if err != nil {
		return fmt.Errorf("error loading font: %v", err)
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("error parsing font: %v", err)
	}

	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(f)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.Black)
	c.SetHinting(font.HintingFull)

	// ข้อความทดสอบ
	longText := "ทดสอบภาษาไทยโดยหลับปุ๋ย ทดสอบการตัดข้อความและความถูกต้องในการพิมพ์เมื่อข้อความยาว ทดสอบ123ทดสอบ123ทดสอบ123ทดสอบ123"

	// แบ่งข้อความเป็นบรรทัด
	var lines []string
	words := strings.Fields(longText)
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) > maxLineWidth {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// คำเดียวยาวเกิน แบ่งกลางคำ
				lines = append(lines, word[:maxLineWidth])
				currentLine = word[maxLineWidth:]
			}
		} else {
			currentLine = testLine
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// วาดแต่ละบรรทัด
	y := int(fontSize * 1.5) // เริ่มต้นจากด้านบน

	for _, line := range lines {
		pt := freetype.Pt(20, y)
		_, err = c.DrawString(line, pt)
		if err != nil {
			return fmt.Errorf("error drawing text: %v", err)
		}
		y += int(fontSize * lineSpacing)
	}

	// วาดข้อความทดสอบอื่นๆ
	testText := []string{
		"สวัสดีครับ/ค่ะ",
		"ทดสอบ 1234567890",
		"กขคงจฉชซ",
		"ฟหกด่าสวง",
	}

	for _, text := range testText {
		pt := freetype.Pt(20, y)
		_, err = c.DrawString(text, pt)
		if err != nil {
			return fmt.Errorf("error drawing text: %v", err)
		}
		y += int(fontSize * lineSpacing)
	}

	// แปลงเป็น ESC/POS command
	var buf bytes.Buffer
	buf.Write([]byte{0x1B, 0x40})             // Initialize printer
	buf.Write([]byte{0x1D, 0x76, 0x30, 0x00}) // Set bitmap mode

	widthBytes := (width + 7) / 8
	buf.WriteByte(byte(widthBytes & 0xFF))
	buf.WriteByte(byte(widthBytes >> 8))
	buf.WriteByte(byte(height & 0xFF))
	buf.WriteByte(byte(height >> 8))

	// แปลงภาพเป็น bitmap โดยปรับ threshold
	for y := 0; y < height; y++ {
		for x := 0; x < width; x += 8 {
			var b byte
			for bit := 0; bit < 8; bit++ {
				if x+bit < width {
					r, g, b_, _ := img.At(x+bit, y).RGBA()
					// ปรับ threshold ให้ตัวอักษรชัดขึ้น
					if (r+g+b_)/3 < 0x7FFF {
						b |= 1 << (7 - bit)
					}
				}
			}
			buf.WriteByte(b)
		}
	}

	buf.Write([]byte{0x1D, 0x56, 0x41, 0x03}) // Cut paper

	// ส่งไปยังเครื่องพิมพ์
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", printerIP, port))
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("print failed: %v", err)
	}

	return nil
}

func saveImageForDebug(img image.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}