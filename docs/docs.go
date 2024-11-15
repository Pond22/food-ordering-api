// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "https://example.com/terms",
        "contact": {
            "name": "API Support",
            "url": "https://www.example.com/support",
            "email": "support@example.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/categories": {
            "post": {
                "description": "ฟังก์ชันนี้ใช้สำหรับสร้างหมวดหมู่ใหม่ โดยต้องระบุข้อมูลชื่อหมวดหมู่",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "สร้างหมวดหมู่ใหม่",
                "parameters": [
                    {
                        "description": "ข้อมูลหมวดหมู่ใหม่",
                        "name": "category",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.Category"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "รายละเอียดของหมวดหมู่ที่สร้างเสร็จแล้ว",
                        "schema": {
                            "$ref": "#/definitions/models.Category"
                        }
                    },
                    "400": {
                        "description": "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "500": {
                        "description": "เกิดข้อผิดพลาดในการสร้างหมวดหมู่ใหม่",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/getCategory": {
            "get": {
                "description": "ฟังก์ชันนี้ใช้สำหรับเรียกข้อมูลหมวดหมู่ทั้งหมดที่มีอยู่ในระบบ",
                "produces": [
                    "application/json"
                ],
                "summary": "เรียกรายการหมวดหมู่ทั้งหมด",
                "responses": {
                    "200": {
                        "description": "รายการหมวดหมู่ทั้งหมด",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Category"
                            }
                        }
                    },
                    "500": {
                        "description": "เกิดข้อผิดพลาดในการดึงข้อมูลหมวดหมู่",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/menu": {
            "get": {
                "description": "ฟังก์ชันนี้ใช้สำหรับเรียกข้อมูลเมนู โดยสามารถระบุ action ได้ 3 แบบ คือ getByID, getByCategory, และ getAll เพื่อกำหนดรูปแบบการค้นหาเมนูตาม action ที่ระบุเนื่องจากใช้หลักการ Switchcase ทำให้ใน Document มีตัวอย่างเพัยงแค่การเลือก Actions",
                "produces": [
                    "application/json"
                ],
                "summary": "เลือกการดำเนินการกับข้อมูลเมนู",
                "parameters": [
                    {
                        "type": "string",
                        "description": "รูปแบบการค้นหาเมนู: getByID, getByCategory หรือ getAll",
                        "name": "action",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "รายละเอียดของเมนูที่ค้นพบจากการค้นหา",
                        "schema": {
                            "$ref": "#/definitions/models.MenuItem"
                        }
                    },
                    "400": {
                        "description": "เกิดข้อผิดพลาดจาก action ที่ไม่ถูกต้อง",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            },
            "post": {
                "description": "ฟังก์ชันนี้ใช้สำหรับสร้างเมนูใหม่ โดยต้องระบุข้อมูลที่จำเป็นในการสร้าง เช่น ชื่อเมนูและ ID ของหมวดหมู่ที่เกี่ยวข้อง",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "สร้างเมนูใหม่",
                "parameters": [
                    {
                        "description": "ข้อมูลเมนูใหม่",
                        "name": "menuItem",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.MenuItem"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "รายละเอียดของเมนูที่สร้างเสร็จแล้ว",
                        "schema": {
                            "$ref": "#/definitions/models.MenuItem"
                        }
                    },
                    "400": {
                        "description": "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "500": {
                        "description": "เกิดข้อผิดพลาดในการสร้างเมนูใหม่",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/migrate": {
            "post": {
                "description": "api เส้นนี้ใช้เพื่อ migrate ฐานข้อมูลของ orm",
                "produces": [
                    "application/json"
                ],
                "summary": "อัปเดต orm",
                "responses": {}
            }
        },
        "/order": {
            "get": {
                "description": "เข้าสู่โต๊ะนั้นๆ ซึ่ง api เส้นนี้ไม่จำเป็นต้องถูกใช้งานโดยตรงเพราะ url ของแต่ละโต๊ะจะสามารถเข้าได้ผ่าน qr_code เท่านั้นจากฟังก์ชัน",
                "produces": [
                    "application/json"
                ],
                "summary": "เข้าถึง dynamic link โต๊ะนั้นๆ",
                "parameters": [
                    {
                        "type": "string",
                        "description": "เลขโต๊ะ",
                        "name": "action",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "รายละเอียดของเมนูที่ค้นพบจากการค้นหา",
                        "schema": {
                            "$ref": "#/definitions/models.MenuItem"
                        }
                    },
                    "400": {
                        "description": "เกิดข้อผิดพลาดจาก action ที่ไม่ถูกต้อง",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            },
            "post": {
                "description": "สั่งอาหารของโต๊ะ",
                "produces": [
                    "application/json"
                ],
                "summary": "สั่งอาหาร",
                "parameters": [
                    {
                        "type": "string",
                        "description": "เลขโต๊ะ",
                        "name": "action",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "รายละเอียดของเมนูที่ค้นพบจากการค้นหา",
                        "schema": {
                            "$ref": "#/definitions/models.MenuItem"
                        }
                    },
                    "400": {
                        "description": "เกิดข้อผิดพลาดจาก action ที่ไม่ถูกต้อง",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/qr_code": {
            "get": {
                "description": "เข้าสู่โต๊ะนั้นๆ ซึ่ง api เส้นนี้ไม่จำเป็นต้องถูกใช้งานโดยตรงเพราะ url ของแต่ละโต๊ะจะสามารถเข้าได้ผ่าน qr_code เท่านั้นจากฟังก์ชัน",
                "produces": [
                    "application/json"
                ],
                "summary": "เข้าถึง dynamic link โต๊ะนั้นๆ",
                "responses": {
                    "200": {
                        "description": "รายละเอียดของเมนูที่ค้นพบจากการค้นหา",
                        "schema": {
                            "$ref": "#/definitions/models.MenuItem"
                        }
                    },
                    "400": {
                        "description": "เกิดข้อผิดพลาดจาก action ที่ไม่ถูกต้อง",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "models.Category": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "models.MenuItem": {
            "type": "object",
            "properties": {
                "category": {
                    "description": "ลิงก์ไปยังตาราง Category",
                    "allOf": [
                        {
                            "$ref": "#/definitions/models.Category"
                        }
                    ]
                },
                "categoryID": {
                    "description": "foreign key ที่เชื่อมกับ Category",
                    "type": "integer"
                },
                "createdAt": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "image": {
                    "description": "ฟิลด์ Image เป็น type bytea",
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "name": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Food Ordering API",
	Description:      "This is a simple API for ordering food.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
