/*
 Input Pullup Serial
 
 This example demonstrates the use of pinMode(INPUT_PULLUP). It reads a 
 digital input on pin 2 and prints the results to the serial monitor.
 
 The circuit: 
 * Momentary switch attached from pin 2 to ground 
 * Built-in LED on pin 13
 
 Unlike pinMode(INPUT), there is no pull-down resistor necessary. An internal 
 20K-ohm resistor is pulled to 5V. This configuration causes the input to 
 read HIGH when the switch is open, and LOW when it is closed. 
 
 created 14 March 2012
 by Scott Fitzgerald
 
 http://www.arduino.cc/en/Tutorial/InputPullupSerial
 
 This example code is in the public domain
 
 */
#include <digitalWriteFast.h>

void setup(){
  //start serial connection
  Serial.begin(115200);
  //configure pin2 as an input and enable the internal pull-up resistor
  pinMode(11, INPUT);
  pinMode(5, INPUT);
  pinMode(2, OUTPUT);

  pinMode(13, OUTPUT); 

}

void loop(){
  //read the pushbutton value into a variable

  int count = 7;
  int val=0;
  int bitval = 0;
  int output = 0;
  int i = 0;
  //print out the value of the pushbutton
  for(;;){
      int CLK = PINB & 0x8;
    
      if (CLK != 8){ 
         for (count=7; count>=0; count--){
           bitval = (output >> count) & 1;
           PORTD  = (bitval & ~(1<<2)) | (bitval<<2); 
           delayMicroseconds(60);       
           val |= ((PIND >> 5 ) & 0x1) <<count;
           delayMicroseconds(45);
         }
         Serial.write(val);
      }
         val = 0;
         i = Serial.read();
         if (i != -1) {
             output = i; 
         }

   
   

}

}

