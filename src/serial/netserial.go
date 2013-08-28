
package serial

import (
    "fmt"   
    "ic"
    "constants"
    "net"
    "time"
)

const (
    port = ":08313"
    CMD_WRITE = 0x1
    CMD_ACK = 0x2

    CMD_CLOSE = 0x3
    CMD_DONE= 0x4

)
type NetSerial struct {
    SB uint8
    SC uint8
    ic *ic.IC
    bytes_sent bool
    sock   net.Conn
}
func (s *NetSerial)  listen() {
    fmt.Println("Waiting for connection to gameboy")
    l, err := net.Listen("tcp",port)
    if err != nil {
	    panic("Unable to listen on tcp port!!!")
    }
	conn , err := l.Accept()
	if err != nil {
	    panic(err)
	}
    s.sock = conn
    
}
func (s *NetSerial)  connect(addr string) {
    fmt.Println("Connecting to"+addr)

    conn, err := net.Dial("tcp", addr+port)
    if err != nil {
        panic(err)
    }
    s.sock = conn
    
}

func NewNetSerial(ic *ic.IC,listen bool, addr string) *NetSerial {
	serial := new(NetSerial)
    serial.ic = ic
    if listen ==  true {
        serial.listen()

    }else if addr != "" {
        serial.connect(addr)
    }else{

        panic("Need arguments in order to create netserial!!")
    }
	return serial
}


func (s *NetSerial) sendBytes(cmd uint8,value uint8) {
    var  msg [2]byte
    msg[0] = cmd
    msg[1] = value
    count,err  := s.sock.Write(msg[0:])
    if err != nil || count ==0 {

        panic(err)
    }
        fmt.Println("SERIAL:sent ",count,"bytes")

}

func (s *NetSerial) getBytes() {
    var  msg [2]byte
    s.sock.SetReadDeadline(time.Now().Add(time.Duration(1) *time.Millisecond))

    count,err := s.sock.Read(msg[0:])
    if err != nil {
        //panic("Error getting bytes")
        
    }
    
    if count > 0 {
        fmt.Println("SERIAL:Recieved ",count,"bytes")
        fmt.Println("SERIAL:Recieved ",msg[1],)
        if msg[0] == CMD_WRITE {
            s.sendBytes(CMD_ACK,s.SB)
                    s.ic.Assert(constants.SERIAL)
                    s.SB = msg[1]

        fmt.Println("SERIAL:Recieved WRITE ",msg[1],)
        } else if msg[0] == CMD_ACK { 
            s.bytes_sent = false
            s.ic.Assert(constants.SERIAL)
                    s.SB = msg[1]
        fmt.Println("SERIAL:Recieved ACK ",msg[1],)


        }else {
            panic("invalid condition")
        }

            s.SC &=   (^uint8(0x80))

    }

}
func (s *NetSerial) Update() uint8 {

    if s.SC & 0x80  == 0x80 {
        s.getBytes()
    }
    return 0

}

func (s *NetSerial) Read(addr uint16) uint8 {
	switch addr {
	case SB_ADDR:
    	return s.SB
	case SC_ADDR:
        return s.SC 
    default:
        panic("mis-routed serial write!")
    }
}

func (s *NetSerial) Write(addr uint16,val uint8) {
	switch addr {

	case SB_ADDR:
        fmt.Printf("->SERIALB:%04X\n", val)
		s.SB = val

	case SC_ADDR:
        s.SC = val
        fmt.Printf("->SERIALC:%04X\n", val)

        if val & 0x81 == 0x81 && s.bytes_sent == false{
            //s.SC &=  val &(^uint8(0x80))
            //s.ic.Assert(constants.SERIAL)
            s.sendBytes(CMD_WRITE,s.SB)
            s.bytes_sent = true
        }
        
    default:
        panic("mis-routed serial write!")
    }
}