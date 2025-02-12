package main

import (
	"food-ordering-api/api_handlers"
	"food-ordering-api/db"
	_ "food-ordering-api/docs"
	"food-ordering-api/routes"
	service "food-ordering-api/services"
	utils "food-ordering-api/utility"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
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
// @description Enter the token with the `Bearer: ` prefix, e.g. "Bearer abcde12345".

// @tags categories - การจัดการหมวดหมู่อาหาร
// @tags menu - การจัดการเมนูอาหาร
// @tags orders - การจัดการคำสั่งซื้อ
// @tags tables - การจัดกโต๊ะ
func main() {
	db.InitDatabase()
	utils.InitAPIKeys()

	// printerIP := "192.168.1.100" // เปลี่ยนเป็น IP จริงของเครื่องพิมพ์
	// port := 9100

	// err := testDirectImagePrint(printerIP, port)
	// if err != nil {
	// 	fmt.Printf("Test failed: %v\n", err)
	// 	return
	// }
	// fmt.Println("Test completed successfully")

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
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-API-Key",
		ExposeHeaders:    "Content-Length, Access-Control-Allow-Origin",
		AllowCredentials: false,
		MaxAge:           300,
	}))

	go api_handlers.ManageTableReservationStatus()
	go api_handlers.AutoManageReservations()

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Get("/ws/tables", utils.WebSocketAPIKeyMiddleware, api_handlers.TableWebSocketHandler(db.DB))

	// app.Use("/ws/printer", api_handlers.HandlePrinterWebSocket)
	app.Use("/ws/printer", service.HandlePrinterWebSocket)

	routes.SetupRoutes(app)

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}

// type Command struct {
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

// func (c *Command) AddThaiString(s string) {
// 	tis620Bytes := utf8ToTis620(s)
// 	c.data = append(c.data, tis620Bytes...)
// }

// func utf8ToTis620(s string) []byte {
// 	result := make([]byte, 0, len(s))
// 	for _, r := range s {
// 		if r <= 0xFF {
// 			result = append(result, byte(r))
// 		} else if r >= 0x0E01 && r <= 0x0E5B {
// 			result = append(result, byte(r-0x0D60))
// 		}
// 	}
// 	return result
// }

// func testCharacterControl(printerIP string, port int) error {
// 	controls := []struct {
// 		name string
// 		cmds [][]byte
// 	}{
// 		// ชุดทดสอบพื้นฐาน TIS-620
// 		{
// 			"TIS-620 Basic", [][]byte{
// 				{0x1B, 0x40},       // Initialize
// 				{0x1B, 0x74, 0x11}, // TIS-620
// 				{0x1B, 0x52, 0x0D}, // Thai character set
// 			},
// 		},
// 		// ชุดทดสอบพื้นฐาน CP-874
// 		{
// 			"CP-874 Basic", [][]byte{
// 				{0x1B, 0x40},       // Initialize
// 				{0x1B, 0x74, 0x1C}, // CP-874
// 				{0x1B, 0x52, 0x0D}, // Thai character set
// 			},
// 		},
// 		// ชุดทดสอบ Thai 3-Pass
// 		{
// 			"Thai 3-Pass Mode", [][]byte{
// 				{0x1B, 0x40},       // Initialize
// 				{0x1B, 0x74, 0x1C}, // CP-874
// 				{0x1B, 0x52, 0x0D}, // Thai character set
// 				{0x1C, 0x43, 0x01}, // Thai 3-pass mode
// 			},
// 		},
// 		// ชุดทดสอบตัวหนา
// 		{
// 			"Bold Thai", [][]byte{
// 				{0x1B, 0x40},       // Initialize
// 				{0x1B, 0x74, 0x1C}, // CP-874
// 				{0x1B, 0x45, 0x01}, // Bold
// 				{0x1C, 0x43, 0x01}, // Thai 3-pass
// 			},
// 		},
// 		// ชุดทดสอบขนาดตัวอักษร
// 		{
// 			"Double Size Thai", [][]byte{
// 				{0x1B, 0x40},       // Initialize
// 				{0x1B, 0x74, 0x1C}, // CP-874
// 				{0x1B, 0x21, 0x30}, // Double width & height
// 				{0x1C, 0x43, 0x01}, // Thai 3-pass
// 			},
// 		},
// 		// ชุดทดสอบระยะห่าง
// 		{
// 			"Thai Spacing", [][]byte{
// 				{0x1B, 0x40},       // Initialize
// 				{0x1B, 0x74, 0x1C}, // CP-874
// 				{0x1B, 0x20, 0x08}, // Character spacing
// 				{0x1B, 0x33, 0x24}, // Line spacing
// 				{0x1C, 0x43, 0x01}, // Thai 3-pass
// 			},
// 		},
// 		// ชุดทดสอบแบบผสม
// 		{
// 			"Mixed Thai Format", [][]byte{
// 				{0x1B, 0x40},       // Initialize
// 				{0x1B, 0x74, 0x1C}, // CP-874
// 				{0x1B, 0x21, 0x30}, // Double size
// 				{0x1B, 0x45, 0x01}, // Bold
// 				{0x1B, 0x2D, 0x01}, // Underline
// 				{0x1C, 0x43, 0x01}, // Thai 3-pass
// 			},
// 		},
// 	}

