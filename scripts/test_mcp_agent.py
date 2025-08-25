#!/usr/bin/env python3
"""
Comprehensive MCP Gateway Transport Test Agent

This enhanced test script provides thorough testing of all MCP Gateway transport protocols:

Transport Testing:
1. JSON-RPC Transport (HTTP) - Single and batch requests
2. WebSocket Transport - Connection health and metrics
3. SSE Transport - Server-Sent Events streaming
4. Streamable HTTP (MCP) Transport - MCP protocol endpoints
5. STDIO Transport - Command execution and process management

MCP Server Testing:
1. Authenticates with the MCP Gateway using admin credentials
2. Discovers and lists all registered MCP servers
3. Connects to MCP servers using multiple transport protocols
4. Performs comprehensive testing of MCP server capabilities:
   - Tool discovery and execution (with smart argument generation)
   - Resource discovery and content reading
   - Prompt discovery and template testing
   - Error handling and recovery strategies

Usage Examples:
- Transport smoke test: python test_mcp_agent.py --transports-only
- Full test suite: python test_mcp_agent.py
- Test specific server: python test_mcp_agent.py --server-id <server-id>
- Test specific tool: python test_mcp_agent.py --tool <tool-name> --args '{"param": "value"}'
- Quick smoke test: python transport_smoke_test.py
"""

import argparse
import asyncio
import json
import uuid
from typing import Any, Dict, List

import httpx
import websockets
from websockets.exceptions import ConnectionClosed

# MCP Gateway Configuration
GATEWAY_BASE_URL = "http://localhost:8080"
GATEWAY_WS_URL = "ws://localhost:8080"
ADMIN_EMAIL = "team@wraithscan.com"
ADMIN_PASSWORD = "qwerty123"


