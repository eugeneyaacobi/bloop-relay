from http.server import BaseHTTPRequestHandler, HTTPServer


class Handler(BaseHTTPRequestHandler):
    def _write_response(self, payload: bytes):
        self.send_response(200)
        self.send_header("Content-Type", "text/plain; charset=utf-8")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def do_GET(self):
        body = f"echo GET {self.path}\n".encode()
        self._write_response(body)

    def do_POST(self):
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length)
        response = b"echo POST " + self.path.encode() + b"\n" + body
        self._write_response(response)

    def log_message(self, format, *args):
        pass


HTTPServer(("127.0.0.1", 19090), Handler).serve_forever()