// 	thaiTestStrings := []string{
// 		"ทดสอบภาษาไทย",
// 		"สวัสดีครับ/ค่ะ",
// 		"กขคงจฉชซ",
// 		"ฟหกด่าสวงเืท",
// 		"๐๑๒๓๔๕๖๗๘๙",
// 	}

// 	for _, control := range controls {
// 		cmd := NewCommand()

// 		if control.name == "Code Page Test" {
// 			// ทดสอบทุก code page ที่เป็นไปได้
// 			for i := 0; i < 256; i++ {
// 				cmd.AddBytes([]byte{0x1B, 0x40})          // Initialize
// 				cmd.AddBytes([]byte{0x1B, 0x74, byte(i)}) // Set code page
// 				cmd.AddString(fmt.Sprintf("\n=== Code Page %d ===\n", i))

// 				for _, thaiStr := range thaiTestStrings {
// 					cmd.AddThaiString(thaiStr + "\n")
// 				}
// 				cmd.AddString("---------------\n")
// 				time.Sleep(1 * time.Second)
// 			}
// 		} else {
// 			cmd.AddString(fmt.Sprintf("\n=== %s ===\n", control.name))

// 			// ใส่คำสั่งควบคุมเครื่องพิมพ์
// 			for _, command := range control.cmds {
// 				cmd.AddBytes(command)
// 			}

// 			// พิมพ์ข้อความทดสอบ
// 			for _, thaiStr := range thaiTestStrings {
// 				cmd.AddThaiString(thaiStr + "\n")
// 			}
// 			cmd.AddString("---------------\n")
// 		}

// 		// ตัดกระดาษ
// 		cmd.AddBytes([]byte{0x1D, 0x56, 0x41, 0x03})

// 		// เชื่อมต่อและส่งข้อมูล
// 		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", printerIP, port))
// 		if err != nil {
// 			return fmt.Errorf("connection failed for %s: %v", control.name, err)
// 		}

// 		_, err = conn.Write(cmd.data)
// 		if err != nil {
// 			conn.Close()
// 			return fmt.Errorf("write failed for %s: %v", control.name, err)
// 		}

// 		conn.Close()
// 		time.Sleep(4 * time.Second) // รอระหว่างแต่ละชุดทดสอบ
// 	}

// 	return nil
// }

// func generateDashedLine(lineWidth int) string {
// 	// คำนวณจำนวนขีดที่ต้องการ โดยหารด้วยความกว้างของเครื่องหมาย '-'
// 	// ในที่นี้ใช้ 9 พิกเซลต่อขีด (ปรับตามขนาดฟอนต์และ DPI)
// 	count := lineWidth / 9
// 	// สร้าง slice ของเครื่องหมายขีด
// 	dashes := make([]byte, count)
// 	for i := range dashes {
// 		dashes[i] = '-'
// 	}
// 	return string(dashes)
// }

// type wordDict map[string]bool

// func loadDictionary(filePath string) (wordDict, error) {
// 	// อ่านข้อมูลจากไฟล์ lexitron.txt
// 	fileData, err := os.ReadFile(filePath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// แยกคำเป็น slice
// 	words := strings.Split(string(fileData), "\n")

// 	// สร้าง wordDict
// 	dict := make(wordDict)
// 	for _, word := range words {
// 		dict[word] = true
// 	}

// 	return dict, nil
// }

// func segmentText(text string, maxWidth int, d *font.Drawer) []string {
// 	var segments []string
// 	currentSegment := ""

