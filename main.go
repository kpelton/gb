package main

import ("cpu"
        "flag"
        "fmt"
        )
func main() {
    var host = flag.Bool("serv", false, "Host server for gameport")
    var serv = flag.String("conenct", "", "connect to server")
    var scale = flag.Int("s", 4, "window scale")

    flag.Parse()

    fmt.Println(*host)
    var c = cpu.NewCpu(*host,*serv,*scale)
    c.Exec()

}
