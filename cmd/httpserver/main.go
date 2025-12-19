package main

import (
	"HTTPFTCP/internal/server"
	"HTTPFTCP/internal/request"
	"HTTPFTCP/internal/response"
	"HTTPFTCP/internal/headers"
	"log"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"net/http"
	"io"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

const port = 42069

func main() {
    server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func proxyHandler(w *response.Writer, req *request.Request) {
	stripped := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + stripped

	resp, err := http.Get(url); if err != nil {
		handler500(w,req)
		return
	}
	defer resp.Body.Close()
	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	delete(h, "Content-Length")
	h.Override("Transfer-Encoding", "chunked")
	//State the trailers that will show up later
	h.Override("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(h)

	buf := make([]byte, 1024)
	var fB []byte //keep track of bytes for SHA256 hash and length
	for {
    	n, err := resp.Body.Read(buf)
    	if n > 0 {
			fB = append(fB, buf[:n]...)
        	if _, werr := w.WriteChunkedBody(buf[:n]); werr != nil {
            	log.Println("error writing chunk:", werr)
				break
        	}
    	}
    	if err == io.EOF {
        	break
    	}
    	if err != nil {
        	break
    	}
	}

	if _, err := w.WriteChunkedBodyDone(); err != nil {
		log.Println("error finishing chunked body:", err)
		return
	}

	//compute sha256 hash and length then write trailers
	hash := sha256.Sum256(fB)
	hexString := hex.EncodeToString(hash[:])
	conLen := strconv.Itoa(len(fB))
	trailers := headers.Headers{}
	trailers["X-Content-SHA256"] = hexString
	trailers["X-Content-Length"] = conLen

	if err := w.WriteTrailers(trailers); err != nil {
		log.Println("error finishing trailer/s", err)
	}
}

func handler(w *response.Writer, req *request.Request) {
	
	const badRequestHTML =     
			`<html>
  				<head>
    			  <title>400 Bad Request</title>
  				</head>
  				<body>
    			  <h1>Bad Request</h1>
    			  <p>Your request honestly kinda sucked.</p>
  				</body>
			</html>`

	const internalServerErrHTML =
	`<html>
  		<head>
    	  <title>500 Internal Server Error</title>
  		</head>
  		<body>
    	  <h1>Internal Server Error</h1>
    	  <p>Okay, you know what? This one is on me.</p>
  		</body>
	</html>`

	const okHTML =
	`<html>
  		<head>
    	  <title>200 OK</title>
  		</head>
  		<body>
    	  <h1>Success!</h1>
    	  <p>Your request was an absolute banger.</p>
  		</body>
	</html>`

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
        proxyHandler(w, req)
        return
    }

	if req.RequestLine.RequestTarget == "/yourproblem" {
		w.WriteStatusLine(response.StatusCodeBadRequest)
		body := []byte(badRequestHTML)
		h := response.GetDefaultHeaders(len(body))
		h.Override("Content-Type", "text/html")
		w.WriteHeaders(h)
		if _, err := w.WriteBody([]byte(badRequestHTML)); err != nil {
    log.Println("write body error (400):", err)
}
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		w.WriteStatusLine(response.StatusCodeInternalServerError)
		body := []byte(internalServerErrHTML)
		h := response.GetDefaultHeaders(len(body))
		h.Override("Content-Type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody([]byte(internalServerErrHTML))
		return
	}
	w.WriteStatusLine(response.StatusCodeSuccess)
	body := []byte(okHTML)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	if _, err := w.WriteBody([]byte(okHTML)); err != nil {
    log.Println("write body error (200):", err)
}
	return
}

func handler500(w *response.Writer, _ *request.Request) {
	const internalServerErrHTML = `<html>
<head>
  <title>500 Internal Server Error</title>
</head>
<body>
  <h1>Internal Server Error</h1>
  <p>Okay, you know what? This one is on me.</p>
</body>
</html>`

	w.WriteStatusLine(response.StatusCodeInternalServerError)
	body := []byte(internalServerErrHTML)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	_, _ = w.WriteBody(body)
}