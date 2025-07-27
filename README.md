# go-passage-of-time-mcp-server
Passage of Time MCP Server in Golang.  This is heavily inspired by the "[Passage of Time](https://github.com/jlumbroso/passage-of-time-mcp/blob/main/README.md)" MCP server written in Python by @jlumbroso.  This MCP server enables an LLM to perform basic time operations and understand(?) the notion of the passage of time.  The Docker image can be found on [DockerHub](https://hub.docker.com/repository/docker/kevensen/go-pot-mcp-server/general)

## Installation
### From Source
```bash
go install github.com/kevensen/go-passage-of-time-mcp-server/go-potms@latest
```

### From pre-built package
Download one of the provided packages.

## Execution
### As Binary
By default, `go-potms` will run as an "stdio" MCP server.  To run as an Streamable HTTP server, provide a port with the `-port` potion.
```bash
go-potms -port 8080
```
### Docker Image
```
docker run  kevensen/go-pot-mcp-server:latest
```