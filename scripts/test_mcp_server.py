#!/usr/bin/env python3
"""
Simple MCP Server for testing the enhanced test agent.

This server provides basic tools, resources, and prompts to thoroughly test
the comprehensive testing capabilities of the MCP gateway.
"""

import asyncio
import json
import sys
from typing import Any, Dict, List


class SimpleMCPServer:
    def __init__(self):
        self.tools = {
            "echo": {
                "name": "echo",
                "description": "Echo back the provided message",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "message": {
                            "type": "string",
                            "description": "The message to echo back",
                        }
                    },
                    "required": ["message"],
                },
            },
            "add": {
                "name": "add",
                "description": "Add two numbers together",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "a": {"type": "number", "description": "First number"},
                        "b": {"type": "number", "description": "Second number"},
                    },
                    "required": ["a", "b"],
                },
            },
            "list_files": {
                "name": "list_files",
                "description": "List files in the current directory",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": "Directory path to list (optional)",
                            "default": ".",
                        }
                    },
                },
            },
        }

        self.resources = {
            "config://test": {
                "uri": "config://test",
                "name": "Test Config",
                "description": "A test configuration resource",
                "mimeType": "application/json",
            },
            "data://sample": {
                "uri": "data://sample",
                "name": "Sample Data",
                "description": "Sample data for testing",
                "mimeType": "text/plain",
            },
        }

        self.prompts = {
            "greeting": {
                "name": "greeting",
                "description": "Generate a personalized greeting",
                "arguments": [
                    {
                        "name": "name",
                        "description": "Name of the person to greet",
                        "required": True,
                    },
                    {
                        "name": "language",
                        "description": "Language for the greeting",
                        "required": False,
                    },
                ],
            },
            "summary": {
                "name": "summary",
                "description": "Generate a summary prompt",
                "arguments": [
                    {
                        "name": "topic",
                        "description": "Topic to summarize",
                        "required": True,
                    }
                ],
            },
        }

    async def handle_request(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Handle incoming JSON-RPC requests"""
        method = request.get("method")
        params = request.get("params", {})
        request_id = request.get("id")

        try:
            if method == "initialize":
                return {
                    "jsonrpc": "2.0",
                    "id": request_id,
                    "result": {
                        "protocolVersion": "2024-11-05",
                        "capabilities": {
                            "tools": {"listChanged": True},
                            "resources": {"subscribe": True, "listChanged": True},
                            "prompts": {"listChanged": True},
                        },
                        "serverInfo": {
                            "name": "simple-test-server",
                            "version": "1.0.0",
                        },
                    },
                }

            elif method == "tools/list":
                return {
                    "jsonrpc": "2.0",
                    "id": request_id,
                    "result": {"tools": list(self.tools.values())},
                }

            elif method == "tools/call":
                tool_name = params.get("name")
                arguments = params.get("arguments", {})

                if tool_name == "echo":
                    message = arguments.get("message", "Hello!")
                    return {
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "result": {
                            "content": [{"type": "text", "text": f"Echo: {message}"}]
                        },
                    }

                elif tool_name == "add":
                    a = arguments.get("a", 0)
                    b = arguments.get("b", 0)
                    result = a + b
                    return {
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "result": {
                            "content": [
                                {"type": "text", "text": f"{a} + {b} = {result}"}
                            ]
                        },
                    }

                elif tool_name == "list_files":
                    import os

                    path = arguments.get("path", ".")
                    try:
                        files = os.listdir(path)
                        return {
                            "jsonrpc": "2.0",
                            "id": request_id,
                            "result": {
                                "content": [
                                    {
                                        "type": "text",
                                        "text": f"Files in {path}: {', '.join(files[:10])}"
                                        + ("..." if len(files) > 10 else ""),
                                    }
                                ]
                            },
                        }
                    except Exception as e:
                        return {
                            "jsonrpc": "2.0",
                            "id": request_id,
                            "error": {"code": -32000, "message": str(e)},
                        }

                else:
                    return {
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "error": {
                            "code": -32601,
                            "message": f"Tool '{tool_name}' not found",
                        },
                    }

            elif method == "resources/list":
                return {
                    "jsonrpc": "2.0",
                    "id": request_id,
                    "result": {"resources": list(self.resources.values())},
                }

            elif method == "resources/read":
                uri = params.get("uri")
                if uri == "config://test":
                    return {
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "result": {
                            "contents": [
                                {
                                    "uri": uri,
                                    "mimeType": "application/json",
                                    "text": json.dumps(
                                        {
                                            "test": True,
                                            "server": "simple-mcp",
                                            "version": "1.0",
                                        }
                                    ),
                                }
                            ]
                        },
                    }
                elif uri == "data://sample":
                    return {
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "result": {
                            "contents": [
                                {
                                    "uri": uri,
                                    "mimeType": "text/plain",
                                    "text": "This is sample data from the simple MCP server.\nIt demonstrates resource reading capabilities.\n",
                                }
                            ]
                        },
                    }
                else:
                    return {
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "error": {
                            "code": -32602,
                            "message": f"Resource not found: {uri}",
                        },
                    }

            elif method == "prompts/list":
                return {
                    "jsonrpc": "2.0",
                    "id": request_id,
                    "result": {"prompts": list(self.prompts.values())},
                }

            elif method == "prompts/get":
                prompt_name = params.get("name")
                arguments = params.get("arguments", {})

                if prompt_name == "greeting":
                    name = arguments.get("name", "World")
                    language = arguments.get("language", "English")
                    greeting_map = {
                        "English": "Hello",
                        "Spanish": "Hola",
                        "French": "Bonjour",
                        "German": "Hallo",
                    }
                    greeting = greeting_map.get(language, "Hello")
                    return {
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "result": {
                            "description": "Personalized greeting prompt",
                            "messages": [
                                {
                                    "role": "user",
                                    "content": {
                                        "type": "text",
                                        "text": f"{greeting}, {name}! How can I help you today?",
                                    },
                                }
                            ],
                        },
                    }

                elif prompt_name == "summary":
                    topic = arguments.get("topic", "general topic")
                    return {
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "result": {
                            "description": "Summary generation prompt",
                            "messages": [
                                {
                                    "role": "user",
                                    "content": {
                                        "type": "text",
                                        "text": f"Please provide a comprehensive summary of {topic}. Include key points, important details, and conclusions.",
                                    },
                                }
                            ],
                        },
                    }

                else:
                    return {
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "error": {
                            "code": -32601,
                            "message": f"Prompt '{prompt_name}' not found",
                        },
                    }

            else:
                return {
                    "jsonrpc": "2.0",
                    "id": request_id,
                    "error": {
                        "code": -32601,
                        "message": f"Method '{method}' not found",
                    },
                }

        except Exception as e:
            return {
                "jsonrpc": "2.0",
                "id": request_id,
                "error": {"code": -32603, "message": f"Internal error: {str(e)}"},
            }


async def main():
    server = SimpleMCPServer()

    # Read from stdin and write to stdout for STDIO transport
    while True:
        try:
            line = sys.stdin.readline()
            if not line:
                break

            line = line.strip()
            if not line:
                continue

            # Parse JSON-RPC request
            request = json.loads(line)

            # Handle request
            response = await server.handle_request(request)

            # Send response
            print(json.dumps(response), flush=True)

        except json.JSONDecodeError as e:
            error_response = {
                "jsonrpc": "2.0",
                "id": None,
                "error": {"code": -32700, "message": "Parse error"},
            }
            print(json.dumps(error_response), flush=True)
        except Exception as e:
            error_response = {
                "jsonrpc": "2.0",
                "id": None,
                "error": {"code": -32603, "message": f"Internal error: {str(e)}"},
            }
            print(json.dumps(error_response), flush=True)


if __name__ == "__main__":
    asyncio.run(main())
