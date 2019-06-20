#!/usr/bin/expect -f

set timeout 30 
spawn ./seed --config
expect {
"Please input passowrd" { send "1\r" ; exp_continue }
"Please input the name first" { send "client\r" ; exp_continue }
"Please input the config" { send "172.17.0.1,172.17.0.2,172.17.0.3,172.17.0.4,172.17.0.5,172.17.0.6,172.17.0.7,172.17.0.8,172.17.0.9,172.17.0.10,172.17.0.11,172.17.0.12,172.17.0.13,172.17.0.14,172.17.0.15,172.17.0.16,172.17.0.17,172.17.0.18,172.17.0.19,172.17.0.20\r" }
}
interact