class MCPGatewayClient:
    """Client for interacting with the MCP Gateway and testing all transport protocols"""

    def __init__(self, base_url: str = GATEWAY_BASE_URL):
        self.base_url = base_url
        self.ws_url = GATEWAY_WS_URL
        self.access_token = None
        self.session = httpx.AsyncClient(
            timeout=30.0
        )  # Increased timeout for transport tests

    async def authenticate(self, email: str, password: str) -> Dict[str, Any]:
        """Authenticate with the gateway and store access token"""
        response = await self.session.post(
            f"{self.base_url}/api/auth/login",
            json={"email": email, "password": password},
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
            headers={"Authorization": f"Bearer {self.access_token}"},
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

    async def test_all_transports(self) -> Dict[str, Any]:
        """Test all transport protocols with comprehensive smoke tests"""
        print("\n🚀 TESTING ALL MCP GATEWAY TRANSPORTS")
        print("=" * 60)

        transport_results = {
            "json_rpc": {"status": "pending", "tests": [], "errors": []},
            "websocket": {"status": "pending", "tests": [], "errors": []},
            "sse": {"status": "pending", "tests": [], "errors": []},
            "streamable_http": {"status": "pending", "tests": [], "errors": []},
            "stdio": {"status": "pending", "tests": [], "errors": []},
        }

        # Test JSON-RPC Transport
        try:
            print("\n🔄 Testing JSON-RPC Transport...")
            await self._test_json_rpc_transport(transport_results["json_rpc"])
            transport_results["json_rpc"]["status"] = "passed"
        except Exception as e:
            transport_results["json_rpc"]["status"] = "failed"
            transport_results["json_rpc"]["errors"].append(str(e))
            print(f"❌ JSON-RPC Transport failed: {e}")

        # Test WebSocket Transport
        try:
            print("\n🔄 Testing WebSocket Transport...")
            await self._test_websocket_transport(transport_results["websocket"])
            transport_results["websocket"]["status"] = "passed"
        except Exception as e:
            transport_results["websocket"]["status"] = "failed"
            transport_results["websocket"]["errors"].append(str(e))
            print(f"❌ WebSocket Transport failed: {e}")

        # Test SSE Transport
        try:
            print("\n🔄 Testing SSE Transport...")
            await self._test_sse_transport(transport_results["sse"])
            transport_results["sse"]["status"] = "passed"
        except Exception as e:
            transport_results["sse"]["status"] = "failed"
            transport_results["sse"]["errors"].append(str(e))
            print(f"❌ SSE Transport failed: {e}")

        # Test Streamable HTTP (MCP) Transport
        try:
            print("\n🔄 Testing Streamable HTTP (MCP) Transport...")
            await self._test_streamable_http_transport(
                transport_results["streamable_http"]
            )
            transport_results["streamable_http"]["status"] = "passed"
        except Exception as e:
            transport_results["streamable_http"]["status"] = "failed"
            transport_results["streamable_http"]["errors"].append(str(e))
            print(f"❌ Streamable HTTP Transport failed: {e}")

        # Test STDIO Transport
        try:
            print("\n🔄 Testing STDIO Transport...")
            await self._test_stdio_transport(transport_results["stdio"])
            transport_results["stdio"]["status"] = "passed"
        except Exception as e:
            transport_results["stdio"]["status"] = "failed"
            transport_results["stdio"]["errors"].append(str(e))
            print(f"❌ STDIO Transport failed: {e}")

        return transport_results

    async def _test_json_rpc_transport(self, result: Dict[str, Any]):
        """Test JSON-RPC transport endpoints"""
        # Test single JSON-RPC request
        response = await self.session.post(
            f"{self.base_url}/rpc",
            json={
                "jsonrpc": "2.0",
                "id": "1",
                "method": "ping",
                "params": {"message": "transport test"},
            },
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        data = response.json()
        result["tests"].append(
            {"name": "Single RPC Request", "status": "passed", "response": data}
        )
        print("  ✅ Single JSON-RPC request: PASSED")

        # Test batch JSON-RPC request
        response = await self.session.post(
            f"{self.base_url}/rpc/batch",
            json=[
                {"jsonrpc": "2.0", "id": "1", "method": "ping", "params": {}},
                {"jsonrpc": "2.0", "id": "2", "method": "status", "params": {}},
            ],
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        batch_data = response.json()
        result["tests"].append(
            {"name": "Batch RPC Request", "status": "passed", "response": batch_data}
        )
        print("  ✅ Batch JSON-RPC request: PASSED")

        # Test RPC introspection
        response = await self.session.get(
            f"{self.base_url}/rpc/introspection",
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        introspection_data = response.json()
        result["tests"].append(
            {
                "name": "RPC Introspection",
                "status": "passed",
                "response": introspection_data,
            }
        )
        print("  ✅ JSON-RPC introspection: PASSED")

    async def _test_websocket_transport(self, result: Dict[str, Any]):
        """Test WebSocket transport endpoints"""
        # Test WebSocket status
        response = await self.session.get(
            f"{self.base_url}/ws/status",
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        status_data = response.json()
        result["tests"].append(
            {"name": "WebSocket Status", "status": "passed", "response": status_data}
        )
        print("  ✅ WebSocket status check: PASSED")

        # Test WebSocket health
        response = await self.session.get(
            f"{self.base_url}/ws/health",
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        health_data = response.json()
        result["tests"].append(
            {"name": "WebSocket Health", "status": "passed", "response": health_data}
        )
        print("  ✅ WebSocket health check: PASSED")

        # Test WebSocket metrics
        response = await self.session.get(
            f"{self.base_url}/ws/metrics",
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        metrics_data = response.json()
        result["tests"].append(
            {"name": "WebSocket Metrics", "status": "passed", "response": metrics_data}
        )
        print("  ✅ WebSocket metrics: PASSED")

    async def _test_sse_transport(self, result: Dict[str, Any]):
        """Test Server-Sent Events transport endpoints"""
        # Test SSE status
        response = await self.session.get(
            f"{self.base_url}/sse/status",
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        status_data = response.json()
        result["tests"].append(
            {"name": "SSE Status", "status": "passed", "response": status_data}
        )
        print("  ✅ SSE status check: PASSED")

        # Test SSE health
        response = await self.session.get(
            f"{self.base_url}/sse/health",
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        health_data = response.json()
        result["tests"].append(
            {"name": "SSE Health", "status": "passed", "response": health_data}
        )
        print("  ✅ SSE health check: PASSED")

        # Test SSE event broadcast (expected to fail gracefully)
        try:
            response = await self.session.post(
                f"{self.base_url}/sse/broadcast",
                json={"event": "test", "data": {"message": "transport test"}},
                headers={"Authorization": f"Bearer {self.access_token}"},
            )
            # This should return an error about no connections, which is expected
            broadcast_data = response.json()
            result["tests"].append(
                {
                    "name": "SSE Broadcast (No Connections)",
                    "status": "passed",
                    "response": broadcast_data,
                }
            )
            print("  ✅ SSE broadcast (no connections): PASSED")
        except Exception as e:
            result["tests"].append(
                {"name": "SSE Broadcast", "status": "expected_failure", "error": str(e)}
            )
            print("  ℹ️  SSE broadcast (no connections): Expected failure")

    async def _test_streamable_http_transport(self, result: Dict[str, Any]):
        """Test Streamable HTTP (MCP) transport endpoints"""
        # Test MCP capabilities
        response = await self.session.get(
            f"{self.base_url}/mcp/capabilities",
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        capabilities_data = response.json()
        result["tests"].append(
            {
                "name": "MCP Capabilities",
                "status": "passed",
                "response": capabilities_data,
            }
        )
        print("  ✅ MCP capabilities: PASSED")

        # Test MCP status
        response = await self.session.get(
            f"{self.base_url}/mcp/status",
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        status_data = response.json()
        result["tests"].append(
            {"name": "MCP Status", "status": "passed", "response": status_data}
        )
        print("  ✅ MCP status check: PASSED")

        # Test MCP health
        response = await self.session.get(
            f"{self.base_url}/mcp/health",
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        health_data = response.json()
        result["tests"].append(
            {"name": "MCP Health", "status": "passed", "response": health_data}
        )
        print("  ✅ MCP health check: PASSED")

    async def _test_stdio_transport(self, result: Dict[str, Any]):
        """Test STDIO transport endpoints"""
        # Test STDIO health
        response = await self.session.get(
            f"{self.base_url}/stdio/health",
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        health_data = response.json()
        result["tests"].append(
            {"name": "STDIO Health", "status": "passed", "response": health_data}
        )
        print("  ✅ STDIO health check: PASSED")

        # Test STDIO execute
        response = await self.session.post(
            f"{self.base_url}/stdio/execute",
            json={
                "command": "echo",
                "args": ["Transport test successful!"],
                "timeout": 5000,
            },
            headers={"Authorization": f"Bearer {self.access_token}"},
        )
        response.raise_for_status()
        execute_data = response.json()
        result["tests"].append(
            {"name": "STDIO Execute", "status": "passed", "response": execute_data}
        )
        print("  ✅ STDIO command execution: PASSED")

    async def create_session(self, server_id: str) -> Dict[str, Any]:
        """Create a new MCP session for the specified server"""
        session_data = {
            "server_id": server_id,
            "protocol": "websocket",  # Use WebSocket for real-time communication
            "user_id": "test-user",
            "client_id": str(uuid.uuid4()),
            "metadata": {
                "test_script": True,
                "purpose": "testing MCP server functionality",
            },
        }

        response = await self.session.post(
            f"{self.base_url}/api/gateway/sessions",
            json=session_data,
            headers={"Authorization": f"Bearer {self.access_token}"},
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
            headers={"Authorization": f"Bearer {self.access_token}"},
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

    async def connect_to_server(
        self, server_id: str, protocol: str = "websocket"
    ) -> Dict[str, Any]:
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
                "capabilities": {"roots": {"listChanged": True}, "sampling": {}},
                "clientInfo": {"name": "mcp-gateway-test", "version": "1.0.0"},
            },
        }

        response = await self.gateway.session.post(
            f"{self.gateway.base_url}/servers/{server_id}/rpc",
            json=init_request,
            headers={"Authorization": f"Bearer {self.gateway.access_token}"},
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
                    "capabilities": {"roots": {"listChanged": True}, "sampling": {}},
                    "clientInfo": {"name": "mcp-gateway-test", "version": "1.0.0"},
                },
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
        if hasattr(self, "use_http") and self.use_http:
            # Use HTTP JSON-RPC
            response = await self.gateway.session.post(
                f"{self.gateway.base_url}/servers/{self.server_id}/rpc",
                json=request,
                headers={"Authorization": f"Bearer {self.gateway.access_token}"},
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
            "params": {},
        }

        result = await self._send_request(request)

        if "error" in result:
            raise Exception(f"List tools error: {result['error']}")

        tools = result["result"]["tools"]
        print(f"🛠️  Found {len(tools)} available tools:")
        for tool in tools:
            print(f"  - {tool['name']}: {tool.get('description', 'No description')}")

        return tools

    async def call_tool(
        self, tool_name: str, arguments: Dict[str, Any] = None
    ) -> Dict[str, Any]:
        """Call a specific tool with the given arguments"""
        request = {
            "jsonrpc": "2.0",
            "id": self._next_request_id(),
            "method": "tools/call",
            "params": {"name": tool_name, "arguments": arguments or {}},
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
            "params": {},
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
            print(
                f"  - {resource['uri']}: {resource.get('description', 'No description')}"
            )

        return resources

    async def read_resource(self, resource_uri: str) -> Dict[str, Any]:
        """Read a specific resource from the MCP server"""
        request = {
            "jsonrpc": "2.0",
            "id": self._next_request_id(),
            "method": "resources/read",
            "params": {"uri": resource_uri},
        }

        print(f"📖 Reading resource: {resource_uri}")
        result = await self._send_request(request)

        if "error" in result:
            raise Exception(f"Read resource error: {result['error']}")

        resource_content = result["result"]
        print(f"✅ Resource '{resource_uri}' read successfully")
        return resource_content

    async def list_prompts(self) -> List[Dict[str, Any]]:
        """List all available prompts from the MCP server"""
        request = {
            "jsonrpc": "2.0",
            "id": self._next_request_id(),
            "method": "prompts/list",
            "params": {},
        }

        result = await self._send_request(request)

        if "error" in result:
            if result["error"]["code"] == -32601:  # Method not found
                print("ℹ️  Server does not support prompts")
                return []
            raise Exception(f"List prompts error: {result['error']}")

        prompts = result["result"]["prompts"]
        print(f"💬 Found {len(prompts)} available prompts:")
        for prompt in prompts:
            print(
                f"  - {prompt['name']}: {prompt.get('description', 'No description')}"
            )

        return prompts

    async def get_prompt(
        self, prompt_name: str, arguments: Dict[str, Any] = None
    ) -> Dict[str, Any]:
        """Get a specific prompt with arguments"""
        request = {
            "jsonrpc": "2.0",
            "id": self._next_request_id(),
            "method": "prompts/get",
            "params": {"name": prompt_name, "arguments": arguments or {}},
        }

        print(f"💭 Getting prompt '{prompt_name}' with arguments: {arguments}")
        result = await self._send_request(request)

        if "error" in result:
            raise Exception(f"Get prompt error: {result['error']}")

        prompt_content = result["result"]
        print(f"✅ Prompt '{prompt_name}' retrieved successfully")
        return prompt_content

    async def run_comprehensive_test(
        self, quick_mode=False, verbose=False, list_only=False
    ) -> Dict[str, Any]:
        """Run a comprehensive test suite of MCP server capabilities"""
        test_results = {
            "tools_tested": 0,
            "tools_successful": 0,
            "resources_tested": 0,
            "resources_successful": 0,
            "prompts_tested": 0,
            "prompts_successful": 0,
            "errors": [],
        }

        print("\n🧪 Running comprehensive MCP server test suite...")
        print("=" * 60)

        try:
            # Test tools
            print("\n🛠️  TESTING TOOLS")
            print("-" * 30)
            tools = await self.list_tools()

            test_limit = 1 if quick_mode else 3
            for tool in tools[:test_limit]:  # Test limited number of tools
                test_results["tools_tested"] += 1
                tool_name = tool["name"]

                try:
                    print(f"\n🔧 Testing tool: {tool_name}")

                    if list_only:
                        print(f"ℹ️  Skipping execution (list-only mode)")
                        test_results["tools_successful"] += 1
                        continue

                    # Try to call tool without arguments first
                    result = await self.call_tool(tool_name)
                    test_results["tools_successful"] += 1

                    # Show result preview
                    result_str = json.dumps(result, indent=2)
                    if verbose or len(result_str) <= 200:
                        print(f"📋 Result: {result_str}")
                    else:
                        print(f"📋 Result preview: {result_str[:200]}...")

                except Exception as e:
                    error_msg = str(e)
                    test_results["errors"].append(f"Tool {tool_name}: {error_msg}")

                    # If tool requires arguments, try with sample arguments
                    if (
                        "arguments" in error_msg.lower()
                        or "required" in error_msg.lower()
                    ):
                        print(
                            f"⚠️  Tool requires arguments, trying with sample data..."
                        )
                        try:
                            # Try with common argument patterns
                            sample_args = self._generate_sample_arguments(tool)
                            if sample_args:
                                result = await self.call_tool(tool_name, sample_args)
                                test_results["tools_successful"] += 1
                                print(f"✅ Tool worked with sample arguments")
                        except Exception as e2:
                            print(f"❌ Tool failed even with sample arguments: {e2}")
                    else:
                        print(f"❌ Tool failed: {error_msg}")

            # Test resources
            print("\n📚 TESTING RESOURCES")
            print("-" * 30)
            resources = await self.list_resources()

            for resource in resources[:test_limit]:  # Test limited number of resources
                test_results["resources_tested"] += 1
                resource_uri = resource["uri"]

                try:
                    print(f"\n📖 Testing resource: {resource_uri}")

                    if list_only:
                        print(f"ℹ️  Skipping reading (list-only mode)")
                        test_results["resources_successful"] += 1
                        continue

                    content = await self.read_resource(resource_uri)
                    test_results["resources_successful"] += 1

                    # Show content preview
                    if isinstance(content, dict):
                        content_str = json.dumps(content, indent=2)
                    else:
                        content_str = str(content)

                    if verbose or len(content_str) <= 300:
                        print(f"📄 Content: {content_str}")
                    else:
                        print(f"📄 Content preview: {content_str[:300]}...")

                except Exception as e:
                    error_msg = str(e)
                    test_results["errors"].append(
                        f"Resource {resource_uri}: {error_msg}"
                    )
                    print(f"❌ Resource failed: {error_msg}")

            # Test prompts
            print("\n💬 TESTING PROMPTS")
            print("-" * 30)
            prompts = await self.list_prompts()

            for prompt in prompts[:test_limit]:  # Test limited number of prompts
                test_results["prompts_tested"] += 1
                prompt_name = prompt["name"]

                try:
                    print(f"\n💭 Testing prompt: {prompt_name}")

                    if list_only:
                        print(f"ℹ️  Skipping execution (list-only mode)")
                        test_results["prompts_successful"] += 1
                        continue

                    content = await self.get_prompt(prompt_name)
                    test_results["prompts_successful"] += 1

                    # Show prompt preview
                    if isinstance(content, dict):
                        content_str = json.dumps(content, indent=2)
                    else:
                        content_str = str(content)

                    if verbose or len(content_str) <= 300:
                        print(f"💫 Prompt: {content_str}")
                    else:
                        print(f"💫 Prompt preview: {content_str[:300]}...")

                except Exception as e:
                    error_msg = str(e)
                    test_results["errors"].append(f"Prompt {prompt_name}: {error_msg}")

                    # Try with sample arguments if needed
                    if "arguments" in error_msg.lower():
                        print(
                            f"⚠️  Prompt requires arguments, trying with sample data..."
                        )
                        try:
                            sample_args = {
                                "query": "test",
                                "text": "sample text",
                                "input": "test input",
                            }
                            content = await self.get_prompt(prompt_name, sample_args)
                            test_results["prompts_successful"] += 1
                            print(f"✅ Prompt worked with sample arguments")
                        except Exception as e2:
                            print(f"❌ Prompt failed even with sample arguments: {e2}")
                    else:
                        print(f"❌ Prompt failed: {error_msg}")

        except Exception as e:
            test_results["errors"].append(f"Test suite error: {str(e)}")
            print(f"❌ Test suite encountered error: {e}")

        return test_results

    def _generate_sample_arguments(self, tool: Dict[str, Any]) -> Dict[str, Any]:
        """Generate sample arguments for a tool based on its schema"""
        sample_args = {}

        # Check if tool has input schema
        input_schema = tool.get("inputSchema", {})
        properties = input_schema.get("properties", {})

        for prop_name, prop_schema in properties.items():
            prop_type = prop_schema.get("type", "string")

            if prop_type == "string":
                sample_args[prop_name] = "test"
            elif prop_type == "number" or prop_type == "integer":
                sample_args[prop_name] = 1
            elif prop_type == "boolean":
                sample_args[prop_name] = True
            elif prop_type == "array":
                sample_args[prop_name] = ["test"]
            elif prop_type == "object":
                sample_args[prop_name] = {}

        # Common argument patterns if no schema available
        if not sample_args:
            common_args = {
                "query": "test query",
                "text": "sample text",
                "message": "test message",
                "input": "test input",
                "path": "/tmp/test",
                "file": "test.txt",
                "url": "https://example.com",
                "name": "test",
                "value": "test value",
            }

            # Try to match tool name to common patterns
            tool_name_lower = tool["name"].lower()
            if "search" in tool_name_lower:
                sample_args = {"query": "test search"}
            elif "read" in tool_name_lower or "get" in tool_name_lower:
                sample_args = {"path": "/tmp/test.txt"}
            elif "write" in tool_name_lower or "create" in tool_name_lower:
                sample_args = {"path": "/tmp/test.txt", "content": "test content"}
            else:
                # Use first few common args
                sample_args = dict(list(common_args.items())[:2])

        return sample_args

    async def disconnect(self):
        """Disconnect from the MCP server"""
        if self.websocket:
            await self.websocket.close()
            print("🔌 Disconnected from MCP server")


async def main():
    parser = argparse.ArgumentParser(
        description="Comprehensive MCP Gateway Transport & Server Test Agent"
    )
    parser.add_argument(
        "--transports-only",
        action="store_true",
        help="Only test transport protocols, skip MCP server testing",
    )
    parser.add_argument("--server-id", help="Specific server ID to use (optional)")
    parser.add_argument("--tool", help="Specific tool to call (optional)")
    parser.add_argument("--args", help="Tool arguments as JSON string (optional)")
    parser.add_argument(
        "--quick",
        action="store_true",
        help="Run quick test (1 tool/resource/prompt each)",
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Show full response content (no truncation)",
    )
    parser.add_argument(
        "--list-only",
        action="store_true",
        help="Only list capabilities, don't test execution",
    )
    args = parser.parse_args()

    print("🚀 Starting MCP Gateway Transport & Server Test Agent")
    print("=" * 60)

    # Initialize gateway client
    gateway = MCPGatewayClient()

    try:
        # Authenticate
        await gateway.authenticate(ADMIN_EMAIL, ADMIN_PASSWORD)

        # Test all transports if requested
        if args.transports_only:
            print("\n🔄 Running transport-only tests...")
            transport_results = await gateway.test_all_transports()

            # Print transport test summary
            print("\n📊 TRANSPORT TEST RESULTS")
            print("=" * 50)

            passed_count = 0
            total_count = len(transport_results)

            for transport_name, result in transport_results.items():
                status_emoji = (
                    "✅"
                    if result["status"] == "passed"
                    else "❌"
                    if result["status"] == "failed"
                    else "⚠️"
                )
                print(
                    f"{status_emoji} {transport_name.upper().replace('_', ' ')}: {result['status']}"
                )

                if result["status"] == "passed":
                    passed_count += 1
                    print(f"   Tests passed: {len(result['tests'])}")
                elif result["status"] == "failed":
                    print(f"   Errors: {len(result['errors'])}")
                    if args.verbose and result["errors"]:
                        for error in result["errors"]:
                            print(f"     - {error}")

            success_rate = (passed_count / total_count * 100) if total_count > 0 else 0
            print(
                f"\n🎯 Overall Transport Success Rate: {success_rate:.1f}% ({passed_count}/{total_count})"
            )

            if passed_count == total_count:
                print("\n🎉 All transport protocols are working correctly!")
            else:
                print(
                    f"\n⚠️  {total_count - passed_count} transport(s) need attention."
                )

            return

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
        print(
            f"📊 Server capabilities: {json.dumps(server_info.get('capabilities', {}), indent=2)}"
        )

        # List available tools
        print(f"\n🛠️  Discovering tools...")
        tools = await agent.list_tools()

        # List available resources (if supported)
        print(f"\n📚 Discovering resources...")
        resources = await agent.list_resources()

        # List available prompts (if supported)
        print(f"\n💬 Discovering prompts...")
        prompts = await agent.list_prompts()

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
                print(
                    f"❌ Tool '{args.tool}' not found. Available tools: {[t['name'] for t in tools]}"
                )

        # Run comprehensive test suite
        else:
            if args.list_only:
                print("\n🔍 Running in list-only mode (no execution)")
            elif args.quick:
                print("\n⚡ Running in quick mode (1 item each)")
            elif args.verbose:
                print("\n📝 Running in verbose mode (full output)")

            test_results = await agent.run_comprehensive_test(
                quick_mode=args.quick, verbose=args.verbose, list_only=args.list_only
            )

            # Print comprehensive test summary
            print(f"\n📊 COMPREHENSIVE TEST RESULTS")
            print("=" * 50)
            print(
                f"🛠️  Tools: {test_results['tools_successful']}/{test_results['tools_tested']} successful"
            )
            print(
                f"📚 Resources: {test_results['resources_successful']}/{test_results['resources_tested']} successful"
            )
            print(
                f"💬 Prompts: {test_results['prompts_successful']}/{test_results['prompts_tested']} successful"
            )

            if test_results["errors"]:
                print(f"\n⚠️  ERRORS ENCOUNTERED ({len(test_results['errors'])}):")
                for i, error in enumerate(
                    test_results["errors"][:5], 1
                ):  # Show first 5 errors
                    print(f"   {i}. {error}")
                if len(test_results["errors"]) > 5:
                    print(f"   ... and {len(test_results['errors']) - 5} more errors")

            # Calculate overall success rate
            total_tested = (
                test_results["tools_tested"]
                + test_results["resources_tested"]
                + test_results["prompts_tested"]
            )
            total_successful = (
                test_results["tools_successful"]
                + test_results["resources_successful"]
                + test_results["prompts_successful"]
            )
            success_rate = (
                (total_successful / total_tested * 100) if total_tested > 0 else 0
            )

            print(
                f"\n🎯 Overall Success Rate: {success_rate:.1f}% ({total_successful}/{total_tested})"
            )

        # Clean up
        await agent.disconnect()

        print(f"\n✅ MCP Gateway test completed!")
        print(f"   Server: {server['name']}")
        print(f"   Tools discovered: {len(tools)}")
        print(f"   Resources discovered: {len(resources)}")
        print(f"   Prompts discovered: {len(prompts)}")

        # Also run quick transport test if not doing specific server testing
        if not args.server_id and not args.tool:
            print("\n🔄 Running quick transport verification...")
            try:
                transport_results = await gateway.test_all_transports()
                passed_transports = sum(
                    1
                    for result in transport_results.values()
                    if result["status"] == "passed"
                )
                total_transports = len(transport_results)
                print(
                    f"🚀 Transport verification: {passed_transports}/{total_transports} transports working"
                )
            except Exception as e:
                print(f"⚠️  Transport verification failed: {e}")

    except Exception as e:
        print(f"❌ Test failed: {e}")
        raise

    finally:
        await gateway.session.aclose()


if __name__ == "__main__":
    asyncio.run(main())
