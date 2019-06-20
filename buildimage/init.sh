#!/usr/bin/expect -f

set timeout 30 
spawn ./seed
expect {
"The seed has not been initialized" { send "yes\r" ; exp_continue }
"Please input password" { send "1\r" }
}
interact
