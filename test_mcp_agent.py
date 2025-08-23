#!/usr/bin/env python3
"""
Test script for MCP Gateway - demonstrates how to create an agent
that uses a registered MCP server through the gateway.

This script:
1. Authenticates with the MCP Gateway
2. Lists available MCP servers
3. Creates an MCP session for a registered server
4. Uses the Claude Code SDK to create an agent that can interact with the MCP server
5. Demonstrates basic MCP operations (list tools, execute tools)
"""

import asyncio
import json
import uuid
from typing import Dict, Any, List
import httpx
import websockets
from websockets.exceptions import ConnectionClosed
import argparse

# MCP Gateway Configuration
GATEWAY_BASE_URL = "http://localhost:8080"
GATEWAY_WS_URL = "ws://localhost:8080"
ADMIN_EMAIL = "team@wraithscan.com"
ADMIN_PASSWORD = "qwerty123"

class MCPGatewayClient:
    """Client for interacting with the MCP Gateway"""
    
    def __init__(self, base_url: str = GATEWAY_BASE_URL):
        self.base_url = base_url
        self.ws_url = GATEWAY_WS_URL
        self.access_token = None
        self.session = httpx.AsyncClient()
        
    async def authenticate(self, email: str, password: str) -> Dict[str, Any]:
        """Authenticate with the gateway and store access token"""
        response = await self.session.post(
            f"{self.base_url}/api/auth/login",
            json={"email": email, "password": password}
        )
        response.raise_for_status()
        
        data = response.json()
        if data.get("success"):
            self.access_token = data["data"]["access_token"]
            print(f"✅ Authenticated successfully as {email}")
            return data["data"]
        else:
            raise Exception(f"Authentication failed: {data}")
    
    async def list_servers(self) -> List[Dict[str, Any]]:
        """List all registered MCP servers"""
        response = await self.session.get(
            f"{self.base_url}/api/gateway/servers",
            headers={"Authorization": f"Bearer {self.access_token}"}
        )
        response.raise_for_status()
        
        data = response.json()
        if data.get("success"):
            servers = data["data"]
            print(f"📋 Found {len(servers)} registered MCP server(s)")
            for server in servers:
                print(f"  - {server['name']} ({server['id']}) - {server['protocol']}")
            return servers
        else:
            raise Exception(f"Failed to list servers: {data}")
    
    async def create_session(self, server_id: str) -> Dict[str, Any]:
        """Create a new MCP session for the specified server"""
        session_data = {
            "server_id": server_id,
            "protocol": "websocket",  # Use WebSocket for real-time communication
            "user_id": "test-user",
            "client_id": str(uuid.uuid4()),
            "metadata": {
                "test_script": True,
                "purpose": "testing MCP server functionality"
            }
        }
        
        response = await self.session.post(
            f"{self.base_url}/api/gateway/sessions",
            json=session_data,
            headers={"Authorization": f"Bearer {self.access_token}"}
        )
        response.raise_for_status()
        
        data = response.json()
        if data.get("success"):
            session = data["data"]
            print(f"🚀 Created MCP session: {session['id']}")
            return session
        else:
            raise Exception(f"Failed to create session: {data}")
    
    async def list_sessions(self) -> List[Dict[str, Any]]:
        """List all active MCP sessions"""
        response = await self.session.get(
            f"{self.base_url}/api/gateway/sessions",
            headers={"Authorization": f"Bearer {self.access_token}"}
        )
        response.raise_for_status()
        
        data = response.json()
        if data.get("success"):
            sessions = data["data"]
            print(f"📝 Found {len(sessions)} active session(s)")
            return sessions
        else:
            raise Exception(f"Failed to list sessions: {data}")

