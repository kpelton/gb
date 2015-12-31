package main

import ("cpu"
        "flag"
        "fmt"
        "runtime"

        )
func main() {
    var host = flag.Bool("serv", false, "Host server for gameport")
    var serv = flag.String("connect", "", "connect to server")
    var scale = flag.Int("s", 4, "window scale")
    var serialp = flag.String("serial", "", "Use real link port - arduino usb device")
    runtime.LockOSThread() 
    flag.Parse()

    fmt.Println(*host)
    var c = cpu.NewCpu(*host,*serv,*scale,*serialp)
    c.Exec()

}
