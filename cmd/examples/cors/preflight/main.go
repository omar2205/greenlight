package main

import (
	"flag"
	"log"
	"net/http"
)

const html = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
</head>
	<body>
	<h1>Preflight CORS</h1>
	<div id="output"></div>
	<script>
	document.addEventListener('DOMContentLoaded', function () {
		fetch(
			'https://omar2205-code50-1373867-vgj6qr62pqvg-4000.githubpreview.dev/v1/tokens/authentication',
			{
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					email: 'alice@google.com',
					password: 'pa55word',
				}),
			}
		).then(
			function (response) {
				response.text().then(function (text) {
					document.getElementById('output').innerHTML = text
				})
			},
			function (err) {
				document.getElementById('output').innerHTML = err
			}
		)
	})	
	</script>
	</body>
</html>`

func main() {
	addr := flag.String("addr", ":9000", "server address")
	flag.Parse()

	log.Printf("starting server on %s", *addr)
	err := http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	log.Fatal(err)
}
