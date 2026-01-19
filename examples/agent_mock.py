import sys
import time
import signal
import threading
from http.server import BaseHTTPRequestHandler, HTTPServer
from socketserver import ThreadingMixIn

# Mock importing the SDK we just wrote
sys.path.append('../../sdk/python')
from aeterna import AeternaClient

# Simulated AI Context (Short-term Memory)
memory_context = {
    "session_id": "uuid-1234",
    "conversation": [],
    "model_version": "v1.0"
}

class ThreadedHTTPServer(ThreadingMixIn, HTTPServer):
    """Handle requests in a separate thread."""
    pass

class AgentHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        global memory_context
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()

        # Simulate processing
        response = f"{{'status': 'alive', 'memory_size': {len(memory_context['conversation'])}, 'history': {memory_context['conversation']}}}"
        self.wfile.write(response.encode())

    def do_POST(self):
        length = int(self.headers.get('content-length', 0))
        data = self.rfile.read(length).decode()

        # Update Memory
        memory_context["conversation"].append(data)

        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"Message added to memory.")

def handle_sigterm(signum, frame):
    print(f"[Agent] Received SIGTERM. Cleaning up...")
    # In a real scenario, we might dump state here if we use a Push model
    sys.exit(0)

def main():
    global memory_context
    client = AeternaClient()

    # 1. Load Context (SRP)
    restored_state = client.load_context()
    if restored_state:
        memory_context = restored_state
        print(f"[Agent] 🧠 MEMORY RESTORED! Resuming conversation...")
    else:
        print(f"[Agent] 👶 Starting fresh (Cold Start).")

    # 2. Get Inherited Socket
    sock = client.get_listener_socket()

    # 3. Start Server
    server = ThreadedHTTPServer(("0.0.0.0", 8080), AgentHandler, bind_and_activate=False)
    server.socket = sock
    server.server_bind()
    server.server_activate()

    print(f"[Agent] Serving on port {sock.getsockname()[1]} (PID: {os.getpid()})")

    signal.signal(signal.SIGTERM, handle_sigterm)

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass

if __name__ == "__main__":
    main()

# Personal.AI order the ending