// 	// แยกข้อความออกเป็นคำๆ
// 	words := strings.Fields(text)

// 	for _, word := range words {
// 		// ทดสอบความยาวของประโยคปัจจุบัน
// 		testSegment := currentSegment + word
// 		if currentSegment != "" {
// 			testSegment = currentSegment + " " + word
// 		}

// 		// วัดความกว้างของข้อความ
// 		testWidth := d.MeasureString(testSegment).Ceil()

// 		if testWidth > maxWidth {
// 			// ถ้าความยาวเกินที่กำหนด ให้บันทึกส่วนปัจจุบัน
// 			if currentSegment != "" {
// 				segments = append(segments, strings.TrimSpace(currentSegment))
// 			}
// 			currentSegment = word
// 		} else {
// 			// ถ้าสามารถใส่คำได้ ให้เพิ่มคำ
// 			if currentSegment == "" {
// 				currentSegment = word
// 			} else {
// 				currentSegment += " " + word
// 			}
// 		}
// 	}

// 	// เพิ่มส่วนสุดท้าย
// 	if currentSegment != "" {
// 		segments = append(segments, strings.TrimSpace(currentSegment))
// 	}

// 	return segments
// }

// func testDirectImagePrint(printerIP string, port int) error {
// 	// กำหนดค่าพื้นฐาน
// 	width := 576 //640
// 	dpi := 203.0
// 	fontSize := 14.0
// 	lineSpacing := 2.5
// 	leftPadding := 10

// 	// คำนวณความกว้างที่พิมพ์ได้จริง
// 	printableWidth := width - (leftPadding * 2)

// 	// โหลดฟอนต์
// 	fontBytes, err := os.ReadFile("THSarabunNew.ttf")
// 	if err != nil {
// 		return fmt.Errorf("error loading font: %v", err)
// 	}

// 	f, err := truetype.Parse(fontBytes)
// 	if err != nil {
// 		return fmt.Errorf("error parsing font: %v", err)
// 	}

// 	// สร้าง font.Face
// 	face := truetype.NewFace(f, &truetype.Options{
// 		Size:    fontSize,
// 		DPI:     dpi,
// 		Hinting: font.HintingNone,
// 	})
// 	defer face.Close()

// 	// สร้าง drawer
// 	d := &font.Drawer{
// 		Dst:  nil, // จะกำหนดหลังจากสร้างภาพ
// 		Src:  image.Black,
// 		Face: face,
// 	}

// 	// สร้างเส้นขีด
// 	dashedLine := generateDashedLine(printableWidth)

// 	// ข้อความตัวอย่าง
// 	headerText := []string{
// 		"ร้านอาหารครัวคุณแม่",
// 		"123 ถนนสุขุมวิท แขวงคลองเตย",
// 		"เขตคลองเตย กรุงเทพฯ 10110",
// 		"โทร. 02-123-4567",
// 	}

// 	contentText := []string{
// 		dashedLine,
// 		"ใบเสร็จรับเงิน/ใบกำกับภาษีอย่างย่อ",
// 		"วันที่: " + time.Now().Format("02/01/2006 15:04:05"),
// 		"โต๊ะที่: A1",
// 		"พนักงาน: สมชาย",
// 		dashedLine,
// 		"1. ข้าวผัดหมู ของพิเศษ เสิร์ฟพร้อมไข่ดาวและของตกแต่งพิเศษ x1    60.00",
// 		"   - ไข่ดาว                        +10.00",
// 		"   - พิเศษ                         +20.00",
// 		"2. ต้มยำกุ้งสดใหม่ ปรุงพิเศษ รสชาติเข้มข้น x1   120.00",
// 		"3. น้ำเปล่า                   x2    30.00",
// 		dashedLine,
// 		"รวม                               240.00",
// 		"Service Charge 10%                24.00",
// 		"VAT 7%                            16.80",
// 		"รวมทั้งสิ้น                      280.80",
// 		dashedLine,
// 		"เงินสด                           300.00",
// 		"เงินทอน                           19.20",
// 		dashedLine,
// 	}

// 	footerText := []string{
// 		"",
// 		"ขอบคุณที่ใช้บริการ",
// 		"Welcome back again",
// 		"",
// 	}

// 	// เตรียมข้อมูลที่จะวาด
// 	var allLines []string

// 	// Wrap header text
// 	allLines = append(allLines, headerText...)

