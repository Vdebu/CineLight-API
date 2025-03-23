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
    <title>Simple CORS</title>
</head>
<body>

    <h1>Simple CORS</h1>
    <div id="output"></div>
    <script>
        document.addEventListener('DOMContentLoaded',function () {
            fetch("http://localhost:3939/v1/healthcheck").then(
                function (response) {
                    response.text().then(function (text) {
                        document.getElementById("output").innerHTML=text;
                    });
                },
                function (err) {
                    document.getElementById("output").innerHTML=err;
                }
            );
        });
    </script>

</body>
</html>`

func main() {
	addr := flag.String("addr", ":9393", "Server address")
	flag.Parse()
	log.Printf("server start on:%s", *addr)
	err := http.ListenAndServe(*addr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(html))
	}))
	log.Fatal(err)
}
