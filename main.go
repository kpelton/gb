package main

import ("cpu"
        "flag"
        "fmt"
        )
func main() {
    var host = flag.Bool("s", false, "Host server for gameport")
    var serv = flag.String("c", "", "connect to server")
    flag.Parse()

    fmt.Println(*host)
    var c = cpu.NewCpu(*host,*serv)
    c.Exec()

}