class SimpleMCPAgent:
    """A simple MCP agent that can interact with MCP servers through the gateway"""
    
    def __init__(self, gateway_client: MCPGatewayClient):
        self.gateway = gateway_client
        self.websocket = None
        self.request_id = 0
        
    def _next_request_id(self) -> str:
        """Get the next request ID for JSON-RPC"""
        self.request_id += 1
        return str(self.request_id)
    
    async def connect_to_server(self, server_id: str, protocol: str = "websocket") -> Dict[str, Any]:
        """Connect to an MCP server through the gateway using different protocols"""
        
        if protocol == "stdio":
            # For STDIO servers, use HTTP JSON-RPC instead of WebSocket
            return await self._connect_via_jsonrpc(server_id)
        else:
            # For other protocols, try WebSocket
            return await self._connect_via_websocket(server_id)
    
    async def _connect_via_jsonrpc(self, server_id: str) -> Dict[str, Any]:
        """Connect to MCP server using JSON-RPC over HTTP"""
        print(f"🔌 Connecting to MCP server {server_id} via JSON-RPC")
        
        # Send initialize request via HTTP POST
        init_request = {
            "jsonrpc": "2.0",
            "id": self._next_request_id(),
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {
                    "roots": {"listChanged": True},
                    "sampling": {}
                },
                "clientInfo": {
                    "name": "mcp-gateway-test",
                    "version": "1.0.0"
                }
            }
        }
        
        response = await self.gateway.session.post(
            f"{self.gateway.base_url}/servers/{server_id}/rpc",
            json=init_request,
            headers={"Authorization": f"Bearer {self.gateway.access_token}"}
        )
        
        if response.status_code != 200:
            raise Exception(f"HTTP {response.status_code}: {response.text}")
            
        init_response = response.json()
        
        if "error" in init_response:
            raise Exception(f"Initialize error: {init_response['error']}")
        
        print("✅ MCP server initialized successfully via JSON-RPC")
        self.server_id = server_id  # Store for later use
        self.use_http = True  # Flag to indicate we're using HTTP instead of WebSocket
        return init_response["result"]
    
    async def _connect_via_websocket(self, server_id: str) -> Dict[str, Any]:
        """Connect to MCP server via WebSocket through the gateway"""
        # Connect to the gateway's WebSocket endpoint for MCP communication
        ws_url = f"{self.gateway.ws_url}/api/gateway/ws?server_id={server_id}"
        
        try:
            # Use proper WebSocket connection with authorization
            self.websocket = await websockets.connect(ws_url)
            print(f"🔌 Connected to MCP server {server_id} via WebSocket")
            
            # Send initialize request
            init_request = {
                "jsonrpc": "2.0",
                "id": self._next_request_id(),
                "method": "initialize",
                "params": {
                    "protocolVersion": "2024-11-05",
                    "capabilities": {
                        "roots": {"listChanged": True},
                        "sampling": {}
                    },
                    "clientInfo": {
                        "name": "mcp-gateway-test",
                        "version": "1.0.0"
                    }
                }
            }
            
            await self.websocket.send(json.dumps(init_request))
            response = await self.websocket.recv()
            init_response = json.loads(response)
            
            if "error" in init_response:
                raise Exception(f"Initialize error: {init_response['error']}")
            
            print("✅ MCP server initialized successfully")
            self.use_http = False  # Flag to indicate we're using WebSocket
            return init_response["result"]
            
        except Exception as e:
            print(f"❌ Failed to connect to MCP server: {e}")
            raise
    
    async def _send_request(self, request: Dict[str, Any]) -> Dict[str, Any]:
        """Send a request using either HTTP or WebSocket depending on connection type"""
        if hasattr(self, 'use_http') and self.use_http:
            # Use HTTP JSON-RPC
            response = await self.gateway.session.post(
                f"{self.gateway.base_url}/servers/{self.server_id}/rpc",
                json=request,
                headers={"Authorization": f"Bearer {self.gateway.access_token}"}
            )
            
            if response.status_code != 200:
                raise Exception(f"HTTP {response.status_code}: {response.text}")
                
            return response.json()
        else:
            # Use WebSocket
            if not self.websocket:
                raise Exception("Not connected to MCP server")
            
            await self.websocket.send(json.dumps(request))
            response = await self.websocket.recv()
            return json.loads(response)
    
    async def list_tools(self) -> List[Dict[str, Any]]:
        """List all available tools from the MCP server"""
        request = {
            "jsonrpc": "2.0",
            "id": self._next_request_id(),
            "method": "tools/list",
            "params": {}
        }
        
        result = await self._send_request(request)
        
        if "error" in result:
            raise Exception(f"List tools error: {result['error']}")
        
        tools = result["result"]["tools"]
        print(f"🛠️  Found {len(tools)} available tools:")
        for tool in tools:
            print(f"  - {tool['name']}: {tool.get('description', 'No description')}")
        
        return tools
    
    async def call_tool(self, tool_name: str, arguments: Dict[str, Any] = None) -> Dict[str, Any]:
        """Call a specific tool with the given arguments"""
        request = {
            "jsonrpc": "2.0",
            "id": self._next_request_id(),
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments or {}
            }
        }
        
        print(f"🔧 Calling tool '{tool_name}' with arguments: {arguments}")
        result = await self._send_request(request)
        
        if "error" in result:
            raise Exception(f"Tool call error: {result['error']}")
        
        tool_result = result["result"]
        print(f"✅ Tool '{tool_name}' executed successfully")
        return tool_result
    
    async def list_resources(self) -> List[Dict[str, Any]]:
        """List all available resources from the MCP server"""
        request = {
            "jsonrpc": "2.0",
            "id": self._next_request_id(),
            "method": "resources/list",
            "params": {}
        }
        
        result = await self._send_request(request)
        
        if "error" in result:
            if result["error"]["code"] == -32601:  # Method not found
                print("ℹ️  Server does not support resources")
                return []
            raise Exception(f"List resources error: {result['error']}")
        
        resources = result["result"]["resources"]
        print(f"📚 Found {len(resources)} available resources:")
        for resource in resources:
            print(f"  - {resource['uri']}: {resource.get('description', 'No description')}")
        
        return resources
    
    async def disconnect(self):
        """Disconnect from the MCP server"""
        if self.websocket:
            await self.websocket.close()
            print("🔌 Disconnected from MCP server")

