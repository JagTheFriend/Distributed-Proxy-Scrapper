# Residential Proxy Node Manager (Go)

A high-performance backend system for managing residential proxy nodes using WebSockets and Valkey (Redis-compatible). This service acts as the **central router** that connects API requests to distributed residential nodes and relays responses back in real time.

## Overview

This project implements a **node-based proxy architecture** where:

* Nodes (clients) connect via WebSocket
* Each node acts as an **exit point** for HTTP requests
* The server assigns jobs to nodes
* Nodes execute requests and return results
* The server relays responses back to API consumers

## Architecture

Core Components:

* **Node Manager**

  * Tracks node lifecycle
  * Stores metadata in Valkey
  * Maintains node states: `idle`, `busy`, `blocked`

* **WebSocket Layer**

  * Persistent connection with nodes
  * Handles heartbeat, job dispatch, and result collection

* **Trigger API**

  * Accepts external requests
  * Selects available node
  * Dispatches job
  * Waits for response

* **Valkey (Redis)**

  * Stores node registry
  * Tracks node status
  * Enables fast lookup

## Data Model

### Client (Node)

```json
{
  "clientId": "node_123",
  "clientType": "mobile",
  "ip": "192.168.1.1",
  "geo": {
    "country": "IN",
    "city": "Bhubaneswar"
  },
  "status": "idle",
  "connectedAt": 1710000000,
  "latencyMs": 20,
  "load": 0.
```
