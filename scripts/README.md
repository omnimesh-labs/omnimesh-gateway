# MCP Gateway Scripts

This directory contains Python scripts and tools for testing and working with the MCP Gateway.

## Setup

1. Create and activate the virtual environment:
   ```bash
   python3 -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

2. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```

## Scripts

### `transport_smoke_test.py`
Quick smoke test for all MCP Gateway transport protocols (JSON-RPC, WebSocket, SSE, MCP, STDIO).

```bash
python transport_smoke_test.py
```

### `test_mcp_agent.py`
Comprehensive MCP Gateway and server testing tool with multiple modes:

```bash
# Transport-only testing
python test_mcp_agent.py --transports-only

# Full server testing
python test_mcp_agent.py

# Test specific server
python test_mcp_agent.py --server-id <server-id>

# Test specific tool
python test_mcp_agent.py --tool <tool-name> --args '{"param": "value"}'

# Quick mode (1 item each)
python test_mcp_agent.py --quick

# Verbose output
python test_mcp_agent.py --verbose

# List capabilities only
python test_mcp_agent.py --list-only
```

### `test_mcp_server.py`
A simple MCP server for testing purposes. Provides basic tools, resources, and prompts.

```bash
python test_mcp_server.py
```

## Requirements

- Python 3.7+
- httpx>=0.24.0
- websockets>=11.0

All dependencies are listed in `requirements.txt`.