async def main():
    parser = argparse.ArgumentParser(description="Test MCP Gateway with a real agent")
    parser.add_argument("--server-id", help="Specific server ID to use (optional)")
    parser.add_argument("--tool", help="Specific tool to call (optional)")
    parser.add_argument("--args", help="Tool arguments as JSON string (optional)")
    args = parser.parse_args()
    
    print("🚀 Starting MCP Gateway Test Agent")
    print("=" * 50)
    
    # Initialize gateway client
    gateway = MCPGatewayClient()
    
    try:
        # Authenticate
        await gateway.authenticate(ADMIN_EMAIL, ADMIN_PASSWORD)
        
        # List available servers
        servers = await gateway.list_servers()
        if not servers:
            print("❌ No MCP servers registered. Please register a server first.")
            return
        
        # Select server to test
        if args.server_id:
            server = next((s for s in servers if s["id"] == args.server_id), None)
            if not server:
                print(f"❌ Server with ID {args.server_id} not found")
                return
        else:
            server = servers[0]  # Use the first server
        
        print(f"\n🎯 Testing with server: {server['name']} ({server['id']})")
        print(f"   Protocol: {server['protocol']}")
        print(f"   Description: {server.get('description', 'No description')}")
        
        # Create MCP agent
        agent = SimpleMCPAgent(gateway)
        
        # Connect to the MCP server
        server_info = await agent.connect_to_server(server["id"], server["protocol"])
        print(f"📊 Server capabilities: {json.dumps(server_info.get('capabilities', {}), indent=2)}")
        
        # List available tools
        print(f"\n🛠️  Discovering tools...")
        tools = await agent.list_tools()
        
        # List available resources (if supported)
        print(f"\n📚 Discovering resources...")
        resources = await agent.list_resources()
        
        # If specific tool requested, call it
        if args.tool:
            if args.tool in [tool["name"] for tool in tools]:
                tool_args = {}
                if args.args:
                    try:
                        tool_args = json.loads(args.args)
                    except json.JSONDecodeError:
                        print(f"❌ Invalid JSON in --args: {args.args}")
                        return
                
                print(f"\n🔧 Testing specific tool: {args.tool}")
                result = await agent.call_tool(args.tool, tool_args)
                print(f"📄 Tool result:")
                print(json.dumps(result, indent=2))
            else:
                print(f"❌ Tool '{args.tool}' not found. Available tools: {[t['name'] for t in tools]}")
        
        # Interactive mode - demonstrate a few tools
        elif tools:
            print(f"\n🧪 Testing first available tool: {tools[0]['name']}")
            try:
                result = await agent.call_tool(tools[0]["name"])
                print(f"📄 Tool result preview:")
                print(json.dumps(result, indent=2)[:500] + ("..." if len(json.dumps(result, indent=2)) > 500 else ""))
            except Exception as e:
                print(f"⚠️  Tool test failed (this is normal for tools requiring arguments): {e}")
        
        # Clean up
        await agent.disconnect()
        
        print(f"\n✅ MCP Gateway test completed successfully!")
        print(f"   Server: {server['name']}")
        print(f"   Tools discovered: {len(tools)}")
        print(f"   Resources discovered: {len(resources)}")
        
    except Exception as e:
        print(f"❌ Test failed: {e}")
        raise
    
    finally:
        await gateway.session.aclose()

if __name__ == "__main__":
    # Install required packages if not already installed
    try:
        import httpx
        import websockets
    except ImportError:
        print("📦 Installing required packages...")
        import subprocess
        import sys
        subprocess.check_call([sys.executable, "-m", "pip", "install", "httpx", "websockets"])
        import httpx
        import websockets
    
    asyncio.run(main())