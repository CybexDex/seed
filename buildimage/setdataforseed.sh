#!/usr/bin/expect -f

set timeout 30 
spawn ./seed --data
expect {
"Please input passowrd" { send "1\r" ; exp_continue }
"Please input the name first" { send "ETH\r" ; exp_continue }
"Please input the data" { send "0x0050AbF889f6EeAC47801aC7bDa27e2E1a1eEaB5\r" ; exp_continue }
"Please input the name first" { send "LTC\r" ; exp_continue }
"Please input the data" { send "mmsFKCZWY42RiEg6HFFzYNHRYbyuG7svs5\r" ; exp_continue }
"Please input the name first" { send "NEO\r" ; exp_continue }
"Please input the data" { send "AWj1CNTPAs7g4zxV3FSbS7i1ZNtniTXTxz\r" ; exp_continue }
"Please input the name first" { send "USDT\r" ; exp_continue }
"Please input the data" { send "msQKET8mMYS5p4xbuqgGX9fAHV2rW39k66\r" ; exp_continue }
"Please input the name first" { send "BTC\r" ; exp_continue }
"Please input the data" { send "mxjaiokGCamtqAR9GQyPrYKArgTJTKSQjL\r" ; exp_continue }
"Please input the name first" { send "CYB\r" ; exp_continue }
"Please input the data" { send "test100/P5KFnbs2gEwEy3tzzzJYFBR7SqJtV9iZUpXPsffrjSDXB/andytest1\r" ; exp_continue }
"Please input the name first" { send "EOS\r" ; exp_continue }
"Please input the data" { send "hotwallet/5KBefyZPfqRH6pJiaGFgua7dupA4sQeVomzq9QssyBWX14udekE/coldwallet\r"}
}
interact
