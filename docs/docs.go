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
        "/add_category": {
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
                            "$ref": "#/definitions/models.CreateCategoryRequest"
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
        "/add_menu": {
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
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.CreateMenuRequest"
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
                "description": "ฟังก์ชันนี้ใช้สำหรับเรียกข้อมูลเมนู โดยสามารถระบุ action ได้ 3 แบบ",
                "produces": [
                    "application/json"
                ],
                "summary": "เลือกการดำเนินการกับข้อมูลเมนู",
                "parameters": [
                    {
                        "enum": [
                            "getByID",
                            "getByCategory",
                            "getAll"
                        ],
                        "type": "string",
                        "description": "รูปแบบการค้นหาเมนู: getByID, getByCategory หรือ getAll",
                        "name": "action",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "ID ของเมนู (ใช้กับ action=getByID)",
                        "name": "id",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "ID ของหมวดหมู่ (ใช้กับ action=getByCategory)",
                        "name": "category_id",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "รายการเมนูที่ค้นพบ",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.MenuItem"
                            }
                        }
                    },
                    "400": {
                        "description": "เกิดข้อผิดพลาดจากการระบุพารามิเตอร์",
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
                        "name": "table",
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
                "description": "สั่งอาหารสำหรับโต๊ะที่ระบุ",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "สั่งอาหาร......",
                "parameters": [
                    {
                        "description": "ข้อมูลการสั่งอาหาร",
                        "name": "order",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api_handlers.OrderRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "รายละเอียดออเดอร์ที่สร้าง",
                        "schema": {
                            "$ref": "#/definitions/models.Order"
                        }
                    },
                    "400": {
                        "description": "ข้อมูลไม่ถูกต้อง",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "500": {
                        "description": "เกิดข้อผิดพลาดในการประมวลผล",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    }
                }
            }
        },
        "/qr_code/{table}": {
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
                        "name": "table",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "รายละเอียดของตาราง qr_code",
                        "schema": {
                            "$ref": "#/definitions/models.QRCode"
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
        "api_handlers.OrderItemRequest": {
            "type": "object",
            "properties": {
                "menu_item_id": {
                    "type": "integer"
                },
                "notes": {
                    "type": "string"
                },
                "quantity": {
                    "type": "integer"
                }
            }
        },
        "api_handlers.OrderRequest": {
            "type": "object",
            "properties": {
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/api_handlers.OrderItemRequest"
                    }
                },
                "table_id": {
                    "type": "integer"
                },
                "uuid": {
                    "type": "string"
                }
            }
        },
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
        "models.CreateCategoryRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "name": {
                    "type": "string"
                }
            }
        },
        "models.CreateMenuRequest": {
            "type": "object",
            "required": [
                "CategoryID",
                "Name",
                "Price"
            ],
            "properties": {
                "CategoryID": {
                    "type": "integer",
                    "minimum": 1
                },
                "Name": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 2
                },
                "Price": {
                    "type": "integer",
                    "minimum": 1
                },
                "description": {
                    "type": "string",
                    "maxLength": 255
                },
                "image": {
                    "description": "max 5MB",
                    "type": "array",
                    "maxItems": 5242880,
                    "items": {
                        "type": "integer"
                    }
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
                "price": {
                    "type": "integer"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "models.Order": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.OrderItem"
                    }
                },
                "status": {
                    "description": "e.g., \"pending\", \"cooking\", \"served\", \"completed\"",
                    "type": "string"
                },
                "tableID": {
                    "description": "Foreign key to Table",
                    "type": "integer"
                },
                "total": {
                    "type": "number"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "models.OrderItem": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "menuItem": {
                    "$ref": "#/definitions/models.MenuItem"
                },
                "menuItemID": {
                    "description": "Foreign key to MenuItem",
                    "type": "integer"
                },
                "notes": {
                    "description": "Special instructions",
                    "type": "string"
                },
                "order": {
                    "$ref": "#/definitions/models.Order"
                },
                "orderID": {
                    "description": "Foreign key to Order",
                    "type": "integer"
                },
                "price": {
                    "description": "Price at time of order",
                    "type": "number"
                },
                "quantity": {
                    "type": "integer"
                },
                "status": {
                    "description": "e.g., \"pending\", \"cooking\", \"served\"",
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "models.QRCode": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "expiry_at": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "is_active": {
                    "type": "boolean"
                },
                "qr_Image": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "table_id": {
                    "description": "One-to-One กับ Table",
                    "type": "integer"
                },
                "uuid": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "127.0.0.1:8080",
	BasePath:         "/",
	Schemes:          []string{"http"},
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