// 	// Wrap content text with proper line wrapping
// 	for _, text := range contentText {
// 		wrappedLines := segmentText(text, printableWidth, d)
// 		allLines = append(allLines, wrappedLines...)
// 	}

// 	// Wrap footer text
// 	allLines = append(allLines, footerText...)

// 	// นับจำนวนบรรทัดทั้งหมด
// 	totalLines := len(allLines)

// 	// คำนวณความสูงของภาพ
// 	height := int(float64(totalLines)*fontSize*lineSpacing) + 100

// 	// สร้างภาพใหม่พื้นขาว
// 	img := image.NewRGBA(image.Rect(0, 0, width, height))
// 	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

// 	// กำหนดภาพให้กับ drawer
// 	d.Dst = img

// 	// ฟังก์ชันจัดกึ่งกลาง
// 	drawCenteredString := func(text string, y int) error {
// 		textWidth := d.MeasureString(text).Ceil()
// 		x := leftPadding + (printableWidth-textWidth)/2
// 		d.Dot = fixed.Point26_6{
// 			X: fixed.I(x),
// 			Y: fixed.I(y),
// 		}
// 		d.DrawString(text)
// 		return nil
// 	}

// 	// วาดส่วนหัว (จัดกึ่งกลาง)
// 	y := int(fontSize * 3.0)
// 	for _, text := range headerText {
// 		if err := drawCenteredString(text, y); err != nil {
// 			return err
// 		}
// 		y += int(fontSize * lineSpacing)
// 	}

// 	// วาดส่วนเนื้อหา (ชิดซ้าย)
// 	for _, text := range allLines {
// 		d.Dot = fixed.Point26_6{
// 			X: fixed.I(leftPadding),
// 			Y: fixed.I(y),
// 		}
// 		d.DrawString(text)
// 		y += int(fontSize * lineSpacing)
// 	}

// 	// วาดส่วนท้าย (จัดกึ่งกลาง)
// 	for _, text := range footerText {
// 		if text != "" {
// 			if err := drawCenteredString(text, y); err != nil {
// 				return err
// 			}
// 		}
// 		y += int(fontSize * lineSpacing)
// 	}

// 	// แปลงเป็น ESC/POS command
// 	var buf bytes.Buffer

// 	// Initialize printer
// 	buf.Write([]byte{0x1B, 0x40})

// 	// Set character size and font
// 	buf.Write([]byte{0x1D, 0x21, 0x00}) // Normal size
// 	buf.Write([]byte{0x1B, 0x4D, 0x00}) // Font A

// 	// Set line spacing
// 	buf.Write([]byte{0x1B, 0x33, 60}) // Set line spacing to 60 dots

// 	// Set print density
// 	buf.Write([]byte{0x1D, 0x7C, 0x08}) // Highest density

// 	// Set bitmap mode
// 	buf.Write([]byte{0x1D, 0x76, 0x30, 0x00})

// 	// Calculate and send bitmap size
// 	widthBytes := (width + 7) / 8
// 	buf.WriteByte(byte(widthBytes & 0xFF))
// 	buf.WriteByte(byte(widthBytes >> 8))
// 	buf.WriteByte(byte(height & 0xFF))
// 	buf.WriteByte(byte(height >> 8))

// 	// Convert to bitmap with improved threshold
// 	for y := 0; y < height; y++ {
// 		for x := 0; x < width; x += 8 {
// 			var b byte
// 			for bit := 0; bit < 8; bit++ {
// 				if x+bit < width {
// 					r, g, b_, _ := img.At(x+bit, y).RGBA()
// 					if (r+g+b_)/3 < 0xCFFF {
// 						b |= 1 << (7 - bit)
// 					}
// 				}
// 			}
// 			buf.WriteByte(b)
// 		}
// 	}

// 	// Feed and cut
// 	buf.Write([]byte{0x1B, 0x64, 0x10})       // Feed more lines
// 	buf.Write([]byte{0x1D, 0x56, 0x41, 0x03}) // Partial cut with feed

// 	// ส่งข้อมูลไปยังเครื่องพิมพ์
// 	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", printerIP, port))
// 	if err != nil {
// 		return fmt.Errorf("connection failed: %v", err)
// 	}
// 	defer conn.Close()

// 	_, err = conn.Write(buf.Bytes())
// 	if err != nil {
// 		return fmt.Errorf("print failed: %v", err)
// 	}

// 	return nil
// }
