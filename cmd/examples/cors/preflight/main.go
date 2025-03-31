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
        document.addEventListener("DOMContentLoaded",function () {
            fetch("http://localhost:3939/v1/tokens/authentication",{
                method:"POST",
                headers:{
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    email:"mikudayo@vocaloid.com",
                    password:"mikudayo3939"
                })
            }).then(
                function (response) {
                    response.text().then(function (text) {
                        document.getElementById("output").innerHTML = text;
                    });
                },
                function (err) {
                    document.getElementById("output").innerHTML = err;
                }
            );
        });
    </script>

</body>
</html>
`

// 测试CORS预检请求
func main() {
	addr := flag.String("addr", ":9393", "Server address")
	flag.Parse()
	// 于API启动时设置可信源
	// -cors-trusted-origins=http://localhost:9393
	log.Printf("starting server on %s", *addr)
	err := http.ListenAndServe(*addr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(html))
	}))
	// 输出错误
	log.Fatal(err)
}
