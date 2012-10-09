package main


type CalcFunc func(cpu *CPU) ()

type Val interface {
    Get()
}

type Address interface {
    Val
}



type CPU struct{

    A  uint8
    B  uint8
    C  uint8
    D  uint8
    E  uint8
    F  uint8
    H  uint8
    L  uint8

}


func GenLd(a Address, b Address) (CalcFunc) {
    f:=func(cpu *CPU) (){
      a = b

    }   
    return f

}

func main() {
   f:= GenLd(1,2)
   var c CPU


}