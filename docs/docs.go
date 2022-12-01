// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2022-11-29 15:53:02.049203 +0800 CST m=+0.041635126
package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/login/oauth": {
            "get": {
                "description": "第三方登录",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "第三方登录",
                "parameters": [
                    {
                        "type": "string",
                        "name": "code",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "name": "redirect_uri",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "properties": {
                                "code": {
                                    "type": "integer"
                                },
                                "data": {
                                    "$ref": "#/definitions/user.OAuthCallbackResponse"
                                },
                                "msg": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/BadRequest"
                        }
                    }
                }
            }
        },
        "/script": {
            "post": {
                "description": "创建脚本",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "script"
                ],
                "summary": "创建脚本",
                "parameters": [
                    {
                        "type": "object",
                        "name": "body",
                        "in": "body"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "properties": {
                                "code": {
                                    "type": "integer"
                                },
                                "data": {
                                    "$ref": "#/definitions/script.CreateResponse"
                                },
                                "msg": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/BadRequest"
                        }
                    }
                }
            }
        },
        "/user": {
            "get": {
                "description": "获取当前登录的用户信息",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "获取当前登录的用户信息",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "properties": {
                                "code": {
                                    "type": "integer"
                                },
                                "data": {
                                    "$ref": "#/definitions/user.CurrentUserResponse"
                                },
                                "msg": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/BadRequest"
                        }
                    }
                }
            }
        },
        "/user/{uid}/info": {
            "get": {
                "description": "获取指定用户信息",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "获取指定用户信息",
                "parameters": [
                    {
                        "type": "integer",
                        "name": "uid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "properties": {
                                "code": {
                                    "type": "integer"
                                },
                                "data": {
                                    "$ref": "#/definitions/user.InfoResponse"
                                },
                                "msg": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/BadRequest"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "BadRequest": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "错误码",
                    "type": "integer",
                    "format": "int32"
                },
                "msg": {
                    "description": "错误信息",
                    "type": "string"
                }
            }
        },
        "httputils.PageResponse": {
            "type": "object",
            "properties": {
                "list": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "$ref": "#/definitions/script.Item"
                    }
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "script.CreateResponse": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                }
            }
        },
        "script.Item": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                }
            }
        },
        "script.ListResponse": {
            "type": "object",
            "properties": {
                "list": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "$ref": "#/definitions/script.Item"
                    }
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "user.CurrentUserResponse": {
            "type": "object",
            "properties": {
                "avatar": {
                    "type": "string"
                },
                "uid": {
                    "type": "integer"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "user.InfoResponse": {
            "type": "object",
            "properties": {
                "avatar": {
                    "type": "string"
                },
                "uid": {
                    "type": "integer"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "user.OAuthCallbackResponse": {
            "type": "object",
            "properties": {
                "redirect_uri": {
                    "type": "string"
                },
                "uid": {
                    "type": "integer"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "2.0.0",
	Host:        "",
	BasePath:    "/api/v2",
	Schemes:     []string{},
	Title:       "脚本站 API 文档",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"escape": func(v interface{}) string {
			// escape tabs
			str := strings.Replace(v.(string), "\t", "\\t", -1)
			// replace " with \", and if that results in \\", replace that with \\\"
			str = strings.Replace(str, "\"", "\\\"", -1)
			return strings.Replace(str, "\\\\\"", "\\\\\\\"", -1)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
